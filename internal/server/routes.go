package server

import (
	"golang-template-htmx-alpine/internal/handlers"
	"golang-template-htmx-alpine/internal/templates"
	"golang-template-htmx-alpine/internal/todo"
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	render templates.RenderFunc,
	todoService *todo.TodoService,
) {
	mux.HandleFunc("GET /healthz", handlers.HealthzHandler)
	mux.HandleFunc("GET /", handlers.TodoListHandler(render, todoService))
	mux.HandleFunc("POST /todos", handlers.CreateTodoHandler(logger, render, todoService))
	mux.HandleFunc("GET /todos/{id}/form", handlers.GetTodoFormHandler(logger, render, todoService))
	mux.HandleFunc("PUT /todos/{id}", handlers.UpdateTodoHandler(logger, render, todoService))
	mux.HandleFunc("DELETE /todos/{id}", handlers.DeleteTodoHandler(logger, todoService))
}
