package handlers

import (
	"log/slog"
	"net/http"

	"github.com/AltSoyuz/soy-experiments/apps/todo/auth"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web/forms"
	"github.com/AltSoyuz/soy-experiments/lib/httpserver"
)

func handleLogout(as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := as.GetSessionFrom(r)
		if err != nil {
			slog.Error("error getting session", "error", err)
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

func handleAuthWithPassword(authService *auth.Service, csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		form, err := forms.LoginFrom(r)
		if err != nil {
			slog.Error("error getting login form", "error", err)
			csrftoken := csrf.GenerateToken()
			web.RenderLoginForm(w, csrftoken, err.Error())
			return
		}

		session, token, err := authService.AuthenticateWithPassword(ctx, form.Email, form.Password)
		if err != nil {
			slog.Error("error authenticating with password", "error", err)
			csrftoken := csrf.GenerateToken()
			web.RenderLoginForm(w, csrftoken, err.Error())
			return
		}

		auth.SetSessionCookie(w, token, session.ExpiresAt)

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRenderRegisterView(csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := csrf.GenerateToken()
		web.RenderRegisterPage(w, csrfToken)
	}
}

func handleRenderLoginView(csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := csrf.GenerateToken()
		web.RenderLoginPage(w, csrfToken)
	}
}

func handleRenderVerifyEmail(csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := csrf.GenerateToken()
		web.RenderVerifyEmail(w, csrfToken)
	}
}

func handleEmailVerification(as *auth.Service, csrf *httpserver.CSRFProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := auth.GetTokenFromCookie(r)

		form, err := forms.CodeFrom(r)
		if err != nil {
			slog.Error("error getting verification form", "error", err)
			csrfToken := csrf.GenerateToken()
			web.RenderVerifyEmailForm(w, csrfToken, err.Error())
			return
		}

		err = as.VerifyEmail(ctx, token, form.Code)
		if err != nil {
			slog.Error("error verifying email", "error", err)
			csrfToken := csrf.GenerateToken()
			web.RenderVerifyEmailForm(w, csrfToken, err.Error())
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}
