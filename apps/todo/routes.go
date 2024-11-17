package main

import (
	"context"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/handlers"
	"golang-template-htmx-alpine/apps/todo/todo"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/ratelimit"
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	render views.RenderFunc,
	queries *db.Queries,
	ts *todo.TodoService,
	sm *auth.SessionManager,
	limiter *auth.Limiter,
) {
	protect := protectedRoute(sm)
	limitRegister := limitedRoute(limiter.RegisterLimiter)
	limitLogin := limitedRoute(limiter.LoginLimiter)

	mux.HandleFunc("GET /health", healthHandler)

	// Authentication
	mux.HandleFunc("POST /users", limitRegister(handlers.CreateUser(render, queries)))
	mux.HandleFunc("GET /login", handlers.RenderLoginView(render))
	mux.HandleFunc("GET /register", handlers.RenderRegisterView(render))
	mux.HandleFunc("POST /authenticate/password", limitLogin(handlers.AuthWithPassword(render, queries, sm)))
	mux.HandleFunc("GET /logout", handlers.Logout(sm))

	// Todos
	mux.HandleFunc("GET /", protect(handlers.RenderTodoList(render, ts)))
	mux.HandleFunc("POST /todos", protect(handlers.CreateTodoFragment(render, ts)))
	mux.HandleFunc("GET /todos/{id}/form", protect(handlers.GetTodoFormFragment(render, ts)))
	mux.HandleFunc("PUT /todos/{id}", protect(handlers.UpdateTodoFragment(render, ts)))
	mux.HandleFunc("DELETE /todos/{id}", protect(handlers.DeleteTodo(ts)))

	// mux.HandleFunc("GET /users/{id}",
	// mux.HandleFunc("DELETE /users/{id}",
	// mux.HandleFunc("POST /users/{id}/update-password",

	// mux.HandleFunc("POST /users/{id}/email-verification-request",
	// mux.HandleFunc("GET /users/{id}/email-verification-request",
	// mux.HandleFunc("DELETE /users/{id}/email-verification-request",
	// mux.HandleFunc("POST /users/{id}/very-email",
}

func protectedRoute(sm *auth.SessionManager) func(h http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := auth.GetTokenFromCookie(r)
			if token != "" {
				result, err := sm.ValidateSession(r.Context(), token)
				if err == nil {
					// Store session info in context
					ctx := context.WithValue(r.Context(), "session", result)
					r = r.WithContext(ctx)
					// Update cookie if session was renewed
					auth.SetSessionCookie(w, token, result.Session.ExpiresAt)
				} else if err == auth.ErrSessionExpired || err == auth.ErrSessionInvalid {
					auth.DeleteSessionCookie(w)
				}
			} else {
				// Redirect if no token is present
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func limitedRoute(limiter *ratelimit.RateLimited) func(h http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			limiter := limiter.GetVisitorLimiter(ip)
			// Check if request is allowed
			if !limiter.Allow() {
				slog.Warn("rate limit exceeded", "ip", ip)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	slog.Info("health check")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
