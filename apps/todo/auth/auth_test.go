package auth

import (
	"testing"
)

func TestVerifyPasswordStrength(t *testing.T) {
	f := func(password string, expect error) {
		t.Helper()

		err := VerifyPasswordStrength(password)
		if err != expect {
			t.Fatalf("unexpected error; got %v; want %v", err, expect)
		}
	}

	f("tata", ErrWeakPassword)

	f("tata123456789", nil)
}
