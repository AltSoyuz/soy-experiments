package auth

import (
	"github.com/AltSoyuz/soy-experiments/apps/todo/config"
)

func givenTestConfig() *config.Config {
	c := &config.Config{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		SenderEmail: "sender@example.com",
		SenderPass:  "password",
		Env:         "test",
	}
	return c
}
