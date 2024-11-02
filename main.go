package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Todo struct {
	Name        string
	Description string
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func addRoutes(mux *http.ServeMux, logger *slog.Logger, render renderFunc) {
	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /", todoListHandler(render))
	mux.HandleFunc("POST /todos", createTodoHandler(logger, render))
	mux.HandleFunc("GET /todos/{name}/form", getTodoFormHandler(logger, render))
	mux.HandleFunc("PUT /todos/{name}", updateTodoHandler(logger, render))
	mux.HandleFunc("DELETE /todos/{name}", deleteTodoHandler(logger))
}

func newServer(logger *slog.Logger, tmpl *template.Template) http.Handler {
	mux := http.NewServeMux()

	render := newRender(logger, tmpl)

	addRoutes(mux, logger, render)
	var handler http.Handler = mux

	return handler
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	tmpl, err := InitTodoTemplates()
	if err != nil {
		return fmt.Errorf("error initializing templates: %w", err)
	}

	srv := newServer(logger, tmpl)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", "8080"),
		Handler: srv,
	}

	// Start server
	go func() {
		logger.Info("starting http server on 8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error starting http server: %s\n", err)
		}
	}()

	// Handle shutdown in separate goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

// waitForReady calls the specified endpoint until it gets a 200
// response or until the context is cancelled or the timeout is
// reached.
func waitForReady(
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
			fmt.Printf("Error making request: %s\n", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is ready!")
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			// wait a little while between checks
			time.Sleep(250 * time.Millisecond)
		}
	}
}
