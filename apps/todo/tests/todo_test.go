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
			"username="+randomUsername()+"&password=Str0ngP@ssw0rd!",
			false,
		).assertStatus(expectedStatus)
	}

	checkServerErrors(t, errChan)
}

func TestLoginRateLimit(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	for i := 0; i < defaultTestConfig.MaxRetries; i++ {
		expectedStatus := http.StatusUnauthorized
		if i >= defaultTestConfig.RateLimit {
			expectedStatus = http.StatusTooManyRequests
		}

		server.sendRequest(
			http.MethodPost,
			"/authenticate/password",
			"username="+randomUsername()+"&password=Str0ngP@ssw0rd!",
			false,
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
		"username="+randomUsername()+"&password=test",
		false,
	).assertStatus(http.StatusBadRequest).
		assertContains("Password does not meet requirements")

	checkServerErrors(t, errChan)
}

func TestSuccessfulRegistrationAndLoginAndLogout(t *testing.T) {
	server, errChan := setupServer(t, defaultTestConfig)
	defer server.cancel()

	username := randomUsername()
	password := "Str0ngP@ssw0rd!"

	// Register
	server.givenNewUser(username, password)

	// Login
	resp := server.sendRequest(
		http.MethodPost,
		"/authenticate/password",
		"username="+username+"&password="+password,
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
