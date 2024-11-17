package main

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/httpserver"
	"net/http"
)

func newServer(
	render views.RenderFunc,
	queries *db.Queries,
	todoService *todo.TodoService,
	sessionManager *auth.SessionManager,
	limiter *auth.Limiter,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		render,
		queries,
		todoService,
		sessionManager,
		limiter,
	)

	return httpserver.CSRFProtection(mux)
}
