package server

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/handlers"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/httpserver"
	"net/http"
)

func New(
	render views.RenderFunc,
	authService *auth.Service,
	todoStore *todo.TodoStore,
) http.Handler {
	mux := http.NewServeMux()

	handlers.AddRoutes(
		mux,
		render,
		authService,
		todoStore,
	)

	return httpserver.CSRFProtection(mux)
}
