package tests

import (
	"context"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/server"
	"golang-template-htmx-alpine/lib/httpserver"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestConfig holds test configuration parameters
type TestConfig struct {
	Port       int
	BaseURL    string
	RateLimit  int
	MaxRetries int
	Timeout    time.Duration
}

var defaultTestConfig = TestConfig{
	BaseURL:    "http://localhost:",
	Port:       8081,
	RateLimit:  5,
	MaxRetries: 10,
	Timeout:    2 * time.Minute,
}

type testServer struct {
	baseURL string
	client  *http.Client
	t       *testing.T
	ctx     context.Context
	cancel  func()
}

type TestResponse struct {
	*http.Response
	t    *testing.T
	body string
}

// Helper functions for common test scenarios
type AuthenticatedUser struct {
	Email   string
	Cookies []*http.Cookie
}

func (s *testServer) givenNewUser(email, password string) {
	s.t.Helper()
	resp := s.sendRequest(http.MethodGet, "/register",
		RequestOptions{
			HTMX: true,
		},
	).assertStatus(http.StatusOK)

	s.sendRequest(http.MethodPost, "/users",
		RequestOptions{
			Body: "email=" + email +
				"&password=" + password +
				"&confirm-password=" + password +
				"&csrf_token=" + extractCSRFToken(resp.body),
			HTMX: false,
		},
	).assertStatus(http.StatusNoContent).
		assertRedirect("/login")
}

func (s *testServer) givenNewAuthenticatedUser() AuthenticatedUser {
	s.t.Helper()
	email := randomEmail()
	password := "Str0ngP@ssw0rd!"

	s.givenNewUser(email, password)
	resp := s.sendRequest(http.MethodGet, "/login", RequestOptions{}).assertStatus(http.StatusOK)
	csrfToken := extractCSRFToken(resp.body)
	resp = s.sendRequest(http.MethodPost, "/authenticate/password", RequestOptions{
		Body:      "email=" + email + "&password=" + password,
		HTMX:      false,
		CSRFToken: csrfToken,
	}).assertStatus(http.StatusNoContent).
		assertRedirect("/")

	// Test if redirect to verify email
	resp2 := s.sendRequest(http.MethodGet, "/verify-email", RequestOptions{}).assertStatus(http.StatusOK)

	// Verify email with valid code
	csrfToken3 := extractCSRFToken(resp2.body)
	s.sendRequest(http.MethodPost, "/email-verification-request", RequestOptions{
		Body:      "code=" + auth.TestEmailVerificationCode,
		HTMX:      true,
		Cookies:   resp.Cookies(),
		CSRFToken: csrfToken3,
	}).assertStatus(http.StatusNoContent).
		assertRedirect("/login")

	return AuthenticatedUser{
		Email:   email,
		Cookies: resp.Cookies(),
	}
}

func (s *testServer) givenNewTodo(user AuthenticatedUser, name, description string) {
	s.t.Helper()

	resp := s.sendRequest(http.MethodGet, "/", RequestOptions{
		Cookies: user.Cookies,
	}).assertStatus(http.StatusOK)

	s.sendRequest(http.MethodPost, "/todos", RequestOptions{
		Body:      "name=" + name + "&description=" + description,
		HTMX:      true,
		Cookies:   user.Cookies,
		CSRFToken: extractCSRFToken(resp.body),
	}).assertStatus(http.StatusOK).
		assertContains(name)
}

// TestResponse methods
func (tr *TestResponse) assertStatus(expectedStatus int) *TestResponse {
	tr.t.Helper()
	if tr.Response.StatusCode != expectedStatus {
		tr.t.Errorf("expected status %d; got %d (%s)", expectedStatus, tr.Response.StatusCode, tr.Response.Status)
	}
	return tr
}

func (tr *TestResponse) assertContains(expectedStrings ...string) *TestResponse {
	tr.t.Helper()
	for _, str := range expectedStrings {
		if !strings.Contains(tr.body, str) {
			tr.t.Errorf("response body missing expected string: %q", str)
		}
	}
	return tr
}

func (tr *TestResponse) assertSessionCookieDestroyed() *TestResponse {
	tr.t.Helper()
	for _, cookie := range tr.Response.Cookies() {
		if cookie.Name == "session" && cookie.MaxAge != -1 {
			tr.t.Errorf("expected session cookie to be destroyed, but it was not")
		}
	}
	return tr
}

func (tr *TestResponse) assertRedirect(expectedPath string) *TestResponse {
	tr.t.Helper()
	if redirect := tr.Header.Get("HX-Redirect"); redirect != expectedPath {
		tr.t.Errorf("expected redirect to %q; got %q", expectedPath, redirect)
	}
	return tr
}

func setupServer(t *testing.T, config TestConfig) (*testServer, chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	envVars := map[string]string{
		"PORT":         fmt.Sprintf("%d", config.Port),
		"SMTP_HOST":    "smtp.mailtrap.io",
		"SMTP_PORT":    "2525",
		"SENDER_EMAIL": "email",
		"SENDER_PASS":  "password",
		"ENV":          "test",
	}

	// add test environment variables
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	ts := &testServer{
		baseURL: fmt.Sprintf("%s%d", config.BaseURL, config.Port),
		client:  &http.Client{Timeout: config.Timeout},
		t:       t,
		ctx:     ctx,
		cancel:  cancel,
	}

	t.Cleanup(func() {
		cancel()
	})

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Run(ctx)
	}()

	if err := httpserver.WaitForReady(ctx, config.Timeout, config.BaseURL+fmt.Sprintf("%d", config.Port)+"/healthz"); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	return ts, errChan
}

type RequestOptions struct {
	Body      string
	HTMX      bool
	Cookies   []*http.Cookie
	CSRFToken string
}

func (s *testServer) sendRequest(method, path string, opts RequestOptions) *TestResponse {
	s.t.Helper()

	var bodyReader io.Reader
	if opts.Body != "" {
		bodyReader = strings.NewReader(opts.Body)
	}

	req, err := http.NewRequestWithContext(s.ctx, method, s.baseURL+path, bodyReader)
	if err != nil {
		s.t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Origin", s.baseURL)
	req.Header.Set("X-CSRF-Token", opts.CSRFToken)

	if opts.Body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if opts.HTMX {
		req.Header.Set("HX-Request", "true")
	}

	for _, cookie := range opts.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.t.Fatalf("failed to send %s request: %v", method, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.t.Fatalf("failed to read response body: %v", err)
	}

	return &TestResponse{
		Response: resp,
		t:        s.t,
		body:     string(bodyBytes),
	}
}

func randomEmail() string {
	// Get current timestamp
	timestamp := time.Now().UnixNano()

	// Generate random string
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[timestamp%int64(len(letters))]
		timestamp = timestamp / int64(len(letters))
	}

	// Create email with timestamp and random string
	return fmt.Sprintf("test_%d_%s@example.com", time.Now().Unix(), string(b))
}

func extractCSRFToken(body string) string {
	// Look for CSRF token in meta tag
	metaPattern := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)"`)
	if matches := metaPattern.FindStringSubmatch(body); len(matches) > 1 {
		return matches[1]
	}

	// Look for CSRF token in input field
	inputPattern := regexp.MustCompile(`<input[^>]+name="csrf_token"[^>]+value="([^"]+)"`)
	if matches := inputPattern.FindStringSubmatch(body); len(matches) > 1 {
		return matches[1]
	}
	// Look for CSRF token in hx-headers attribute
	hxHeadersPattern := regexp.MustCompile(`hx-headers='{"X-CSRF-Token": "([^"]+)"}`)
	if matches := hxHeadersPattern.FindStringSubmatch(body); len(matches) > 1 {
		return matches[1]
	}

	return ""
}
