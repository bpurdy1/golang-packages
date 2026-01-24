package awsclient

import (
	"github.com/caarlos0/env/v11"
)

// Config holds AWS configuration loaded from environment variables.
type Config struct {
	Region          string `env:"AWS_REGION" envDefault:"us-east-1"`
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	SessionToken    string `env:"AWS_SESSION_TOKEN"`
	Endpoint        string `env:"AWS_ENDPOINT"` // For localstack/testing
}

// LoadConfig loads AWS configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
