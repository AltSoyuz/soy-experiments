package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaitForReady(t *testing.T) {
	// Create a test server that will respond with 200 OK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := WaitForReady(ctx, 2*time.Second, ts.URL)
	if err != nil {
		t.Errorf("waitForReady() error = %v", err)
	}
}
