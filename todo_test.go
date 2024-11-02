package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCreateTodoFromForm(t *testing.T) {
	form := url.Values{}
	form.Add("name", "New Todo")
	form.Add("description", "New Description")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.PostForm = form

	todo := createTodoFromForm(req)

	if todo.Name != "New Todo" {
		t.Errorf("expected name to be 'New Todo', got '%s'", todo.Name)
	}
	if todo.Description != "New Description" {
		t.Errorf("expected description to be 'New Description', got '%s'", todo.Description)
	}
}

func TestFindTodoByName(t *testing.T) {
	todo := findTodoByName("Todo 1")

	if todo.Name != "Todo 1" {
		t.Errorf("expected name to be 'Todo 1', got '%s'", todo.Name)
	}
	if todo.Description != "Description 1" {
		t.Errorf("expected description to be 'Description 1', got '%s'", todo.Description)
	}

	todo = findTodoByName("Nonexistent Todo")
	if todo.Name != "" {
		t.Errorf("expected name to be '', got '%s'", todo.Name)
	}
}

func TestDeleteTodoByName(t *testing.T) {
	addTodo(Todo{Name: "Todo to Delete", Description: "Description to Delete"})
	deleteTodoByName("Todo to Delete")

	todo := findTodoByName("Todo to Delete")
	if todo.Name != "" {
		t.Errorf("expected name to be '', got '%s'", todo.Name)
	}
}

func TestAddTodo(t *testing.T) {
	newTodo := Todo{Name: "New Todo", Description: "New Description"}
	addTodo(newTodo)

	todo := findTodoByName("New Todo")
	if todo.Name != "New Todo" {
		t.Errorf("expected name to be 'New Todo', got '%s'", todo.Name)
	}
	if todo.Description != "New Description" {
		t.Errorf("expected description to be 'New Description', got '%s'", todo.Description)
	}
}

func TestUpdateTodoByName(t *testing.T) {
	addTodo(Todo{Name: "Todo to Update", Description: "Old Description"})
	updatedTodo := Todo{Name: "Todo to Update", Description: "Updated Description"}
	updateTodoByName("Todo to Update", updatedTodo)

	todo := findTodoByName("Todo to Update")
	if todo.Description != "Updated Description" {
		t.Errorf("expected description to be 'Updated Description', got '%s'", todo.Description)
	}
}
