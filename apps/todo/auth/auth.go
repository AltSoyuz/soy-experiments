package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/model"
	"golang-template-htmx-alpine/lib/argon2id"
	"golang-template-htmx-alpine/lib/ratelimit"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var (
	ErrWeakPassword   = errors.New("password too weak or compromised")
	ErrSessionExpired = errors.New("session expired")
	ErrSessionInvalid = errors.New("invalid session")
)

const (
	minPasswordLength       = 12
	sessionDuration         = 24 * time.Hour
	sessionRenewalThreshold = 15 * 24 * time.Hour
	SessionCookieName       = "session"
)

type contextKey string

const userContextKey contextKey = "user"

type Service struct {
	Queries                 db.Querier
	LimitLoginMiddleware    func(http.Handler) http.HandlerFunc
	LimitRegisterMiddleware func(http.Handler) http.HandlerFunc
}

func Init(queries db.Querier) *Service {
	loginLimiter := ratelimit.With(5, time.Minute)
	registerLimiter := ratelimit.With(5, time.Minute)
	return &Service{
		Queries:                 queries,
		LimitLoginMiddleware:    ratelimit.LimitMiddleware(loginLimiter),
		LimitRegisterMiddleware: ratelimit.LimitMiddleware(registerLimiter),
	}
}

// VerifyPasswordStrength checks password strength and HIBP database
func VerifyPasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return ErrWeakPassword
	}

	// Check HIBP database
	passwordHashBytes := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(passwordHashBytes[:])
	hashPrefix := passwordHash[0:5]

	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", hashPrefix))
	if err != nil {
		return fmt.Errorf("failed to check password database: %w", err)
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		hashSuffix := strings.ToLower(scanner.Text()[:35])
		if subtle.ConstantTimeCompare([]byte(passwordHash), []byte(hashPrefix+hashSuffix)) == 1 {
			return ErrWeakPassword
		}
	}

	return nil
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
func (as *Service) ValidateSession(ctx context.Context, token string) (*model.Session, *model.User, error) {
	if token == "" {
		return nil, nil, ErrSessionInvalid
	}

	sessionId := hashToken(token)
	row, err := as.Queries.ValidateSessionToken(ctx, sessionId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrSessionInvalid
		}
		slog.Error("database error", "error", err)
		return nil, nil, fmt.Errorf("database error: %w", err)
	}

	now := time.Now()
	if now.Unix() >= row.ExpiresAt {
		if err := as.Queries.DeleteSession(ctx, sessionId); err != nil {
			slog.Error("failed to delete expired session", "error", err)
			return nil, nil, fmt.Errorf("failed to delete expired session: %w", err)
		}
		return nil, nil, ErrSessionExpired
	}

	session := &model.Session{
		Id:        row.ID,
		UserId:    row.UserID,
		ExpiresAt: row.ExpiresAt,
	}

	// Renew session if it's close to expiration
	if now.Unix() >= row.ExpiresAt-int64(sessionRenewalThreshold.Seconds()) {
		updatedSession, err := as.Queries.UpdateSession(ctx, db.UpdateSessionParams{
			ExpiresAt: now.Add(sessionDuration).Unix(),
			ID:        session.Id,
		})
		if err != nil {
			slog.Error("failed to renew session", "error", err)
			return nil, nil, fmt.Errorf("failed to renew session: %w", err)
		}
		session.ExpiresAt = updatedSession.ExpiresAt
	}

	user := &model.User{
		Id:       row.UserID,
		Username: row.Username,
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

// CreateSession creates a new session for the given user
func (as *Service) CreateSession(ctx context.Context, userId int64) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	sessionId := hashToken(token)
	expiresAt := time.Now().Add(sessionDuration)

	_, err = as.Queries.CreateSession(ctx, db.CreateSessionParams{
		ID:        sessionId,
		UserID:    userId,
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		slog.Error("failed to create session", "error", err)
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

// GetSessionFrom extracts the session from the request and validates it
func (as *Service) GetSessionFrom(r *http.Request) (*model.Session, error) {
	token := GetTokenFromCookie(r)
	if token == "" {
		return nil, ErrSessionInvalid
	}
	session, _, err := as.ValidateSession(r.Context(), token)
	return session, err
}

// GetSessionUserFrom extracts the user from the request context
func GetSessionUserFrom(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(userContextKey).(*model.User)
	return user, ok
}

// InvalidateSession deletes the session from the database
func (as *Service) InvalidateSession(ctx context.Context, sessionId string) error {
	err := as.Queries.DeleteSession(ctx, sessionId)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// ProtectedRouteMiddleware is a middleware that protects routes from unauthorized access
func (as *Service) ProtectedRouteMiddleware(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := GetTokenFromCookie(r)
		if token != "" {
			session, user, err := as.ValidateSession(r.Context(), token)
			if err == nil {
				// Store session info in context
				ctx := context.WithValue(r.Context(), userContextKey, user)
				r = r.WithContext(ctx)
				// Update cookie if session was renewed
				SetSessionCookie(w, token, session.ExpiresAt)
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

// GetUserByUsername retrieves a user by username
func (as *Service) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	row, err := as.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return row, fmt.Errorf("user not found: %w", err)
		}
		slog.Error("database error", "error", err)
		return row, fmt.Errorf("database error: %w", err)
	}

	return row, nil
}

func (as *Service) CreateUser(ctx context.Context, username, password string) error {
	if password == "" || len(password) > 127 {
		return fmt.Errorf("invalid password: %w", ErrWeakPassword)
	}

	err := VerifyPasswordStrength(password)
	if err != nil {
		return fmt.Errorf("password does not meet requirements: %w", err)
	}

	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	_, err = as.Queries.CreateUser(ctx, db.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}
