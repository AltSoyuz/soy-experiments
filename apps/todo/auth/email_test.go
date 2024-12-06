package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AltSoyuz/soy-experiments/apps/todo/config"
	"github.com/AltSoyuz/soy-experiments/apps/todo/store"
)

func TestCreateAndSendVerificationEmail(t *testing.T) {
	c := givenTestConfig()
	fakeQuerier := store.NewFakeQuerier()
	as := Init(c, fakeQuerier)

	f := func(userId int64, email string, expect error) {
		t.Helper()

		ctx := context.Background()

		err := as.CreateAndSendVerificationEmail(ctx, userId, email)
		if !errors.Is(err, expect) {
			t.Fatalf("unexpected error; got %v; want %v", err, expect)
		}

		if expect == nil {
			if _, ok := fakeQuerier.EmailVerificationRequests[userId]; !ok {
				t.Fatalf("expected email verification request to be created")
			}

			if fakeQuerier.EmailVerificationRequests[userId].Code != TestEmailVerificationCode {
				t.Fatalf("expected email verification code to be set")
			}

			if fakeQuerier.EmailVerificationRequests[userId].ExpiresAt < time.Now().Add(10*time.Minute).Unix() {
				t.Fatalf("expected email verification code to be set")
			}

		}

	}

	f(1, "test@test.com", nil)

}
func TestSendVerificationEmail(t *testing.T) {
	c := givenTestConfig()
	fakeQuerier := store.NewFakeQuerier()
	as := Init(c, fakeQuerier)

	f := func(email, code string, expect error) {
		t.Helper()

		err := as.sendVerificationEmail(email, code)
		if !errors.Is(err, expect) {
			t.Fatalf("unexpected error; got %v; want %v", err, expect)
		}
	}

	f("test@test.com", "toto", nil)
}

func TestGenerateEmailVerificationCode(t *testing.T) {
	f := func(c *config.Config, expect error) {
		t.Helper()

		fakeQuerier := store.NewFakeQuerier()
		as := Init(c, fakeQuerier)

		code := as.generateEmailVerificationCode()

		if expect != nil {
			t.Fatalf("expected error, got: %v", code)
		}

		if len(code) != 8 {
			t.Fatalf("unexpected code length; got %d; want 8", len(code))
		}

		if c.Env == "test" && code != TestEmailVerificationCode {
			t.Fatalf("expected %s, got %s", TestEmailVerificationCode, code)
		}

		if c.Env == "production" && code == TestEmailVerificationCode {
			t.Fatalf("expected different code for production")
		}
	}

	f(givenTestConfig(), nil)
	f(&config.Config{Env: "production"}, nil)
}

func TestIsValidEmail(t *testing.T) {
	f := func(email string, expect bool) {
		t.Helper()

		valid := isValidEmail(email)
		if valid != expect {
			t.Fatalf("unexpected email validation; got %v; want %v", valid, expect)
		}
	}

	f("", false)
	f("test", false)
	f("test@", false)
	f("test@test", false)
	f("test@test.", false)
	f("test@test.com", true)
}
