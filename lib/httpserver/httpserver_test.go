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
