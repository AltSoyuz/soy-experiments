package handlers

import (
	"log/slog"
	"net/http"

	"github.com/AltSoyuz/soy-experiments/apps/todo/auth"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web/forms"
	"github.com/AltSoyuz/soy-experiments/lib/httpserver"
)

// handleCreateUser creates a new user account and redirects to the login page.
func handleCreateUser(authService *auth.Service, csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, err := forms.RegisterFrom(r)
		if err != nil {
			csrfToken := csrf.GenerateToken()
			web.RenderRegisterForm(w, csrfToken, err.Error())
			return
		}

		err = authService.RegisterUser(r.Context(), form.Email, form.Password)

		if err != nil {
			slog.Error("error registering user", "error", err)
			csrfToken := csrf.GenerateToken()
			web.RenderRegisterForm(w, csrfToken, err.Error())
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}
