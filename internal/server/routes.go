package server

import (
	"golang-template-htmx-alpine/internal/handlers"
	"golang-template-htmx-alpine/internal/templates"
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	render templates.RenderFunc,
) {
	mux.HandleFunc("GET /healthz", handlers.HealthzHandler)
	mux.HandleFunc("GET /", handlers.TodoListHandler(render))
	mux.HandleFunc("POST /todos", handlers.CreateTodoHandler(logger, render))
	mux.HandleFunc("GET /todos/{name}/form", handlers.GetTodoFormHandler(logger, render))
	mux.HandleFunc("PUT /todos/{name}", handlers.UpdateTodoHandler(logger, render))
	mux.HandleFunc("DELETE /todos/{name}", handlers.DeleteTodoHandler(logger))
}
