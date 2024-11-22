package tests

import (
	"context"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/server"
	"golang-template-htmx-alpine/lib/httpserver"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestConfig holds test configuration parameters
type TestConfig struct {
	BaseURL    string
	RateLimit  int
	MaxRetries int
	Timeout    time.Duration
}

var defaultTestConfig = TestConfig{
	BaseURL:    "http://localhost:8080",
	RateLimit:  5,
	MaxRetries: 10,
	Timeout:    time.Second,
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
	s.sendRequest(
		http.MethodPost,
		"/users",
		"email="+email+"&password="+password,
		false,
	).assertStatus(http.StatusNoContent).
		assertRedirect("/login")
}

func (s *testServer) givenNewAuthenticatedUser() *AuthenticatedUser {
	s.t.Helper()
	email := randomEmail()
	password := "Str0ngP@ssw0rd!"

	s.givenNewUser(email, password)

	resp := s.sendRequest(
		http.MethodPost,
		"/authenticate/password",
		"email="+email+"&password="+password,
		false,
	).assertStatus(http.StatusNoContent).
		assertRedirect("/")

	s.sendRequest(
		http.MethodPost,
		"/email-verification-request",
		"code="+auth.TestEmailVerificationCode,
		true,
		resp.Cookies()...,
	).assertStatus(http.StatusNoContent).
		assertRedirect("/login")

	return &AuthenticatedUser{
		Email:   email,
		Cookies: resp.Cookies(),
	}
}

func (s *testServer) givenNewTodo(user *AuthenticatedUser, name, description string) {
	s.t.Helper()
	s.sendRequest(
		http.MethodPost,
		"/todos",
		"name="+name+"&description="+description,
		true,
		user.Cookies...,
	).assertStatus(http.StatusOK).
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
		baseURL: config.BaseURL,
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

	if err := httpserver.WaitForReady(ctx, config.Timeout, config.BaseURL+"/healthz"); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	return ts, errChan
}

func (s *testServer) sendRequest(method, path, body string, htmx bool, cookies ...*http.Cookie) *TestResponse {
	s.t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(s.ctx, method, s.baseURL+path, bodyReader)
	if err != nil {
		s.t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Origin", s.baseURL)
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.t.Fatalf("failed to send %s request: %v", req.Method, err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.t.Fatalf("failed to read response body: %v", err)
	}
	resp.Body.Close()

	return &TestResponse{
		Response: resp,
		t:        s.t,
		body:     string(bodyBytes),
	}
}

func randomEmail() string {
	return "test" + time.Now().Format("20060102150405") + "@example.com"
}
