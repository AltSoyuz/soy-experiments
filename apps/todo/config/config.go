package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

// Config defines the application configuration structure
type Config struct {
	SMTPHost    string `yaml:"smtp_host" env:"SMTP_HOST"`
	SMTPPort    int    `yaml:"smtp_port" env:"SMTP_PORT"`
	SenderEmail string `yaml:"sender_email" env:"SENDER_EMAIL"`
	SenderPass  string `yaml:"sender_pass" env:"SENDER_PASS"`
	Env         string `yaml:"env" env:"ENV"`
	Port        string `yaml:"port" env:"PORT"`
}

func Init(filepath string) (*Config, error) {
	config := &Config{}

	// Step 1: Load configuration from YAML file if it exists
	if filepath != "" {
		yamlConfig, err := loadFromYAML(filepath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("error loading YAML: %w", err)
			}
			// Ignore file not found error
		} else {
			// Merge YAML config into the final config
			config = yamlConfig
		}
	}

	// Step 2: Apply environment variable overrides
	if err := applyEnvOverrides(config); err != nil {
		return nil, fmt.Errorf("error applying env overrides: %w", err)
	}

	// Step 3: Validate final configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil

}

func loadFromYAML(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML file: %w", err)
	}
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) error {
	if env := os.Getenv("PORT"); env != "" {
		cfg.Port = env
	}

	if host := os.Getenv("SMTP_HOST"); host != "" {
		cfg.SMTPHost = host
	}

	if port := os.Getenv("SMTP_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return errors.New("invalid SMTP_PORT value")
		}
		cfg.SMTPPort = p
	}

	if email := os.Getenv("SENDER_EMAIL"); email != "" {
		cfg.SenderEmail = email
	}

	if pass := os.Getenv("SENDER_PASS"); pass != "" {
		cfg.SenderPass = pass
	}

	if env := os.Getenv("ENV"); env != "" {
		cfg.Env = env
	} else {
		cfg.Env = "dev"
	}
	return nil
}

func validateConfig(cfg *Config) error {
	if cfg.Port == "" {
		return errors.New("Port is required")
	}
	if cfg.SMTPHost == "" {
		return errors.New("SMTPHost is required")
	}
	if cfg.SMTPPort <= 0 {
		return errors.New("SMTPPort must be a positive integer")
	}
	if cfg.SenderEmail == "" {
		return errors.New("SenderEmail is required")
	}
	if cfg.SenderPass == "" {
		return errors.New("SenderPass is required")
	}
	return nil
}
