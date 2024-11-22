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
	renderer *web.Renderer,
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

	// Auth
	mux.Handle("POST /users", limitRegister(handleCreateUser(renderer, authService)))
	mux.Handle("GET /login", handleRenderLoginView(renderer))
	mux.Handle("GET /register", handleRenderRegisterView(renderer))
	mux.Handle("POST /authenticate/password",
		limitLogin(handleAuthWithPassword(renderer, authService)),
	)
	mux.Handle("GET /logout", handleLogout(authService))
	mux.Handle("GET /verify-email", limitVerifyEmail(handleRenderVerifyEmail(renderer)))
	mux.Handle(
		"POST /email-verification-request",
		limitVerifyEmail(handleEmailVerification(renderer, authService)),
	)

	// Todos
	mux.Handle("GET /", protect(handleRenderTodoList(renderer, todoStore)))
	mux.Handle("POST /todos", protect(handleCreateTodoFragment(renderer, todoStore)))
	mux.Handle("GET /todos/{id}/form", protect(handleGetTodoFormFragment(renderer, todoStore)))
	mux.Handle("PUT /todos/{id}", protect(handleUpdateTodoFragment(renderer, todoStore)))
	mux.Handle("DELETE /todos/{id}", protect(handleDeleteTodo(todoStore)))

	// mux.HandleFunc("GET /users/{id}",
	// mux.HandleFunc("DELETE /users/{id}",
	// mux.HandleFunc("POST /users/{id}/update-password",

	// mux.HandleFunc("DELETE /users/{id}/email-verification-request",
	// mux.HandleFunc("POST /users/{id}/very-email",
}
