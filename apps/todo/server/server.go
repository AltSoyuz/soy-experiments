package server

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/handlers"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/lib/httpserver"
	"net/http"
)

func New(
	config *config.Config,
	csrf *httpserver.CSRFProtection,
	authService *auth.Service,
	todoStore *todo.TodoStore,
) http.Handler {
	mux := http.NewServeMux()

	handlers.AddRoutes(
		config,
		csrf,
		mux,
		authService,
		todoStore,
	)

	return csrf.Middleware(mux)
}
