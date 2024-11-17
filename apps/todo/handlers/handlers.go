package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/views"
	"net/http"
)

func AddRoutes(
	mux *http.ServeMux,
	render views.RenderFunc,
	authService *auth.Service,
	todoStore *todo.TodoStore,
) {
	// Middlewares
	limitRegister := authService.LimitRegisterMiddleware
	limitLogin := authService.LimitLoginMiddleware
	protect := authService.ProtectedRouteMiddleware

	// Health check
	mux.HandleFunc("GET /health", Healthz)

	// Authentication
	mux.Handle("POST /users", limitRegister(handleCreateUser(render, authService)))
	mux.Handle("GET /login", handleRenderLoginView(render))
	mux.Handle("GET /register", handleRenderRegisterView(render))
	mux.Handle("POST /authenticate/password", limitLogin(handleAuthWithPassword(render, authService)))
	mux.Handle("GET /logout", handleLogout(authService))

	// Todos
	mux.Handle("GET /", protect(handleRenderTodoList(render, todoStore)))
	mux.Handle("POST /todos", protect(handleCreateTodoFragment(render, todoStore)))
	mux.Handle("GET /todos/{id}/form", protect(handleGetTodoFormFragment(render, todoStore)))
	mux.Handle("PUT /todos/{id}", protect(handleUpdateTodoFragment(render, todoStore)))
	mux.Handle("DELETE /todos/{id}", protect(handleDeleteTodo(todoStore)))

	// mux.HandleFunc("GET /users/{id}",
	// mux.HandleFunc("DELETE /users/{id}",
	// mux.HandleFunc("POST /users/{id}/update-password",

	// mux.HandleFunc("POST /users/{id}/email-verification-request",
	// mux.HandleFunc("GET /users/{id}/email-verification-request",
	// mux.HandleFunc("DELETE /users/{id}/email-verification-request",
	// mux.HandleFunc("POST /users/{id}/very-email",
}
