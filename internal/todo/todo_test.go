package todo

import (
	"net/http"
	"strings"
	"testing"

	"golang-template-htmx-alpine/internal/model"
)

func TestCreateFromForm(t *testing.T) {
	formData := "name=Todo+2&description=Description+2"
	req, err := http.NewRequest("POST", "/", strings.NewReader(formData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	todo := CreateFromForm(req)
	if todo.Name != "Todo 2" || todo.Description != "Description 2" {
		t.Errorf("Expected Todo 2, Description 2, got %s, %s", todo.Name, todo.Description)
	}
}

func TestFindByName(t *testing.T) {
	todo := FindByName("Todo 1")
	if todo.Name != "Todo 1" || todo.Description != "Description 1" {
		t.Errorf("Expected Todo 1, Description 1, got %s, %s", todo.Name, todo.Description)
	}

	todo = FindByName("Nonexistent")
	if todo.Name != "" || todo.Description != "" {
		t.Errorf("Expected empty todo, got %s, %s", todo.Name, todo.Description)
	}
}

func TestDeleteByName(t *testing.T) {
	Add(model.Todo{Name: "Todo 3", Description: "Description 3"})
	DeleteByName("Todo 3")
	todo := FindByName("Todo 3")
	if todo.Name != "" || todo.Description != "" {
		t.Errorf("Expected empty todo, got %s, %s", todo.Name, todo.Description)
	}
}

func TestAdd(t *testing.T) {
	todo := model.Todo{Name: "Todo 4", Description: "Description 4"}
	Add(todo)
	found := FindByName("Todo 4")
	if found.Name != "Todo 4" || found.Description != "Description 4" {
		t.Errorf("Expected Todo 4, Description 4, got %s, %s", found.Name, found.Description)
	}
}

func TestUpdateByName(t *testing.T) {
	todo := model.Todo{Name: "Todo 5", Description: "Description 5"}
	Add(todo)
	updatedTodo := model.Todo{Name: "Todo 5", Description: "Updated Description"}
	UpdateByName("Todo 5", updatedTodo)
	found := FindByName("Todo 5")
	if found.Description != "Updated Description" {
		t.Errorf("Expected Updated Description, got %s", found.Description)
	}
}
