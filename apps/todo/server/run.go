package server

import (
	"context"
	"flag"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/store"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/lib/buildinfo"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"
)

var (
	configPathFlag = flag.String("config", "./apps/todo/config/config.yml", "Path to the configuration file")
)

func Run(ctx context.Context) error {
	flag.Parse()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	buildinfo.Init()

	slog.Info("Configuration path", "path", *configPathFlag)

	// Build the absolute path to the configuration file
	configPath, err := filepath.Abs(*configPathFlag)
	if err != nil {
		return err
	}

	// Initialize the app configuration
	cfg, err := config.Init(configPath)
	if err != nil {
		return err
	}

	// Initialize the web template system
	err = web.TemplateSystem.PrecompileTemplates()
	if err != nil {
		return err
	}

	// Initialize the store with database queries
	queries, err := store.Init(cfg)
	if err != nil {
		return err
	}

	authService := auth.Init(cfg, queries)
	todoStore := todo.Init(queries)

	srv := New(
		cfg,
		authService,
		todoStore,
	)

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
