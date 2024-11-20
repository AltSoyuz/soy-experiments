package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func WaitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
	// Crear el cliente HTTP una sola vez fuera del loop
	client := &http.Client{
		Timeout: 5 * time.Second, // Añadir timeout al cliente
	}

	// Crear un timer para el timeout general
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	// Crear ticker para los reintentos
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
				// Loggear el error pero continuar intentando
				log.Printf("Error making request: %v", err)
				continue
			}

			// Asegurar que siempre cerramos el body
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				log.Println("Server is ready")
				return nil
			}

			// Loggear códigos de estado no exitosos
			log.Printf("Server not ready yet. Status: %d", resp.StatusCode)
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
