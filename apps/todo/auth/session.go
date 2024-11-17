package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/model"
	"net/http"
	"time"
)

var (
	ErrSessionExpired = errors.New("session expired")
	ErrSessionInvalid = errors.New("invalid session")
)

const (
	sessionDuration         = 24 * time.Hour
	sessionRenewalThreshold = 15 * 24 * time.Hour
	sessionCookieName       = "session"
)

type SessionManager struct {
	queries db.Querier
}

type SessionValidationResult struct {
	Session *model.Session
	User    *model.User
}

// NewSessionManager creates a new session manager instance
func NewSessionManager(queries db.Querier) *SessionManager {
	return &SessionManager{queries: queries}
}

// generateSessionToken creates a cryptographically secure session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // Increased from 20 to 32 bytes for better security
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	// Using base64 instead of base32 for better entropy density
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// ValidateSession validates and optionally renews a session
func (sm *SessionManager) ValidateSession(ctx context.Context, token string) (SessionValidationResult, error) {
	if token == "" {
		return SessionValidationResult{}, ErrSessionInvalid
	}

	sessionId := hashToken(token)
	row, err := sm.queries.ValidateSessionToken(ctx, sessionId)
	if err != nil {
		if err == sql.ErrNoRows {
			return SessionValidationResult{}, ErrSessionInvalid
		}
		return SessionValidationResult{}, fmt.Errorf("database error: %w", err)
	}

	now := time.Now()
	if now.Unix() >= row.ExpiresAt {
		if err := sm.queries.DeleteSession(ctx, sessionId); err != nil {
			return SessionValidationResult{}, fmt.Errorf("failed to delete expired session: %w", err)
		}
		return SessionValidationResult{}, ErrSessionExpired
	}

	session := &model.Session{
		Id:        row.ID,
		UserId:    row.UserID,
		ExpiresAt: row.ExpiresAt,
	}

	// Renew session if it's close to expiration
	if now.Unix() >= row.ExpiresAt-int64(sessionRenewalThreshold.Seconds()) {
		updatedSession, err := sm.queries.UpdateSession(ctx, db.UpdateSessionParams{
			ExpiresAt: now.Add(sessionDuration).Unix(),
			ID:        session.Id,
		})
		if err != nil {
			return SessionValidationResult{}, fmt.Errorf("failed to renew session: %w", err)
		}
		session.ExpiresAt = updatedSession.ExpiresAt
	}

	return SessionValidationResult{
		Session: session,
		User: &model.User{
			Id:       row.UserID,
			Username: row.Username,
		},
	}, nil
}

// Cookie handling
func GetTokenFromCookie(r *http.Request) string {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value
	}
	return ""
}

func SetSessionCookie(w http.ResponseWriter, token string, expiresAt int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Required for production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(expiresAt, 0),
	})
}

func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
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

// CreateSession creates a new session for the given user
func (sm *SessionManager) CreateSession(ctx context.Context, userId int64) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	sessionId := hashToken(token)
	expiresAt := time.Now().Add(sessionDuration)

	_, err = sm.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:        sessionId,
		UserID:    userId,
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

func (sm *SessionManager) GetSessionFrom(r *http.Request) (SessionValidationResult, error) {
	token := GetTokenFromCookie(r)
	if token == "" {
		return SessionValidationResult{}, ErrSessionInvalid
	}
	return sm.ValidateSession(r.Context(), token)
}

func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionId string) error {
	err := sm.queries.DeleteSession(ctx, sessionId)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func GetSessionUserFrom(ctx context.Context) (*model.User, bool) {
	session, ok := ctx.Value("session").(SessionValidationResult)
	return session.User, ok
}
