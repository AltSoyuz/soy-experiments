package main

import (
	"golang-template-htmx-alpine/apps/todo/views"
	"log/slog"
	"net/http"
	"strconv"
)

func todoListHandler(render views.RenderFunc, todoService *TodoService) http.HandlerFunc {
	type todoPageData struct {
		Title string
		Items []Todo
	}

	slog.Info("todo list handler")

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

func createTodoHandler(render views.RenderFunc, todoService *TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := todoService.FromRequest(r)
		todo, err := todoService.CreateFromForm(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo created", "data", todo)
		render(w, t, "todo")
	}
}

func getTodoFormHandler(render views.RenderFunc, todoService *TodoService) http.HandlerFunc {
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
		slog.Info("todo form requested", "name", t.Name)
		render(w, t, "form")
	}
}

func updateTodoHandler(render views.RenderFunc, todoService *TodoService) http.HandlerFunc {
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
		slog.Info("todo updated", "name", t.Name)
		render(w, t, "todo")
	}
}

func deleteTodoHandler(todoService *TodoService) http.HandlerFunc {
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
		slog.Info("todo deleted", "id", id)
		w.WriteHeader(http.StatusOK)
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	slog.Info("health check")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
