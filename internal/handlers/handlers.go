package handlers

import (
	"golang-template-htmx-alpine/internal/model"
	"golang-template-htmx-alpine/internal/templates"
	"golang-template-htmx-alpine/internal/todo"
	"log/slog"
	"net/http"
	"strconv"
)

func TodoListHandler(render templates.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	type todoPageData struct {
		Title string
		Items []model.Todo
	}

	return func(w http.ResponseWriter, r *http.Request) {
		todos, err := todoService.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		page := todoPageData{
			Title: "My Todo List",
			Items: todos,
		}
		render(w, page, "index.html")
	}
}

func CreateTodoHandler(logger *slog.Logger, render templates.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := todoService.FromRequest(r)
		todo, err := todoService.CreateFromForm(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logger.Info("todo created", "data", todo)
		render(w, t, "todo")
	}
}

func GetTodoFormHandler(logger *slog.Logger, render templates.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		t, err := todoService.FindById(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("todo form requested", "name", t.Name)
		render(w, t, "form")
	}
}

func UpdateTodoHandler(logger *slog.Logger, render templates.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		todoForm := todoService.FromRequest(r)
		t, err := todoService.UpdateById(r.Context(), id, todoForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logger.Info("todo updated", "name", t.Name)
		render(w, t, "todo")
	}
}

func DeleteTodoHandler(logger *slog.Logger, todoService *todo.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		err = todoService.DeleteById(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logger.Info("todo deleted", "id", id)
		w.WriteHeader(http.StatusOK)
	}
}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
