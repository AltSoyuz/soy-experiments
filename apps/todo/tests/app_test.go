package tests

import (
	"context"
	"errors"
	"golang-template-htmx-alpine/apps/todo/auth"
	"net/http"
	"testing"
)

func TestRegistrationRateLimit(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	for i := 0; i < defaultTestConfig.MaxRetries; i++ {
		expectedStatus := http.StatusNoContent
		if i >= defaultTestConfig.RateLimit {
			expectedStatus = http.StatusTooManyRequests
		}
		resp := server.sendRequest(http.MethodGet, "/register", RequestOptions{
			HTMX: true,
		}).assertStatus(http.StatusOK)

		csrfToken := extractCSRFToken(resp.body)

		server.sendRequest(http.MethodPost, "/users", RequestOptions{
			Body: "email=" + randomEmail() +
				"&password=" + "Str0ngP@ssw0rd!" +
				"&confirm-password=" + "Str0ngP@ssw0rd!" +
				"&csrf_token=" + csrfToken,
			HTMX:      false,
			CSRFToken: csrfToken,
		}).assertStatus(expectedStatus)
	}

	checkServerErrors(t, errChan)
}

func TestLoginRateLimit(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	for i := 0; i < defaultTestConfig.MaxRetries; i++ {
		expectedStatus := http.StatusOK
		if i >= defaultTestConfig.RateLimit {
			expectedStatus = http.StatusTooManyRequests
		}

		resp := server.sendRequest(http.MethodGet, "/login", RequestOptions{
			HTMX: true,
		}).assertStatus(http.StatusOK)

		csrfToken := extractCSRFToken(resp.body)

		server.sendRequest(http.MethodPost, "/authenticate/password", RequestOptions{
			Body: "email=" + randomEmail() +
				"&password=Str0ngP@ssw0rd!" +
				"&csrf_token=" + csrfToken,
			HTMX:      true,
			CSRFToken: csrfToken,
		}).assertStatus(expectedStatus)
	}

	checkServerErrors(t, errChan)
}

func TestWeakPasswordRegistration(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	resp := server.sendRequest(http.MethodGet, "/register", RequestOptions{}).assertStatus(http.StatusOK)

	csrfToken := extractCSRFToken(resp.body)
	server.sendRequest(http.MethodPost, "/users", RequestOptions{
		Body:      "email=" + randomEmail() + "&password=test&confirm-password=test",
		HTMX:      false,
		CSRFToken: csrfToken,
	}).assertStatus(http.StatusOK).
		assertContains("Password too weak or compromised")

	checkServerErrors(t, errChan)
}

func TestSuccessfulRegistrationAndLoginAndLogout(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	email := randomEmail()
	password := "Str0ngP@ssw0rd!"

	// Register
	server.givenNewUser(email, password)

	// Login
	resp := server.sendRequest(http.MethodGet, "/login", RequestOptions{}).assertStatus(http.StatusOK)

	csrfToken := extractCSRFToken(resp.body)
	resp = server.sendRequest(http.MethodPost, "/authenticate/password", RequestOptions{
		Body:      "email=" + email + "&password=" + password,
		HTMX:      false,
		CSRFToken: csrfToken,
	}).assertStatus(http.StatusNoContent).
		assertRedirect("/")

	// Test if redirect to verify email
	resp2 := server.sendRequest(http.MethodGet, "/verify-email", RequestOptions{}).assertStatus(http.StatusOK)

	// Verify email with invalid code
	csrfToken2 := extractCSRFToken(resp2.body)
	resp3 := server.sendRequest(http.MethodPost, "/email-verification-request", RequestOptions{
		Body:      "code=23",
		HTMX:      true,
		Cookies:   resp.Cookies(),
		CSRFToken: csrfToken2,
	}).assertStatus(http.StatusOK).
		assertContains("Invalid email verification code")

	// Verify email with valid code
	csrfToken3 := extractCSRFToken(resp3.body)
	server.sendRequest(http.MethodPost, "/email-verification-request", RequestOptions{
		Body:      "code=" + auth.TestEmailVerificationCode,
		HTMX:      true,
		Cookies:   resp.Cookies(),
		CSRFToken: csrfToken3,
	}).assertStatus(http.StatusNoContent).
		assertRedirect("/login")

	// Login
	resp = server.sendRequest(http.MethodGet, "/login", RequestOptions{
		HTMX: true,
	}).assertStatus(http.StatusOK)

	csrfToken = extractCSRFToken(resp.body)
	resp = server.sendRequest(http.MethodPost, "/authenticate/password", RequestOptions{
		Body:      "email=" + email + "&password=" + password,
		HTMX:      false,
		CSRFToken: csrfToken,
	}).assertStatus(http.StatusNoContent).
		assertRedirect("/")

	// Logout
	server.sendRequest(http.MethodGet, "/logout", RequestOptions{
		HTMX:    false,
		Cookies: resp.Cookies(),
	}).assertStatus(http.StatusOK).
		assertRedirect("/login").
		assertSessionCookieDestroyed()

	checkServerErrors(t, errChan)
}

func TestTodoCRUD(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	// Setup authenticated user
	user := server.givenNewAuthenticatedUser()

	// Create todo
	server.givenNewTodo(user, "Test Todo", "This is a test todo")

	//  Read todo
	resp := server.sendRequest(http.MethodGet, "/todos/1/form", RequestOptions{
		HTMX:    false,
		Cookies: user.Cookies,
	}).assertStatus(http.StatusOK).
		assertContains("Test Todo", "This is a test todo")

	// Update todo
	resp2 := server.sendRequest(http.MethodPut, "/todos/1", RequestOptions{
		Body:      "name=Updated+Todo",
		HTMX:      true,
		Cookies:   user.Cookies,
		CSRFToken: extractCSRFToken(resp.body),
	}).assertStatus(http.StatusOK).
		assertContains("Updated Todo")

	// Complete todo
	csrfToken := extractCSRFToken(resp2.body)
	resp = server.sendRequest(http.MethodPut, "/todos/1/complete", RequestOptions{
		HTMX:      true,
		Cookies:   user.Cookies,
		CSRFToken: csrfToken,
	}).assertStatus(http.StatusOK).
		assertContains("line-through")

	// Delete todo
	server.sendRequest(http.MethodDelete, "/todos/1", RequestOptions{
		HTMX:      true,
		Cookies:   user.Cookies,
		CSRFToken: extractCSRFToken(resp.body),
	}).assertStatus(http.StatusOK)

	checkServerErrors(t, errChan)
}

func checkServerErrors(t *testing.T, errChan chan error) {
	t.Helper()
	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("server error: %v", err)
		}
	default:
	}
}
