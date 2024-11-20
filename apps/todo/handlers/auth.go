package handlers

import (
	"fmt"
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/web"
	"golang-template-htmx-alpine/lib/argon2id"
	"log/slog"
	"net/http"
	"time"
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

func handleAuthWithPassword(render web.RenderFunc, as *auth.Service) http.HandlerFunc {
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
		username := r.FormValue("email")
		password := r.FormValue("password")

		return LoginForm{Username: username, Password: password}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		form, err := getLoginFrom(r)
		if err != nil {
			slog.Error("error getting login form", "error", err)
			render(w, ErrorFragmentData{Message: "Invalid form data"}, "error-msg")
			return
		}

		user, err := as.GetUserByEmail(r.Context(), form.Username)
		if err != nil {
			slog.Error("error getting user", "error", err)
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

func handleRenderRegisterView(render web.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Register"}, "register.html")
	}
}

func handleRenderLoginView(render web.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Login"}, "login.html")
	}
}

type ErrorFragmentData struct {
	Message string
}

func handleEmailVerificationRequest(as *auth.Service, render web.RenderFunc) http.HandlerFunc {
	type CodeForm struct {
		Code string `form:"code"`
	}

	getForm := func(r *http.Request) (CodeForm, error) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("error parsing form", "error", err)
			return CodeForm{}, fmt.Errorf("error parsing form: %w", err)
		}
		code := r.FormValue("code")
		if code == "" {
			return CodeForm{}, fmt.Errorf("no code provided")
		}
		return CodeForm{Code: code}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Validate session and token
		token := auth.GetTokenFromCookie(r)
		if token == "" {
			slog.Error("no token found")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		_, user, err := as.ValidateSession(ctx, token)
		if err != nil {
			slog.Error("session validation failed", "error", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Retrieve verification request
		verificationRequest, err := as.GetUserEmailVerificationRequest(ctx, user)
		if err != nil {
			slog.Error("error getting email verification request", "error", err)
			renderError(w, render, "Invalid verification request")
			return
		}

		// Check request expiration
		now := time.Now().Unix()
		if now >= verificationRequest.ExpiresAt {
			if err := as.Queries.DeleteUserEmailVerificationRequest(ctx, user.Id); err != nil {
				slog.Error("failed to delete expired verification request", "error", err)
			}
			renderError(w, render, "Verification code expired")
			return
		}

		// Verification form
		form, err := getForm(r)
		if err != nil {
			slog.Error("error getting verification form", "error", err)
			renderError(w, render, "Invalid verification code")
			return
		}

		// Validate verification code
		validCode, err := as.Queries.ValidateEmailVerificationRequest(ctx, db.ValidateEmailVerificationRequestParams{
			UserID:    user.Id,
			Code:      form.Code,
			ExpiresAt: now,
		})
		if err != nil || validCode.Code == "" {
			slog.Error("invalid verification code", "error", err)
			renderError(w, render, "Invalid verification code")
			return
		}

		// Mark email as verified
		if err := as.SetUserEmailVerified(ctx, user.Id); err != nil {
			slog.Error("failed to set email as verified", "error", err)
			renderError(w, render, "Unable to verify email")
			return
		}

		// Redirect on success
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}

// Helper function to render error messages
func renderError(w http.ResponseWriter, render web.RenderFunc, message string) {
	render(w, ErrorFragmentData{Message: message}, "error-msg")
}

func handleRenderVerifyEmail(render web.RenderFunc) http.HandlerFunc {
	type loginPageData struct {
		Title string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, loginPageData{Title: "Register"}, "verify-email.html")
	}
}
