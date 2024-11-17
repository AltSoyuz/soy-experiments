package todo_test

import (
	"context"
	"golang-template-htmx-alpine/apps/todo/queries"
	"golang-template-htmx-alpine/apps/todo/todo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateFromForm(t *testing.T) {
	fakeQuerier := queries.NewFakeQuerier()
	ts := todo.NewTodoService(fakeQuerier)

	ctx := context.Background()

	// Create a new HTTP request with form data
	req := httptest.NewRequest(http.MethodPost, "/todos", nil)
	req.Form = map[string][]string{
		"name":        {"Test Todo"},
		"description": {"This is a test todo"},
	}

	// Call FromRequest to create a Todo model from the request
	todoModel := ts.From(req)

	// Call CreateFromForm
	createdTodo, err := ts.CreateFromForm(ctx, todoModel)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Validate the created todo
	if createdTodo.Name != "Test Todo" {
		t.Fatalf("expected todo name to be 'Test Todo', got %s", createdTodo.Name)
	}
	if createdTodo.Description != "This is a test todo" {
		t.Fatalf("expected todo description to be 'This is a test todo', got %s", createdTodo.Description)
	}

	// Check if the todo exists in the fake querier
	storedTodo, ok := fakeQuerier.Todos[createdTodo.Id]
	if !ok {
		t.Fatalf("expected todo to exist in fake querier")
	}

	// Validate stored todo details
	if storedTodo.Name != "Test Todo" {
		t.Fatalf("expected stored todo name to be 'Test Todo', got %s", storedTodo.Name)
	}

	if storedTodo.Description.String != "This is a test todo" {
		t.Fatalf("expected stored todo description to be 'This is a test todo', got %s", storedTodo.Description.String)
	}
}
