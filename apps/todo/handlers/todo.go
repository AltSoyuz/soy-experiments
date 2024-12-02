package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/model"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/apps/todo/web/forms"
	"log/slog"
	"net/http"
	"strconv"
)

func handleRenderTodoList(todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := auth.GetSessionUserFrom(ctx)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		todos, err := todoStore.List(r.Context(), user.Id)
		if err != nil {
			slog.Error("error getting todos", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		page := web.TodoPageData{
			Title: "My Todo List",
			Items: todos,
			Email: user.Email,
		}

		web.RenderTodoList(w, page)
	}
}

func handleCreateTodoFragment(todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetSessionUserFrom(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		form, err := forms.TodoCreateFrom(r)
		if err != nil {
			slog.Error("error getting todo form", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		todo, err := todoStore.CreateFromForm(r.Context(), form, user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		web.RenderTodoFragment(w, todo)
	}
}

func handleGetTodoFormFragment(todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := auth.GetSessionUserFrom(ctx)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			slog.Error("error parsing id", "error", err)
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		todo, err := todoStore.FindById(ctx, id, user.Id)
		if err != nil {
			slog.Error("error fetching todo", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		web.RenderFormFragment(w, todo)
	}
}

func handleUpdateTodoFragment(todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetSessionUserFrom(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		form, err := forms.TodoUpdateFrom(r)
		if err != nil {
			slog.Error("error getting todo form", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		todo, err := todoStore.Update(r.Context(), model.Todo{
			Id:          id,
			Name:        form.Name,
			Description: form.Description,
			UserId:      user.Id,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		web.RenderTodoFragment(w, todo)
	}
}

func handleDeleteTodo(todoStore *todo.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetSessionUserFrom(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		err = todoStore.Delete(r.Context(), id, user.Id)
		if err != nil {
			slog.Error("error deleting todo", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
