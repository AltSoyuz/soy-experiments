package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/model"
	"log/slog"
	"net/http"
	"time"
)

// generateTokenSession creates a cryptographically secure session token
func generateTokenSession() (string, error) {
	bytes := make([]byte, 32) // Increased from 20 to 32 bytes for better security
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	// Using base64 instead of base32 for better entropy density
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// validateSession validates and optionally renews a session
func (as *Service) validateSession(ctx context.Context, token string) (model.Session, model.User, error) {
	if token == "" {
		return model.Session{}, model.User{}, ErrSessionInvalid
	}

	sessionId := hashToken(token)
	row, err := as.queries.ValidateSessionToken(ctx, sessionId)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Session{}, model.User{}, ErrSessionInvalid
		}
		slog.Error("database error", "error", err)
		return model.Session{}, model.User{}, fmt.Errorf("failed to validate session: %w", err)
	}

	now := time.Now()
	if now.Unix() >= row.ExpiresAt {
		if err := as.queries.DeleteSession(ctx, sessionId); err != nil {
			slog.Error("failed to delete expired session", "error", err)
			return model.Session{}, model.User{}, fmt.Errorf("failed to delete expired session: %w", err)
		}
		return model.Session{}, model.User{}, ErrSessionExpired
	}

	session := model.Session{
		Id:        row.ID,
		UserId:    row.UserID,
		ExpiresAt: row.ExpiresAt,
	}

	// Renew session if it's close to expiration
	if now.Unix() >= row.ExpiresAt-int64(sessionRenewalThreshold.Seconds()) {
		updatedSession, err := as.queries.UpdateSession(ctx, db.UpdateSessionParams{
			ExpiresAt: now.Add(sessionDuration).Unix(),
			ID:        session.Id,
		})
		if err != nil {
			return model.Session{}, model.User{}, fmt.Errorf("failed to renew session: %w", err)
		}
		session.ExpiresAt = updatedSession.ExpiresAt
	}

	emailVerified := false

	if row.EmailVerified != 0 {
		emailVerified = true
	}

	user := model.User{
		Id:            row.UserID,
		Email:         row.Email,
		EmailVerified: emailVerified,
	}

	return session, user, nil
}

// GetTokenFromCookie extracts the session token from the request cookie
func GetTokenFromCookie(r *http.Request) string {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		return cookie.Value
	}
	return ""
}

// SetSessionCookie sets the session token in the response cookie
func SetSessionCookie(w http.ResponseWriter, token string, expiresAt int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Required for production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(expiresAt, 0),
	})
}

// DeleteSessionCookie deletes the session cookie
func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// hashToken creates a SHA-256 hash of the session token
func hashToken(token string) string {
	hash := sha256.New()
	hash.Write([]byte(token))
	return hex.EncodeToString(hash.Sum(nil))
}

// createSession creates a new session for the given user
func (as *Service) createSession(ctx context.Context, userId int64) (string, error) {
	token, err := generateTokenSession()
	if err != nil {
		return "", err
	}

	sessionId := hashToken(token)
	expiresAt := time.Now().Add(sessionDuration)

	_, err = as.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:        sessionId,
		UserID:    userId,
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

// GetSessionFrom extracts the session from the request and validates it
func (as *Service) GetSessionFrom(r *http.Request) (model.Session, error) {
	token := GetTokenFromCookie(r)
	if token == "" {
		return model.Session{}, ErrSessionInvalid
	}
	session, _, err := as.validateSession(r.Context(), token)
	return session, err
}

// GetSessionUserFrom extracts the user from the request context
func GetSessionUserFrom(ctx context.Context) (model.User, bool) {
	user, ok := ctx.Value(UserContextKey).(model.User)
	return user, ok
}

// InvalidateSession deletes the session from the database
func (as *Service) InvalidateSession(ctx context.Context, sessionId string) error {
	err := as.queries.DeleteSession(ctx, sessionId)
	return err
}
