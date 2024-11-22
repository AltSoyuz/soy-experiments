package auth

import (
	"context"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/lib/argon2id"
)

// RegisterUser creates a new user with the given email and password
func (as *Service) RegisterUser(ctx context.Context, email, password string) error {
	if err := validateUserInput(email, password); err != nil {
		return err
	}

	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := as.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return err
	}

	if err := as.CreateAndSendVerificationEmail(ctx, user.ID, email); err != nil {
		return err
	}

	return nil
}

// validateUserInput validates the user input for email and password
func validateUserInput(email, password string) error {
	if password == "" || len(password) > 127 {
		return fmt.Errorf("invalid password: %w", ErrWeakPassword)
	}
	if email == "" || !isValidEmail(email) {
		return fmt.Errorf("invalid email: %w", ErrWeakPassword)
	}
	return verifyPasswordStrength(password)
}
