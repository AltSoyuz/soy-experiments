package config

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	f := func(yamlContent string, envVars map[string]string, wantConfig *Config, wantErr bool) {
		t.Helper()

		// Create temporary config file
		tmpfile, err := os.CreateTemp("", "config*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		// Write test YAML content
		if err := os.WriteFile(tmpfile.Name(), []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// Set environment variables
		for k, v := range envVars {
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}

		// Test Init function
		got, err := Init(tmpfile.Name())

		// Check error
		if (err != nil) != wantErr {
			t.Errorf("Init() error = %v, wantErr %v", err, wantErr)
			return
		}

		// If we expect an error, don't compare configs
		if wantErr {
			return
		}

		// Compare configs
		if got.SMTPHost != wantConfig.SMTPHost {
			t.Errorf("SMTPHost = %v, want %v", got.SMTPHost, wantConfig.SMTPHost)
		}
		if got.SMTPPort != wantConfig.SMTPPort {
			t.Errorf("SMTPPort = %v, want %v", got.SMTPPort, wantConfig.SMTPPort)
		}
		if got.SenderEmail != wantConfig.SenderEmail {
			t.Errorf("SenderEmail = %v, want %v", got.SenderEmail, wantConfig.SenderEmail)
		}
		if got.SenderPass != wantConfig.SenderPass {
			t.Errorf("SenderPass = %v, want %v", got.SenderPass, wantConfig.SenderPass)
		}
	}

	tests := []struct {
		name        string
		yamlContent string
		envVars     map[string]string
		wantConfig  *Config
		wantErr     bool
	}{
		{
			name: "Valid config from YAML",
			yamlContent: `
smtp_host: smtp.example.com
smtp_port: 587
sender_email: test@example.com
sender_pass: password123
`,
			envVars: nil,
			wantConfig: &Config{
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
				SenderEmail: "test@example.com",
				SenderPass:  "password123",
			},
			wantErr: false,
		},
		{
			name: "Environment variables override YAML",
			yamlContent: `
smtp_host: smtp.example.com
smtp_port: 587
sender_email: test@example.com
sender_pass: password123
`,
			envVars: map[string]string{
				"SMTP_HOST":    "smtp.override.com",
				"SMTP_PORT":    "465",
				"SENDER_EMAIL": "override@example.com",
				"SENDER_PASS":  "newpassword",
			},
			wantConfig: &Config{
				SMTPHost:    "smtp.override.com",
				SMTPPort:    465,
				SenderEmail: "override@example.com",
				SenderPass:  "newpassword",
			},
			wantErr: false,
		},
		{
			name: "Invalid YAML",
			yamlContent: `
invalid: yaml: content
`,
			envVars:    nil,
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "Missing required fields",
			yamlContent: `
smtp_host: ""
smtp_port: 0
`,
			envVars:    nil,
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "Invalid port in env vars",
			yamlContent: `
smtp_host: smtp.example.com
smtp_port: 587
sender_email: test@example.com
sender_pass: password123
`,
			envVars: map[string]string{
				"SMTP_PORT": "invalid",
			},
			wantConfig: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f(tt.yamlContent, tt.envVars, tt.wantConfig, tt.wantErr)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	f := func(config Config, wantErr bool) {
		t.Helper()

		err := validateConfig(&config)
		if (err != nil) != wantErr {
			t.Errorf("validateConfig() error = %v, wantErr %v", err, wantErr)
		}
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: Config{
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
				SenderEmail: "test@example.com",
				SenderPass:  "password123",
			},
			wantErr: false,
		},
		{
			name: "Missing SMTP host",
			config: Config{
				SMTPPort:    587,
				SenderEmail: "test@example.com",
				SenderPass:  "password123",
			},
			wantErr: true,
		},
		{
			name: "Invalid port",
			config: Config{
				SMTPHost:    "smtp.example.com",
				SMTPPort:    0,
				SenderEmail: "test@example.com",
				SenderPass:  "password123",
			},
			wantErr: true,
		},
		{
			name: "Missing sender email",
			config: Config{
				SMTPHost:   "smtp.example.com",
				SMTPPort:   587,
				SenderPass: "password123",
			},
			wantErr: true,
		},
		{
			name: "Missing sender password",
			config: Config{
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
				SenderEmail: "test@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f(tt.config, tt.wantErr)
		})
	}
}
