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

func handleAuthWithPassword(render *web.Renderer, authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		form, err := forms.LoginFrom(r)
		if err != nil {
			slog.Error("error getting login form", "error", err)
			render.ErrorFragment(w, "error getting login form")
			return
		}

		session, token, err := authService.AuthenticateWithPassword(ctx, form.Email, form.Password)
		if err != nil {
			slog.Error("error authenticating with password", "error", err)
			render.ErrorFragment(w, "error authenticating with password")
			return
		}

		auth.SetSessionCookie(w, token, session.ExpiresAt)

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRenderRegisterView(render *web.Renderer) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render.RenderPage(w, loginPageData{Title: "Register"}, "register.html")
	}
}

func handleRenderLoginView(render *web.Renderer) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render.RenderPage(w, loginPageData{Title: "Login"}, "login.html")
	}
}

func handleRenderVerifyEmail(render *web.Renderer) http.HandlerFunc {
	type verifyEmailPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render.RenderPage(w, verifyEmailPageData{Title: "Verify Email"}, "verify-email.html")
	}
}

func handleEmailVerification(render *web.Renderer, as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := auth.GetTokenFromCookie(r)

		form, err := forms.CodeFrom(r)
		if err != nil {
			slog.Error("error getting verification form", "error", err)
			render.ErrorFragment(w, "error getting verification form")
			return
		}

		err = as.VerifyEmail(ctx, token, form.Code)
		if err != nil {
			slog.Error("error verifying email", "error", err)
			render.ErrorFragment(w, err.Error())
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}
