package todo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AltSoyuz/soy-experiments/apps/todo/gen/db"
	"github.com/AltSoyuz/soy-experiments/apps/todo/store"
	"github.com/AltSoyuz/soy-experiments/apps/todo/todo"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web/forms"
)

func TestCreateFromForm(t *testing.T) {
	ctx := context.Background()
	fakeQuerier := store.NewFakeQuerier()
	ts := todo.Init(fakeQuerier)
	user, err := fakeQuerier.CreateUser(ctx, db.CreateUserParams{
		Email:        "testuser",
		PasswordHash: "testpassword",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Create a new HTTP request with form data
	req := httptest.NewRequest(http.MethodPost, "/todos", nil)
	req.Form = map[string][]string{
		"name":        {"Test Todo"},
		"description": {"This is a test todo"},
	}

	// Call CreateFromForm
	createdTodo, err := ts.CreateFromForm(ctx, forms.TodoForm{
		Name:        "Test Todo",
		Description: "This is a test todo",
	}, user.ID)

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

	if createdTodo.UserId != user.ID {
		t.Fatalf("expected todo user ID to be %d, got %d", user.ID, createdTodo.UserId)
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
