package handlers

import (
	"database/sql"
	"errors"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/argon2id"
	"log/slog"
	"net/http"
)

func Logout(sm *auth.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sm.GetSessionFrom(r)
		if err != nil {
			slog.Error("error getting session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if s.Session == nil {
			slog.Warn("no session found")
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		err = sm.InvalidateSession(r.Context(), s.Session.Id)
		if err != nil {
			slog.Error("error invalidating session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		auth.DeleteSessionCookie(w)

		w.Header().Set("HX-Redirect", "/login")
		return
	}
}

func AuthWithPassword(render views.RenderFunc, queries *db.Queries, sm *auth.SessionManager) http.HandlerFunc {
	type Data struct {
		Message string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := queries.GetUserByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				slog.Warn("user not found", "username", username)
				w.WriteHeader(http.StatusUnauthorized)
				render(w, Data{Message: "Invalid username or password"}, "error-msg")
				return
			}
			slog.Warn("error getting user", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		validPassword, err := argon2id.Verify(user.PasswordHash, password)
		if err != nil {
			slog.Warn("error verifying password", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !validPassword {
			slog.Warn("invalid password", "username", username)
			w.WriteHeader(http.StatusUnauthorized)
			render(w, Data{Message: "Invalid username or password"}, "error-msg")
			return
		}

		token, err := sm.CreateSession(r.Context(), user.ID)
		if err != nil {
			slog.Error("error creating session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result, err := sm.ValidateSession(r.Context(), token)
		if err != nil {
			slog.Error("error validating session", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		auth.SetSessionCookie(w, token, result.Session.ExpiresAt)

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func RenderRegisterView(render views.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Register"}, "register.html")
	}
}

func RenderLoginView(render views.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Login"}, "login.html")
	}
}
