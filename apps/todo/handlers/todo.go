package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/model"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/views"
	"log/slog"
	"net/http"
	"strconv"
)

func RenderTodoList(render views.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	type todoPageData struct {
		Title    string
		Items    []model.Todo
		Username string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetSessionUserFrom(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		todos, err := todoService.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		page := todoPageData{
			Title:    "My Todo List",
			Items:    todos,
			Username: user.Username,
		}
		render(w, page, "index.html")
	}
}

func CreateTodoFragment(render views.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := todoService.From(r)
		todo, err := todoService.CreateFromForm(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo created", "data", todo)
		render(w, t, "todo")
	}
}

func GetTodoFormFragment(render views.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
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

func UpdateTodoFragment(render views.RenderFunc, todoService *todo.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		todoForm := todoService.From(r)
		t, err := todoService.UpdateById(r.Context(), id, todoForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo updated", "name", t.Name)
		render(w, t, "todo")
	}
}

func DeleteTodo(todoService *todo.TodoService) http.HandlerFunc {
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
