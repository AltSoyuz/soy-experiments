package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/lib/httpserver"
	"net/http"
)

func AddRoutes(
	config *config.Config,
	csrf *httpserver.CSRFProtection,
	mux *http.ServeMux,
	authService *auth.Service,
	todoStore *todo.TodoStore,
) {
	// Middlewares
	limitRegister := authService.LimitRegisterMiddleware
	limitLogin := authService.LimitLoginMiddleware
	limitVerifyEmail := authService.LimitVerifyEmailMiddleware
	protect := authService.ProtectedRouteMiddleware

	// Health check
	mux.HandleFunc("GET /healthz", healthz)
	mux.Handle("GET /about", renderAboutView())
	mux.Handle("GET /404", notFoundView())
	mux.Handle("/", notFoundView())

	// Auth
	mux.Handle("POST /users", limitRegister(handleCreateUser(authService, csrf)))
	mux.Handle("GET /login", handleRenderLoginView(csrf))
	mux.Handle("GET /register", handleRenderRegisterView(csrf))
	mux.Handle("POST /authenticate/password",
		limitLogin(handleAuthWithPassword(authService, csrf)),
	)
	mux.Handle("GET /logout", handleLogout(authService))
	mux.Handle("GET /verify-email", limitVerifyEmail(handleRenderVerifyEmail(csrf)))
	mux.Handle(
		"POST /email-verification-request",
		limitVerifyEmail(handleEmailVerification(authService, csrf)),
	)

	// Todos
	mux.Handle("GET /{$}", protect(handleRenderTodoList(todoStore, csrf)))
	mux.Handle("POST /todos", protect(handleCreateTodoFragment(todoStore, csrf)))
	mux.Handle("GET /todos/{id}/form", protect(handleGetTodoFormFragment(todoStore, csrf)))
	mux.Handle("PUT /todos/{id}", protect(handleUpdateTodoFragment(todoStore, csrf)))
	mux.Handle("DELETE /todos/{id}", protect(handleDeleteTodo(todoStore)))
	mux.Handle("PUT /todos/{id}/complete", protect(handleCompleteTodoFragment(todoStore, csrf)))
}

func notFoundView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.RenderNotFound(w)
	}
}

func renderAboutView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.RenderAbout(w)
	}
}
