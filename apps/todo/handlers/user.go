package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/views"
	"log/slog"
	"net/http"
)

// handleCreateUser creates a new user account and redirects to the login page.
func handleCreateUser(render views.RenderFunc, authService *auth.Service) http.HandlerFunc {
	type ErrorFragmentData struct {
		Message string
	}
	type RegisterForm struct {
		Username string
		Password string
	}
	getRegisterForm := func(r *http.Request) (RegisterForm, error) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("error parsing form", "error", err)
			return RegisterForm{}, err
		}
		username := r.FormValue("username")
		password := r.FormValue("password")

		return RegisterForm{Username: username, Password: password}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		form, err := getRegisterForm(r)
		if err != nil {
			slog.Error("error getting register form", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render(w, ErrorFragmentData{Message: "Invalid form data"}, "error-msg")
			return
		}

		err = authService.CreateUser(r.Context(), form.Username, form.Password)

		if err != nil {
			slog.Error("error creating user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render(w, ErrorFragmentData{Message: "Password does not meet requirements"}, "error-msg")
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}
