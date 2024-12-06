package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/AltSoyuz/soy-experiments/apps/todo/store"
)

func TestCreateUser(t *testing.T) {
	c := givenTestConfig()
	fakeQuerier := store.NewFakeQuerier()
	as := Init(c, fakeQuerier)

	f := func(email, password string, expect error) {
		t.Helper()

		ctx := context.Background()
		err := as.RegisterUser(ctx, email, password)
		if !errors.Is(err, expect) {
			t.Fatalf("unexpected error; got %v; want %v", err, expect)
		}

		if expect == nil {
			user, err := fakeQuerier.GetUserByEmail(ctx, email)
			if err != nil {
				t.Fatalf("expected user to be created")
			}

			if user.Email != email {
				t.Fatalf("expected email %s, got %s", email, user.Email)
			}

			if user.PasswordHash == "" {
				t.Fatalf("expected password hash to be set")
			}

			if fakeQuerier.EmailVerificationRequests[user.ID].UserID != user.ID {
				t.Fatalf("expected email verification request to be created")
			}

			if fakeQuerier.EmailVerificationRequests[user.ID].Code == "" {
				t.Fatalf("expected email verification code to be set")
			}
		}
	}

	f("user1@user.com", "", ErrWeakPassword)
	f("user2@user.com", "short", ErrWeakPassword)
	f("user3@user.com", "validpassword123", nil)
}
