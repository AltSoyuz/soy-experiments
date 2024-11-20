package handlers

import (
	"fmt"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/model"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/web"
	"log/slog"
	"net/http"
	"strconv"
)

func handleRenderTodoList(render web.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
	type todoPageData struct {
		Title string
		Items []model.Todo
		Email string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetSessionUserFrom(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		todos, err := todoStore.List(r.Context(), user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		page := todoPageData{
			Title: "My Todo List",
			Items: todos,
			Email: user.Email,
		}
		render(w, page, "index.html")
	}
}

func handleCreateTodoFragment(render web.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
	getForm := func(r *http.Request) (model.Todo, error) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("error parsing form", "error", err)
			return model.Todo{}, err
		}
		name := r.FormValue("name")
		description := r.FormValue("description")
		if name == "" {
			return model.Todo{}, fmt.Errorf("name is required")
		}
		return model.Todo{
			Name:        name,
			Description: description,
		}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetSessionUserFrom(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		t, err := getForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		todo, err := todoStore.CreateFromForm(r.Context(), t, user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo created", "data", todo)
		render(w, todo, "todo")
	}
}

func handleGetTodoFormFragment(render web.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
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
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		t, err := todoStore.FindById(ctx, id, user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("todo form requested", "name", t.Name)
		render(w, t, "form")
	}
}

func handleUpdateTodoFragment(render web.RenderFunc, todoStore *todo.TodoStore) http.HandlerFunc {
	getForm := func(r *http.Request) (model.Todo, error) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("error parsing form", "error", err)
			return model.Todo{}, err
		}
		name := r.FormValue("name")
		description := r.FormValue("description")
		if name == "" {
			return model.Todo{}, fmt.Errorf("name is required")
		}
		return model.Todo{
			Name:        name,
			Description: description,
		}, nil
	}
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

		todoForm, err := getForm(r)
		todoForm.UserId = user.Id
		todoForm.Id = id

		t, err := todoStore.Update(r.Context(), todoForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo updated", "name", t.Name)
		render(w, t, "todo")
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info("todo deleted", "id", id)
		w.WriteHeader(http.StatusOK)
	}
}
