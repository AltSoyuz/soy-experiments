package auth

import (
	"context"
	"encoding/base64"
	"golang-template-htmx-alpine/apps/todo/model"
	"golang-template-htmx-alpine/apps/todo/queries"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHashToken(t *testing.T) {
	f := func(token, expect string) {
		t.Helper()

		hashed := hashToken(token)
		if hashed != expect {
			t.Fatalf("unexpected hash; got %s; want %s", hashed, expect)
		}
	}

	// empty token
	f("", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	// simple token
	f("simpletoken", "bf4109b50feca5285c8e46692ae952a37d5864f075076be987122eb2d22e7eae")

	// complex token
	f("complex_token_123!@#", "f112f84eb2765718b633e6ee35850c5b063e2916559b63e7263bd9f3a9711533")
}

func TestCreateSession(t *testing.T) {
	fakeQuerier := queries.NewFakeQuerier()
	sm := NewSessionManager(fakeQuerier)

	ctx := context.Background()
	userId := int64(123)

	// Call CreateSession
	token, err := sm.CreateSession(ctx, userId)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Use the hashed token to check in the fake querier
	hashedToken := hashToken(token)
	session, ok := fakeQuerier.Sessions[hashedToken]
	if !ok {
		t.Fatalf("expected session to exist in fake querier")
	}

	// Validate session details
	if session.UserID != userId {
		t.Fatalf("expected session to have user ID %d, got %d", userId, session.UserID)
	}

	// Optionally validate other session details like ExpiresAt
	if session.ExpiresAt == 0 {
		t.Fatalf("expected session to have ExpiresAt set")
	}
}

func TestValidateSession(t *testing.T) {
	fakeQuerier := queries.NewFakeQuerier()
	sm := NewSessionManager(fakeQuerier)

	ctx := context.Background()
	userId := int64(123)

	// Create a session to validate
	token, err := sm.CreateSession(ctx, userId)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Validate the session
	result, err := sm.ValidateSession(ctx, token)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Session.ExpiresAt == 0 {
		t.Fatalf("expected session to be renewed")
	}

	// Validate session details
	if result.Session.UserId != userId {
		t.Fatalf("expected session to have user ID %d, got %d", userId, result.Session.UserId)
	}

	// Validate user details
	if result.User.Id != userId {
		t.Fatalf("expected user ID %d, got %d", userId, result.User.Id)
	}

	// Test expired session
	session := fakeQuerier.Sessions[hashToken(token)]
	session.ExpiresAt = time.Now().Add(-time.Hour).Unix()
	fakeQuerier.Sessions[hashToken(token)] = session
	_, err = sm.ValidateSession(ctx, token)

	if err != ErrSessionExpired {
		t.Fatalf("expected ErrSessionExpired, got: %v", err)
	}
}

func TestGenerateSessionToken(t *testing.T) {
	token, err := generateSessionToken()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(token) == 0 {
		t.Fatalf("expected token to be generated")
	}

	// Check if the token is base64 URL encoded
	_, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(token)
	if err != nil {
		t.Fatalf("expected token to be base64 URL encoded, got: %v", err)
	}

	// Generate another token and ensure it's different
	token2, err := generateSessionToken()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if token == token2 {
		t.Fatalf("expected different tokens, got the same")
	}
}

func TestGetTokenFromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// No cookie set
	token := GetTokenFromCookie(req)
	if token != "" {
		t.Fatalf("expected empty token, got: %s", token)
	}

	// Set cookie and test
	expectedToken := "testtoken"
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: expectedToken,
	})

	token = GetTokenFromCookie(req)
	if token != expectedToken {
		t.Fatalf("expected token %s, got: %s", expectedToken, token)
	}
}

func TestSetSessionCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	token := "testtoken"
	expiresAt := time.Now().Add(sessionDuration).Unix()

	SetSessionCookie(rr, token, expiresAt)

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got: %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != sessionCookieName {
		t.Fatalf("expected cookie name %s, got: %s", sessionCookieName, cookie.Name)
	}
	if cookie.Value != token {
		t.Fatalf("expected cookie value %s, got: %s", token, cookie.Value)
	}
	if cookie.Expires.Unix() != expiresAt {
		t.Fatalf("expected cookie expires at %d, got: %d", expiresAt, cookie.Expires.Unix())
	}
	if !cookie.HttpOnly {
		t.Fatalf("expected HttpOnly to be true")
	}
	if !cookie.Secure {
		t.Fatalf("expected Secure to be true")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite to be %v, got: %v", http.SameSiteStrictMode, cookie.SameSite)
	}
}

func TestDeleteSessionCookie(t *testing.T) {
	rr := httptest.NewRecorder()

	DeleteSessionCookie(rr)

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got: %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != sessionCookieName {
		t.Fatalf("expected cookie name %s, got: %s", sessionCookieName, cookie.Name)
	}
	if cookie.Value != "" {
		t.Fatalf("expected cookie value to be empty, got: %s", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("expected MaxAge to be -1, got: %d", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Fatalf("expected HttpOnly to be true")
	}
	if !cookie.Secure {
		t.Fatalf("expected Secure to be true")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite to be %v, got: %v", http.SameSiteStrictMode, cookie.SameSite)
	}
}

func TestGetSessionFrom(t *testing.T) {
	f := func(r *http.Request, expect string) {
		t.Helper()

		token := GetTokenFromCookie(r)
		if token != expect {
			t.Fatalf("unexpected token; got %s; want %s", token, expect)
		}
	}

	// No cookie set
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	f(req, "")

	// Set cookie and test
	expectedToken := "testtoken"
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: expectedToken,
	})

	f(req, expectedToken)
}
func TestInvalidateSession(t *testing.T) {
	fakeQuerier := queries.NewFakeQuerier()
	sm := NewSessionManager(fakeQuerier)

	ctx := context.Background()
	userId := int64(123)

	// Create a session to invalidate
	token, err := sm.CreateSession(ctx, userId)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Invalidate the session
	sessionId := hashToken(token)
	err = sm.InvalidateSession(ctx, sessionId)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check if the session is deleted
	_, ok := fakeQuerier.Sessions[sessionId]
	if ok {
		t.Fatalf("expected session to be deleted")
	}
}
func TestGetSessionUserFrom(t *testing.T) {
	// Create a context with a valid session
	user := &model.User{Id: 123, Username: "testuser"}
	session := SessionValidationResult{
		Session: &model.Session{
			Id:        "session-id",
			UserId:    user.Id,
			ExpiresAt: time.Now().Add(sessionDuration).Unix(),
		},
		User: user,
	}
	ctx := context.WithValue(context.Background(), "session", session)

	// Test with valid session in context
	retrievedUser, ok := GetSessionUserFrom(ctx)
	if !ok {
		t.Fatalf("expected to retrieve user from context")
	}
	if retrievedUser.Id != user.Id {
		t.Fatalf("expected user ID %d, got %d", user.Id, retrievedUser.Id)
	}

	if retrievedUser.Username != "testuser" {
		t.Fatalf("expected username %s, got %s", "testuser", retrievedUser.Username)
	}

	// Test with no session in context
	ctx = context.Background()
	retrievedUser, ok = GetSessionUserFrom(ctx)
	if ok {
		t.Fatalf("expected not to retrieve user from context")
	}
	if retrievedUser != nil {
		t.Fatalf("expected retrieved user to be nil, got %v", retrievedUser)
	}
}
