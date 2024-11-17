package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/argon2id"
	"log/slog"
	"net/http"
)

func handleLogout(as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := as.GetSessionFrom(r)
		if err != nil {
			slog.Error("error getting session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if s == nil {
			slog.Warn("no session found")
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		err = as.InvalidateSession(r.Context(), s.Id)
		if err != nil {
			slog.Error("error invalidating session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		auth.DeleteSessionCookie(w)

		w.Header().Set("HX-Redirect", "/login")
	}
}

func handleAuthWithPassword(render views.RenderFunc, as *auth.Service) http.HandlerFunc {
	type ErrorFragmentData struct {
		Message string
	}

	type LoginForm struct {
		Username string
		Password string
	}

	getLoginFrom := func(r *http.Request) (LoginForm, error) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("error parsing form", "error", err)
			return LoginForm{}, err
		}
		username := r.FormValue("username")
		password := r.FormValue("password")

		return LoginForm{Username: username, Password: password}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		form, err := getLoginFrom(r)
		if err != nil {
			slog.Error("error getting login form", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render(w, ErrorFragmentData{Message: "Invalid form data"}, "error-msg")
			return
		}

		user, err := as.GetUserByUsername(r.Context(), form.Username)
		if err != nil {
			slog.Error("error getting user", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			render(w, ErrorFragmentData{Message: "Invalid username or password"}, "error-msg")
			return
		}

		validPassword, err := argon2id.Verify(user.PasswordHash, form.Password)
		if err != nil {
			slog.Warn("error verifying password", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !validPassword {
			slog.Warn("invalid password", "username", form.Username)
			w.WriteHeader(http.StatusUnauthorized)
			render(w, ErrorFragmentData{Message: "Invalid username or password"}, "error-msg")
			return
		}

		token, err := as.CreateSession(r.Context(), user.ID)
		if err != nil {
			slog.Error("error creating session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session, _, err := as.ValidateSession(r.Context(), token)
		if err != nil {
			slog.Error("error validating session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		auth.SetSessionCookie(w, token, session.ExpiresAt)

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRenderRegisterView(render views.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Register"}, "register.html")
	}
}

func handleRenderLoginView(render views.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Login"}, "login.html")
	}
}
