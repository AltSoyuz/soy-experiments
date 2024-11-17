package ratelimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimited is a rate limiter that limits the rate of requests to a given
type RateLimited struct {
	limiter  *rate.Limiter
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
}

// New creates a new rate limiter
func New() *RateLimited {
	return &RateLimited{
		limiter:  rate.NewLimiter(rate.Every(time.Second), 5),
		visitors: make(map[string]*rate.Limiter),
	}
}

// getVisitorLimiter returns a rate limiter for an IP address
func (h *RateLimited) GetVisitorLimiter(ip string) *rate.Limiter {
	h.mu.Lock()
	defer h.mu.Unlock()

	limiter, exists := h.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Second), 5)
		h.visitors[ip] = limiter
	}
	return limiter
}

// cleanup periodically removes old limiters
func (h *RateLimited) cleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.mu.Lock()
			h.visitors = make(map[string]*rate.Limiter)
			h.mu.Unlock()
		}
	}
}
