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

func handleRenderTodoList(render views.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
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

		todos, err := todoStore.List(r.Context())
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

func handleCreateTodoFragment(render views.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := todoStore.From(r)
		todo, err := todoStore.CreateFromForm(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo created", "data", todo)
		render(w, t, "todo")
	}
}

func handleGetTodoFormFragment(render views.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		t, err := todoStore.FindById(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("todo form requested", "name", t.Name)
		render(w, t, "form")
	}
}

func handleUpdateTodoFragment(render views.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		todoForm := todoStore.From(r)
		t, err := todoStore.UpdateById(r.Context(), id, todoForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo updated", "name", t.Name)
		render(w, t, "todo")
	}
}

func handleDeleteTodo(todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		err = todoStore.DeleteById(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo deleted", "id", id)
		w.WriteHeader(http.StatusOK)
	}
}
