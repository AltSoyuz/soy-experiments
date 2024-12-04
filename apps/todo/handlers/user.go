package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/apps/todo/web/forms"
	"golang-template-htmx-alpine/lib/httpserver"
	"log/slog"
	"net/http"
)

// handleCreateUser creates a new user account and redirects to the login page.
func handleCreateUser(authService *auth.Service, csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, err := forms.RegisterFrom(r)
		if err != nil {
			csrfToken := csrf.GenerateToken()
			web.RenderRegister(w, csrfToken, err.Error())
			return
		}

		err = authService.RegisterUser(r.Context(), form.Email, form.Password)

		if err != nil {
			slog.Error("error registering user", "error", err)
			csrfToken := csrf.GenerateToken()
			web.RenderRegister(w, csrfToken, err.Error())
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}
