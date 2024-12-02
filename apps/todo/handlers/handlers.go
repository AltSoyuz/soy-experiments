package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/web"
	"net/http"
)

func AddRoutes(
	config *config.Config,
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
	mux.Handle("POST /users", limitRegister(handleCreateUser(authService)))
	mux.Handle("GET /login", handleRenderLoginView())
	mux.Handle("GET /register", handleRenderRegisterView())
	mux.Handle("POST /authenticate/password",
		limitLogin(handleAuthWithPassword(authService)),
	)
	mux.Handle("GET /logout", handleLogout(authService))
	mux.Handle("GET /verify-email", limitVerifyEmail(handleRenderVerifyEmail()))
	mux.Handle(
		"POST /email-verification-request",
		limitVerifyEmail(handleEmailVerification(authService)),
	)

	// Todos
	mux.Handle("GET /{$}", protect(handleRenderTodoList(todoStore)))
	mux.Handle("POST /todos", protect(handleCreateTodoFragment(todoStore)))
	mux.Handle("GET /todos/{id}/form", protect(handleGetTodoFormFragment(todoStore)))
	mux.Handle("PUT /todos/{id}", protect(handleUpdateTodoFragment(todoStore)))
	mux.Handle("DELETE /todos/{id}", protect(handleDeleteTodo(todoStore)))
	mux.Handle("PUT /todos/{id}/complete", protect(handleCompleteTodoFragment(todoStore)))
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
