package handlers

import (
	"golang-template-htmx-alpine/internal/model"
	"golang-template-htmx-alpine/internal/store"
	"golang-template-htmx-alpine/internal/templates"
	"golang-template-htmx-alpine/internal/todo"
	"log/slog"
	"net/http"
)

func TodoListHandler(render templates.RenderFunc) http.HandlerFunc {
	type todoPageData struct {
		Title string
		Items []model.Todo
	}
	return func(w http.ResponseWriter, r *http.Request) {
		page := todoPageData{
			Title: "My Todo List",
			Items: store.Data,
		}
		render(w, page, "index.html")
	}
}

func CreateTodoHandler(logger *slog.Logger, render templates.RenderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := todo.CreateFromForm(r)
		todo.Add(t)
		logger.Info("todo created", "name", t.Name)
		render(w, t, "todo")
	}
}

func GetTodoFormHandler(logger *slog.Logger, render templates.RenderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		t := todo.FindByName(name)
		logger.Info("todo form requested", "name", t.Name)
		render(w, t, "form")
	}
}

func UpdateTodoHandler(logger *slog.Logger, render templates.RenderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		t := todo.CreateFromForm(r)
		todo.UpdateByName(name, t)
		logger.Info("todo updated", "name", t.Name)
		render(w, t, "todo")
	}
}

func DeleteTodoHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		todo.DeleteByName(name)
		logger.Info("todo deleted", "name", name)
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
