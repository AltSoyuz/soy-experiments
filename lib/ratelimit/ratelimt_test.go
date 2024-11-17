package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	t.Run("basic configuration", func(t *testing.T) {
		f := func(requests int, window time.Duration) {
			t.Helper()
			rl := With(requests, window)
			if rl.requests != requests {
				t.Fatalf("expected requests to be %d, got: %d", requests, rl.requests)
			}
		}
		f(100, time.Minute)
	})

	t.Run("rate limiting", func(t *testing.T) {
		rl := With(2, time.Minute)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := LimitMiddleware(rl)(handler)

		// First request should succeed
		req1 := httptest.NewRequest("GET", "/", nil)
		rec1 := httptest.NewRecorder()
		middleware.ServeHTTP(rec1, req1)
		if rec1.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec1.Code)
		}

		// Second request should succeed
		req2 := httptest.NewRequest("GET", "/", nil)
		rec2 := httptest.NewRecorder()
		middleware.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec2.Code)
		}

		// Third request should be rate limited
		req3 := httptest.NewRequest("GET", "/", nil)
		rec3 := httptest.NewRecorder()
		middleware.ServeHTTP(rec3, req3)
		if rec3.Code != http.StatusTooManyRequests {
			t.Errorf("expected status 429, got %d", rec3.Code)
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		rl := With(1, 10*time.Millisecond)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		middleware := LimitMiddleware(rl)(handler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		// Wait for window to expire
		time.Sleep(20 * time.Millisecond)

		// Should succeed after cleanup
		rec2 := httptest.NewRecorder()
		middleware.ServeHTTP(rec2, req)
		if rec2.Code != http.StatusOK {
			t.Errorf("expected status 200 after cleanup, got %d", rec2.Code)
		}
	})
}
