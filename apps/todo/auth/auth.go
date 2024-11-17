package auth

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"golang-template-htmx-alpine/lib/ratelimit"
	"net/http"
	"strings"
	"time"
)

var (
	ErrWeakPassword = errors.New("password too weak or compromised")
)

const (
	minPasswordLength = 12
)

type Limiter struct {
	LoginLimiter    *ratelimit.RateLimited
	RegisterLimiter *ratelimit.RateLimited
}

func NewLimiter() *Limiter {
	return &Limiter{
		LoginLimiter:    ratelimit.New(),
		RegisterLimiter: ratelimit.New(),
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
