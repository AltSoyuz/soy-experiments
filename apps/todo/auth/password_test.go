package auth

import (
	"errors"
	"testing"
)

func TestVerfifyPasswordStrength(t *testing.T) {
	f := func(password string, expect error) {
		t.Helper()

		err := verifyPasswordStrength(password)

		if !errors.Is(err, expect) {
			t.Fatalf("unexpected error; got %v; want %v", err, expect)
		}

		if expect == nil {
			if err := verifyPasswordStrength(password); err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		}
	}

	f("", ErrWeakPassword)
	f("short", ErrWeakPassword)
	f("validpassword123", nil)
}
