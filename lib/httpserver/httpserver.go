package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func WaitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for server: %w", ctx.Err())
		case <-timeoutTimer.C:
			return fmt.Errorf("timeout after %v waiting for server to be ready", timeout)
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error making request: %v", err)
				continue
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				log.Println("Server is ready")
				return nil
			}

			log.Printf("Server not ready yet. Status: %d", resp.StatusCode)
		}
	}
}
