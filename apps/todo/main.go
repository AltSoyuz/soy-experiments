package main

import (
	"context"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/queries"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/buildinfo"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	buildinfo.Init()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpl, err := views.Init()
	if err != nil {
		return fmt.Errorf("error initializing templates: %w", err)
	}

	q, err := queries.Init()
	if err != nil {
		return fmt.Errorf("error initializing store: %w", err)
	}

	todoService := newTodoService(q)

	srv := newServer(tmpl, todoService)

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
