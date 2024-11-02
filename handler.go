package main

import (
	"log/slog"
	"net/http"
)

func todoListHandler(render renderFunc) http.HandlerFunc {
	type todoPageData struct {
		Title string
		Items []Todo
	}
	return func(w http.ResponseWriter, r *http.Request) {
		page := todoPageData{
			Title: "My Todo List",
			Items: todos,
		}
		render(w, page, "index.html")
	}
}

func createTodoHandler(logger *slog.Logger, render renderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		todo := createTodoFromForm(r)
		addTodo(todo)
		logger.Info("todo created", "name", todo.Name)
		render(w, todo, "todo")
	}
}

func getTodoFormHandler(logger *slog.Logger, render renderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		todo := findTodoByName(name)
		logger.Info("todo form requested", "name", todo.Name)
		render(w, todo, "form.html")
	}
}

func updateTodoHandler(logger *slog.Logger, render renderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		todo := createTodoFromForm(r)
		updateTodoByName(name, todo)
		logger.Info("todo updated", "name", todo.Name)
		render(w, todo, "todo")
	}
}

func deleteTodoHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		deleteTodoByName(name)
		logger.Info("todo deleted", "name", name)
		w.WriteHeader(http.StatusOK)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
