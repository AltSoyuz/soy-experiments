package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRFMiddleware(t *testing.T) {
	csrf, err := NewCSRFProtection("http://localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create CSRF protection: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		origin         string
		token          string
		tokenLocation  string // "header", "form", "cookie"
		expectedStatus int
	}{
		{
			name:           "Valid GET request",
			method:         "GET",
			origin:         "http://localhost:8080",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid POST with token in header",
			method:         "POST",
			origin:         "http://localhost:8080",
			token:          csrf.GenerateToken(),
			tokenLocation:  "header",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid POST with token in form",
			method:         "POST",
			origin:         "http://localhost:8080",
			token:          csrf.GenerateToken(),
			tokenLocation:  "form",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid POST with token in cookie",
			method:         "POST",
			origin:         "http://localhost:8080",
			token:          csrf.GenerateToken(),
			tokenLocation:  "cookie",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid origin",
			method:         "POST",
			origin:         "http://malicious.com",
			token:          csrf.GenerateToken(),
			tokenLocation:  "header",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing token",
			method:         "POST",
			origin:         "http://localhost:8080",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid token",
			method:         "POST",
			origin:         "http://localhost:8080",
			token:          "invalid-token",
			tokenLocation:  "header",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}

			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			if tt.token != "" {
				switch tt.tokenLocation {
				case "header":
					req.Header.Set("X-CSRF-Token", tt.token)
				case "form":
					form := url.Values{}
					form.Add("csrf_token", tt.token)
					req = httptest.NewRequest(tt.method, "/", strings.NewReader(form.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					req.Header.Set("Origin", tt.origin)
				case "cookie":
					req.AddCookie(&http.Cookie{Name: "csrf_token", Value: tt.token})
				}
			}

			rr := httptest.NewRecorder()
			csrf.Middleware(http.HandlerFunc(handler)).ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
