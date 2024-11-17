package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func WaitForReady(
	ctx context.Context,
	timeout time.Duration,
	endpoint string,
) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)

		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)

		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Server is ready")
			if err := resp.Body.Close(); err != nil {
				return fmt.Errorf("error closing response body: %w", err)
			}
			return nil
		}

		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("error closing response body: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) > timeout {
				return fmt.Errorf("timeout waiting for server to be ready")
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			origin := r.Header.Get("Origin")
			if origin == "" || (origin != "https://example.com" && origin != "http://localhost:8080") {
				http.Error(w, "Forbidden: Invalid origin", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
