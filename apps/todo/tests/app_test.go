package tests

import (
	"context"
	"errors"
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

		server.sendRequest(
			http.MethodPost,
			"/users",
			"email="+
				randomEmail()+
				"&password=Str0ngP@ssw0rd!"+
				"&confirm-password=Str0ngP@ssw0rd!",
			false,
		).assertStatus(expectedStatus)
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

		server.sendRequest(
			http.MethodPost,
			"/authenticate/password",
			"email="+randomEmail()+"&password=Str0ngP@ssw0rd!",
			true,
		).assertStatus(expectedStatus)
	}

	checkServerErrors(t, errChan)
}

func TestWeakPasswordRegistration(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	server.sendRequest(
		http.MethodPost,
		"/users",
		"email="+randomEmail()+"&password=test"+"&confirm-password=test",
		false,
	).assertStatus(http.StatusOK).
		assertContains("Password too weak or compromised")

	checkServerErrors(t, errChan)
}

func TestSuccessfulRegistrationAndLoginAndLogout(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	username := randomEmail()
	password := "Str0ngP@ssw0rd!"

	// Register
	server.givenNewUser(username, password)

	// Login
	resp := server.sendRequest(
		http.MethodPost,
		"/authenticate/password",
		"email="+username+"&password="+password,
		false,
	).assertStatus(http.StatusNoContent).
		assertRedirect("/")

	// Test if redirect to verify email
	server.sendRequest(
		http.MethodGet,
		"/",
		"",
		false,
		resp.Cookies()...,
	).assertContains("Verify Email")

	// Verify email with invalid code
	server.sendRequest(
		http.MethodPost,
		"/email-verification-request",
		"code=23",
		true,
		resp.Cookies()...,
	).assertStatus(http.StatusOK).
		assertContains("Invalid email verification code")

	// Verify email with valid code
	server.sendRequest(
		http.MethodPost,
		"/email-verification-request",
		"code="+"12345678",
		true,
		resp.Cookies()...,
	).assertStatus(http.StatusNoContent).
		assertRedirect("/login")

	// Login
	server.sendRequest(
		http.MethodPost,
		"/authenticate/password",
		"email="+username+"&password="+password,
		false,
	).assertStatus(http.StatusNoContent).
		assertRedirect("/")

	// Logout
	server.sendRequest(
		http.MethodGet,
		"/logout",
		"",
		false,
		resp.Cookies()...,
	).assertStatus(http.StatusOK).
		assertRedirect("/login").
		assertSessionCookieDestroyed()

	checkServerErrors(t, errChan)

}

func TestFragment(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	// Setup authenticated user
	user := server.givenNewAuthenticatedUser()

	// Create todo
	server.givenNewTodo(user, "Test Todo", "This is a test todo")

	server.sendRequest(
		http.MethodGet,
		"/todos/1/form",
		"",
		true,
		user.Cookies...,
	).assertStatus(http.StatusOK).
		assertContains(
			"Test Todo",
			"This is a test todo",
			"Save Modification",
		)

	checkServerErrors(t, errChan)
}

func TestTodoCRUD(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	// Setup authenticated user
	user := server.givenNewAuthenticatedUser()

	// Create todo
	server.givenNewTodo(user, "Test Todo", "This is a test todo")

	// Read todo
	server.sendRequest(
		http.MethodGet,
		"/todos/1/form",
		"",
		false,
		user.Cookies...,
	).assertStatus(http.StatusOK).
		assertContains("Test Todo", "This is a test todo")

	// Update todo
	server.sendRequest(
		http.MethodPut,
		"/todos/1",
		"name=Updated+Todo",
		true,
		user.Cookies...,
	).assertStatus(http.StatusOK).
		assertContains("Updated Todo")

	// Delete todo
	server.sendRequest(
		http.MethodDelete,
		"/todos/1",
		"",
		true,
		user.Cookies...,
	).assertStatus(http.StatusOK)

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
