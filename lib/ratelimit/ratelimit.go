package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiting middleware
type RateLimiter struct {
	// stores the last request time for each IP
	visitors map[string][]time.Time
	mu       sync.Mutex

	// configuration
	requests int           // number of requests
	window   time.Duration // time window
}

// With creates a new rate limiter
// requests: number of requests allowed
// window: time window for the requests (e.g., 1 minute)
func With(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string][]time.Time),
		requests: requests,
		window:   window,
	}
}

// cleanup removes old entries from the visitors map
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, times := range rl.visitors {
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) <= rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.visitors, ip)
		} else {
			rl.visitors[ip] = valid
		}
	}
}

// LimitMiddleware provides HTTP middleware for rate limiting
func LimitMiddleware(rl *RateLimiter) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get IP address from request
			ip := r.RemoteAddr

			// Periodically cleanup old entries
			go rl.cleanup()

			rl.mu.Lock()

			// Initialize if IP is not in map
			if _, exists := rl.visitors[ip]; !exists {
				rl.visitors[ip] = []time.Time{}
			}

			// Get current time window
			now := time.Now()
			windowStart := now.Add(-rl.window)

			// Filter requests within current window
			var requests []time.Time
			for _, t := range rl.visitors[ip] {
				if t.After(windowStart) {
					requests = append(requests, t)
				}
			}

			// Check if rate limit is exceeded
			if len(requests) >= rl.requests {
				rl.mu.Unlock()
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.requests))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("rate limit exceeded"))
				return
			}

			// Add current request to the list
			rl.visitors[ip] = append(requests, now)
			remaining := rl.requests - len(rl.visitors[ip])

			rl.mu.Unlock()

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
