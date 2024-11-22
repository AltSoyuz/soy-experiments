package auth

import (
	"context"
	"errors"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/lib/ratelimit"
	"net/http"
	"time"
)

var (
	ErrWeakPassword           = errors.New("password too weak or compromised")
	ErrSessionExpired         = errors.New("session expired")
	ErrSessionInvalid         = errors.New("invalid session")
	TestEmailVerificationCode = "12345678"
)

const (
	minPasswordLength       = 12
	sessionDuration         = 24 * time.Hour
	sessionRenewalThreshold = 15 * 24 * time.Hour
	SessionCookieName       = "session"
)

type contextKey string

const UserContextKey contextKey = "user"

type Service struct {
	Config                     *config.Config
	queries                    db.Querier
	LimitLoginMiddleware       func(http.Handler) http.HandlerFunc
	LimitRegisterMiddleware    func(http.Handler) http.HandlerFunc
	LimitVerifyEmailMiddleware func(http.Handler) http.HandlerFunc
}

func Init(config *config.Config, queries db.Querier) *Service {
	loginLimiter := ratelimit.With(5, time.Minute)
	registerLimiter := ratelimit.With(5, time.Minute)
	verifyEmailLimiter := ratelimit.With(5, time.Minute)
	return &Service{
		Config:                     config,
		queries:                    queries,
		LimitLoginMiddleware:       ratelimit.LimitMiddleware(loginLimiter),
		LimitRegisterMiddleware:    ratelimit.LimitMiddleware(registerLimiter),
		LimitVerifyEmailMiddleware: ratelimit.LimitMiddleware(verifyEmailLimiter),
	}
}

// ProtectedRouteMiddleware is a middleware that protects routes from unauthorized access
func (as *Service) ProtectedRouteMiddleware(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := GetTokenFromCookie(r)
		if token != "" {
			session, user, err := as.validateSession(r.Context(), token)

			if err == nil {
				// Store session info in context
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				r = r.WithContext(ctx)
				// Update cookie if session was renewed
				SetSessionCookie(w, token, session.ExpiresAt)

				// Redirect to email verification if email is not verified
				if !user.EmailVerified {
					http.Redirect(w, r, "/verify-email", http.StatusFound)
					return
				}
			} else if err == ErrSessionExpired || err == ErrSessionInvalid {
				DeleteSessionCookie(w)
			}
		} else {
			// Redirect if no token is present
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}
