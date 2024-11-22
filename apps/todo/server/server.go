package server

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/handlers"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/lib/httpserver"
	"net/http"
)

func New(
	config *config.Config,
	render *web.Renderer,
	authService *auth.Service,
	todoStore *todo.TodoStore,
) http.Handler {
	mux := http.NewServeMux()

	handlers.AddRoutes(
		config,
		mux,
		render,
		authService,
		todoStore,
	)

	return httpserver.CSRFProtection(mux)
}
