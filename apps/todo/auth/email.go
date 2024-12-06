package auth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log/slog"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/AltSoyuz/soy-experiments/apps/todo/gen/db"
)

type EmailParams struct {
	To          []string
	Subject     string
	Body        string
	Attachments []string
}

func (as *Service) VerifyEmail(ctx context.Context, token, code string) error {
	_, user, err := as.validateSession(ctx, token)
	if err != nil {
		return err
	}

	verificationRequest, err := as.queries.GetUserEmailVerificationRequest(ctx, user.Id)
	if err != nil {
		return err
	}

	// Check request expiration
	now := time.Now().Unix()
	if now >= verificationRequest.ExpiresAt {
		if err := as.queries.DeleteUserEmailVerificationRequest(ctx, user.Id); err != nil {
			return err
		}
		return fmt.Errorf("email verification request expired")
	}

	// Validate verification code
	validCode, err := as.queries.ValidateEmailVerificationRequest(ctx, db.ValidateEmailVerificationRequestParams{
		UserID:    user.Id,
		Code:      code,
		ExpiresAt: now,
	})
	if err != nil || validCode.Code == "" {
		return fmt.Errorf("invalid email verification code")
	}

	// Mark email as verified
	if err := as.queries.SetUserEmailVerified(ctx, user.Id); err != nil {
		return err
	}

	return nil
}

// CreateAndSendVerificationEmail creates a new email verification request and sends the verification email
func (as *Service) CreateAndSendVerificationEmail(ctx context.Context, userId int64, email string) error {
	code := as.generateEmailVerificationCode()
	slog.Info("Generated code", "code", code, "email", email)

	_, err := as.queries.InsertUserEmailVerificationRequest(ctx, db.InsertUserEmailVerificationRequestParams{
		UserID:    userId,
		CreatedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		Code:      code,
	})
	if err != nil {
		return err
	}

	go as.sendVerificationEmailAsync(email, code, userId)
	return nil
}

// sendVerificationEmail sends a verification email to the given email address
func (s *Service) sendVerificationEmail(email, code string) error {
	emailHeaders := EmailParams{
		To:      []string{email},
		Subject: "Verify your email",
		Body:    fmt.Sprintf("Your verification code is: %s", code),
	}

	// Set up authentication information
	auth := smtp.PlainAuth(
		"",
		s.Config.SenderEmail,
		s.Config.SenderPass,
		s.Config.SMTPHost,
	)

	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = s.Config.SenderEmail
	headers["To"] = strings.Join(emailHeaders.To, ",")
	headers["Subject"] = emailHeaders.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	// Construct message
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + emailHeaders.Body

	if s.Config.Env != "prod" {
		return nil
	}

	// Send email in production
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.Config.SMTPHost, s.Config.SMTPPort),
		auth,
		s.Config.SenderEmail,
		emailHeaders.To,
		[]byte(message),
	)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

// isValidEmail checks if the given email is valid
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^.+@.+\..+$`)
	return emailRegex.MatchString(email)
}

// generateEmailVerificationCode generates a random 5-character email verification code
func (as *Service) generateEmailVerificationCode() string {
	if as.Config.Env == "test" {
		return TestEmailVerificationCode
	}

	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	// Using base32 encoding for better entropy density excluding padding. eg. "C4W5E8JY"
	code := base32.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567").WithPadding(base32.NoPadding).EncodeToString(bytes)
	return code
}

// sendVerificationEmailAsync sends a verification email asynchronously
func (as *Service) sendVerificationEmailAsync(email, code string, userID int64) {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := as.sendVerificationEmail(email, code); err != nil {
		slog.Error(
			"failed to send verification email",
			"error", err,
			"userId", userID,
			"email", email,
		)
	}
}
