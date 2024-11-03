package server

import (
	"golang-template-htmx-alpine/internal/templates"
	"golang-template-htmx-alpine/internal/todo"
	"log/slog"
	"net/http"
)

func New(
	logger *slog.Logger,
	render templates.RenderFunc,
	todoService *todo.TodoService,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, render, todoService)

	var handler http.Handler = mux
	return handler
}
