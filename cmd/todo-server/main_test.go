package main

import (
	"context"
	"errors"
	"golang-template-htmx-alpine/pkg/httpserver"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	// Setup
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	baseURL := "http://localhost:8080"
	errChan := make(chan error, 1)
	go func() {
		errChan <- run(ctx)
	}()

	// Wait for server
	if err := httpserver.WaitForReady(ctx, 1*time.Second, baseURL+"/healthz"); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	// Helper functions
	makeRequest := func(t *testing.T, method, path string, body string, htmx bool) *http.Response {
		t.Helper()
		var bodyReader io.Reader
		if body != "" {
			bodyReader = strings.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		if body != "" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if htmx {
			req.Header.Set("HX-Request", "true")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to send %s request: %v", method, err)
		}

		return resp
	}

	assertResponse := func(t *testing.T, resp *http.Response, expectedStatus int, expectedStrings ...string) {
		t.Helper()
		defer resp.Body.Close()

		if resp.StatusCode != expectedStatus {
			t.Errorf("expected status %d; got %d (%s)", expectedStatus, resp.StatusCode, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		bodyStr := string(body)
		for _, str := range expectedStrings {
			if !strings.Contains(bodyStr, str) {
				t.Errorf("response body missing expected string: %q", str)
			}
		}
	}

	t.Run("Home page", func(t *testing.T) {
		resp := makeRequest(t, http.MethodGet, "/", "", false)
		assertResponse(t, resp, http.StatusOK,
			"Todo List",
			`<script src="https://unpkg.com/htmx.org`,
		)
	})

	t.Run("CRUD operations", func(t *testing.T) {
		// Create
		resp := makeRequest(t, http.MethodPost, "/todos",
			"name=Test+Todo&description=This+is+a+test+todo", true)
		assertResponse(t, resp, http.StatusOK, "Test Todo")

		// Update
		resp = makeRequest(t, http.MethodPut, "/todos/Test+Todo",
			"name=Updated+Todo&description=This+is+an+updated+test+todo", true)
		assertResponse(t, resp, http.StatusOK,
			"Updated Todo",
			"This is an updated test todo",
		)

		// Delete
		resp = makeRequest(t, http.MethodDelete, "/todos/Updated+Todo", "", true)
		assertResponse(t, resp, http.StatusOK)
	})

	// Check for server errors
	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("server error: %v", err)
		}
	default:
	}
}
func TestRun(t *testing.T) {
	t.Run("Server starts and responds to requests", func(t *testing.T) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		baseURL := "http://localhost:8080"
		errChan := make(chan error, 1)
		go func() {
			errChan <- run(ctx)
		}()

		// Wait for server
		if err := httpserver.WaitForReady(ctx, 1*time.Second, baseURL+"/healthz"); err != nil {
			t.Fatalf("server not ready: %v", err)
		}

		// Make a request to the home page
		resp, err := http.Get(baseURL + "/")
		if err != nil {
			t.Fatalf("failed to make GET request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d; got %d (%s)", http.StatusOK, resp.StatusCode, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if !strings.Contains(string(body), "Todo List") {
			t.Errorf("response body missing expected string: %q", "Todo List")
		}

		// Check for server errors
		select {
		case err := <-errChan:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("server error: %v", err)
			}
		default:
		}
	})
}
