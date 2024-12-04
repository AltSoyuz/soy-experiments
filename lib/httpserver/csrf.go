package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type CSRFProtection struct {
	tokens       map[string]time.Time
	mutex        sync.RWMutex
	allowedHosts []string
}

func NewCSRFProtection(allowedHosts ...string) (*CSRFProtection, error) {
	// Normalize and validate allowed hosts
	normalizedHosts := make([]string, 0, len(allowedHosts))
	for _, host := range allowedHosts {
		// Remove protocol and trailing slashes
		parsedURL, err := url.Parse(host)
		if err != nil {
			return nil, err
		}
		normalizedHosts = append(normalizedHosts, parsedURL.Host)
	}

	return &CSRFProtection{
		tokens:       make(map[string]time.Time),
		allowedHosts: normalizedHosts,
	}, nil
}

// Check if origin is allowed
func (cs *CSRFProtection) isOriginAllowed(origin string) bool {
	// If no allowed hosts are specified, deny all
	if len(cs.allowedHosts) == 0 {
		return false
	}

	// Parse the origin
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Normalize the host
	originHost := originURL.Host

	// Check against allowed hosts
	for _, allowedHost := range cs.allowedHosts {
		if strings.EqualFold(originHost, allowedHost) {
			return true
		}
	}

	return false
}

// Generate a new CSRF token
func (cs *CSRFProtection) GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.tokens[token] = time.Now()

	return token
}

// Validate the CSRF token
func (cs *CSRFProtection) ValidateToken(token string) bool {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Clean up expired tokens
	now := time.Now()
	for t, createdAt := range cs.tokens {
		if now.Sub(createdAt) >= time.Hour {
			delete(cs.tokens, t)
		}
	}

	// Check if token exists and isn't too old (1 hour expiry)
	if createdAt, exists := cs.tokens[token]; exists {
		if now.Sub(createdAt) < time.Hour {
			return true
		}
	}
	return false
}

// ConsumeToken after successful validation
func (cs *CSRFProtection) ConsumeToken(token string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.tokens, token)
}

// CSRF Middleware
func (cs *CSRFProtection) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == "GET" || r.Method == "HEAD" ||
			r.Method == "OPTIONS" || r.Method == "TRACE" {
			next.ServeHTTP(w, r)
			return
		}

		// Check Origin header
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		// Validate origin or referer
		var validOrigin bool
		if origin != "" {
			validOrigin = cs.isOriginAllowed(origin)
		} else if referer != "" {
			validOrigin = cs.isOriginAllowed(referer)
		}

		if !validOrigin {
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return
		}

		// For state-changing methods, validate CSRF token
		var token string

		// Check token in different possible locations
		// 1. Form value
		token = r.PostFormValue("csrf_token")

		// 2. If not in form, check header
		if token == "" {
			token = r.Header.Get("X-CSRF-Token")
		}

		// 3. If not found elsewhere, check cookie
		if token == "" {
			cookie, err := r.Cookie("csrf_token")
			if err == nil {
				token = cookie.Value
			}
		}

		// Validate token
		if token == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		if !cs.ValidateToken(token) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		// Wrap the response writer to capture the status
		rwCapture := &responseWriterCapture{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(rwCapture, r)

		// Only consume the token if the request was successful
		if rwCapture.status >= 200 && rwCapture.status < 300 {
			cs.ConsumeToken(token)
		}
	})
}

// Helper struct to capture response status
type responseWriterCapture struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriterCapture) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriterCapture) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
