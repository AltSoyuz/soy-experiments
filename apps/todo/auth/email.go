package auth

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailParams struct {
	To          []string
	Subject     string
	Body        string
	Attachments []string
}

func (s *Service) SendVerificationEmail(email, code string) error {
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

	if s.Config.Env == "prod" {
		// Send email
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

	}

	return nil
}
