package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForReady(t *testing.T) {
	tests := []struct {
		name           string
		serverBehavior func(counter *int32) http.HandlerFunc
		timeout        time.Duration
		contextFunc    func() (context.Context, context.CancelFunc)
		expectedError  bool
		errorContains  string
	}{
		{
			name: "server ready immediately",
			serverBehavior: func(_ *int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}
			},
			timeout:       1 * time.Second,
			contextFunc:   func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			expectedError: false,
		},
		{
			name: "server ready after delay",
			serverBehavior: func(counter *int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					count := atomic.AddInt32(counter, 1)
					if count < 4 { // Cambiamos a responder OK después de 3 intentos
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			},
			timeout:       5 * time.Second, // Aumentamos el timeout para dar más tiempo
			contextFunc:   func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			expectedError: false,
		},
		{
			name: "server timeout",
			serverBehavior: func(_ *int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			},
			timeout:       500 * time.Millisecond,
			contextFunc:   func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			expectedError: true,
			errorContains: "timeout after",
		},
		{
			name: "context cancelled",
			serverBehavior: func(_ *int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			},
			timeout: 1 * time.Second,
			contextFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			expectedError: true,
			errorContains: "context cancelled",
		},
		{
			name: "server returns error status",
			serverBehavior: func(_ *int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			timeout:       500 * time.Millisecond,
			contextFunc:   func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			expectedError: true,
			errorContains: "timeout after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Contador atómico para el estado del servidor
			var counter int32

			// Create test server
			server := httptest.NewServer(tt.serverBehavior(&counter))
			defer server.Close()

			// Get context
			ctx, cancel := tt.contextFunc()
			defer cancel()

			// Call WaitForReady
			err := WaitForReady(ctx, tt.timeout, server.URL)

			// Check results
			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q but got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Test for invalid URL
func TestWaitForReadyInvalidURL(t *testing.T) {
	ctx := context.Background()
	err := WaitForReady(ctx, 1*time.Second, "invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL but got nil")
	}
}

// Benchmark test
func BenchmarkWaitForReady(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := WaitForReady(ctx, 1*time.Second, server.URL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCSRFProtection(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		origin     string
		wantStatus int
	}{
		{
			name:       "GET request without Origin header",
			method:     http.MethodGet,
			origin:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST request with valid Origin header",
			method:     http.MethodPost,
			origin:     "http://localhost:8080",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST request with invalid Origin header",
			method:     http.MethodPost,
			origin:     "https://invalid.com",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "POST request without Origin header",
			method:     http.MethodPost,
			origin:     "",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "POST request with another valid Origin header",
			method:     http.MethodPost,
			origin:     "http://localhost:8081",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST request with yet another valid Origin header",
			method:     http.MethodPost,
			origin:     "http://localhost:8082",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("CSRFProtection() = %v, want %v", status, tt.wantStatus)
			}
		})
	}
}
