package auth

import (
	"context"
	"fmt"

	"github.com/AltSoyuz/soy-experiments/apps/todo/model"
	"github.com/AltSoyuz/soy-experiments/lib/argon2id"
)

func (as *Service) AuthenticateWithPassword(ctx context.Context, email, password string) (s model.Session, t string, err error) {
	user, err := as.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return model.Session{}, "", err
	}

	validPassword, err := argon2id.Verify(user.PasswordHash, password)
	if err != nil {
		return model.Session{}, "", err
	}

	if !validPassword {
		return model.Session{}, "", fmt.Errorf("invalid password")
	}

	token, err := as.createSession(ctx, user.ID)
	if err != nil {
		return model.Session{}, "", err
	}

	session, _, err := as.validateSession(ctx, token)
	if err != nil {
		return model.Session{}, "", err
	}

	return session, token, err
}
