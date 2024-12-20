package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AltSoyuz/soy-experiments/apps/todo/auth"
	"github.com/AltSoyuz/soy-experiments/apps/todo/model"
	"github.com/AltSoyuz/soy-experiments/apps/todo/todo"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web/forms"
	"github.com/AltSoyuz/soy-experiments/lib/httpserver"
)

func handleRenderTodoList(todoStore *todo.TodoStore, csrf *httpserver.CSRFProtection) http.HandlerFunc {
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

		var todoViewModels []web.TodoComponentData

		for _, todo := range todos {
			todoViewModels = append(todoViewModels, web.TodoComponentData{
				Todo:      todo,
				CSRFToken: csrf.GenerateToken(),
			})
		}

		page := web.TodoPageData{
			Title:     "My Todo List",
			Items:     todoViewModels,
			Email:     user.Email,
			CSRFToken: csrf.GenerateToken(),
		}

		web.RenderTodoList(w, page)
	}
}

func handleCreateTodoFragment(todoStore *todo.TodoStore, csrf *httpserver.CSRFProtection) http.HandlerFunc {
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

		csrfToken := csrf.GenerateToken()

		web.RenderTodoFragment(w, todo, csrfToken)
	}
}

func handleGetTodoFormFragment(todoStore *todo.TodoStore, csrf *httpserver.CSRFProtection) http.HandlerFunc {
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

		csrfToken := csrf.GenerateToken()
		web.RenderFormFragment(w, todo, csrfToken)
	}
}

func handleUpdateTodoFragment(todoStore *todo.TodoStore, csrf *httpserver.CSRFProtection) http.HandlerFunc {
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

		slog.Debug("updated todo", "todo", todo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		csrfToken := csrf.GenerateToken()
		web.RenderTodoFragment(w, todo, csrfToken)
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

func handleCompleteTodoFragment(todoStore *todo.TodoStore, csrf *httpserver.CSRFProtection) http.HandlerFunc {
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

		todo, err := todoStore.FindById(r.Context(), id, user.Id)
		if err != nil {
			slog.Error("error fetching todo", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		todo.IsComplete = !todo.IsComplete

		todo, err = todoStore.Update(r.Context(), todo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		csrfToken := csrf.GenerateToken()
		web.RenderTodoFragment(w, todo, csrfToken)
	}
}
