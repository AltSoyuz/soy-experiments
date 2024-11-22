package auth

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// verifyPasswordStrength checks password strength and HIBP database
func verifyPasswordStrength(password string) error {
	// Basic length check
	if len(password) < minPasswordLength {
		return ErrWeakPassword
	}

	// Check if password exists in HIBP database
	if err := checkHIBPDatabase(password); err != nil {
		return err
	}

	return nil
}

// checkHIBPDatabase checks if password exists in the HIBP database
func checkHIBPDatabase(password string) error {
	// Generate SHA256 hash of password
	hash := generatePasswordHash(password)
	hashPrefix := hash[0:5]

	// Query HIBP API
	matches, err := queryHIBPAPI(hashPrefix)
	if err != nil {
		return fmt.Errorf("HIBP API check failed: %w", err)
	}

	// Check if password hash exists in response
	if isHashCompromised(hash, hashPrefix, matches) {
		return ErrWeakPassword
	}

	return nil
}

func generatePasswordHash(password string) string {
	passwordHashBytes := sha256.Sum256([]byte(password))
	return hex.EncodeToString(passwordHashBytes[:])
}

func queryHIBPAPI(hashPrefix string) (*bufio.Scanner, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", hashPrefix))
	if err != nil {
		return nil, err
	}
	return bufio.NewScanner(res.Body), nil
}

func isHashCompromised(fullHash, prefix string, scanner *bufio.Scanner) bool {
	for scanner.Scan() {
		hashSuffix := strings.ToLower(scanner.Text()[:35])
		if subtle.ConstantTimeCompare([]byte(fullHash), []byte(prefix+hashSuffix)) == 1 {
			return true
		}
	}
	return false
}
