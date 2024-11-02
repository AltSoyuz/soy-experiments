package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		if err := run(ctx); err != nil {
			t.Errorf("run() error = %v", err)
		}
	}()

	time.Sleep(1 * time.Second) // Give the server time to start

	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}
}

func TestAddRoutes(t *testing.T) {
	mux := http.NewServeMux()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	tmpl := template.Must(template.New("test").Parse("test"))
	render := newRender(logger, tmpl)

	addRoutes(mux, logger, render)

	tests := []struct {
		method string
		url    string
	}{
		{"GET", "/"},
		{"POST", "/todos"},
		{"GET", "/todos/test/form"},
		{"PUT", "/todos/test"},
		{"DELETE", "/todos/test"},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.url, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code == http.StatusNotFound {
			t.Errorf("Route %s %s not found", test.method, test.url)
		}
	}
}

func TestNewServer(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	tmpl := template.Must(template.New("test").Parse("test"))

	handler := newServer(logger, tmpl)

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK; got %v", rr.Code)
	}
}
func TestWaitForReady(t *testing.T) {
	// Create a test server that will respond with 200 OK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := waitForReady(ctx, 5*time.Second, ts.URL)
	if err != nil {
		t.Errorf("waitForReady() error = %v", err)
	}
}
