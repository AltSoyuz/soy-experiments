package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/apps/todo/web/forms"
	"log/slog"
	"net/http"
)

func handleLogout(as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := as.GetSessionFrom(r)
		if err != nil {
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

func handleAuthWithPassword(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		form, err := forms.LoginFrom(r)
		if err != nil {
			slog.Error("error getting login form", "error", err)
			web.RenderError(w, "error getting login form")
			return
		}

		session, token, err := authService.AuthenticateWithPassword(ctx, form.Email, form.Password)
		if err != nil {
			slog.Error("error authenticating with password", "error", err)
			web.RenderError(w, "error authenticating with password")
			return
		}

		auth.SetSessionCookie(w, token, session.ExpiresAt)

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRenderRegisterView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.RenderRegister(w)
	}
}

func handleRenderLoginView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.RenderLogin(w)
	}
}

func handleRenderVerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.RenderVerifyEmail(w)
	}
}

func handleEmailVerification(as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := auth.GetTokenFromCookie(r)

		form, err := forms.CodeFrom(r)
		if err != nil {
			slog.Error("error getting verification form", "error", err)
			web.RenderError(w, "error getting verification form")
			return
		}

		err = as.VerifyEmail(ctx, token, form.Code)
		if err != nil {
			slog.Error("error verifying email", "error", err)
			web.RenderError(w, err.Error())
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}
