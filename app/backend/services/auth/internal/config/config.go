// Package config defines and loads the auth service configuration from the environment.
package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	baseconfig "github.com/embedded-market/backend/pkg/config"
)

// Config is the fully-resolved auth service configuration.
type Config struct {
	baseconfig.Base
	Postgres baseconfig.Postgres
	Redis    baseconfig.Redis
	Kafka    baseconfig.Kafka

	JWT    JWTConfig
	OAuth  OAuthConfig
	Cookie CookieConfig
	Outbox OutboxConfig

	RateLimitRequests int           `env:"RATE_LIMIT_REQUESTS" envDefault:"60"`
	RateLimitWindow   time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`
}

// JWTConfig configures access-token signing/verification.
type JWTConfig struct {
	PrivateKeyPEM  string        `env:"JWT_PRIVATE_KEY_PEM"`
	PrivateKeyFile string        `env:"JWT_PRIVATE_KEY_FILE"`
	KeyID          string        `env:"JWT_KEY_ID" envDefault:"auth-key-1"`
	Issuer         string        `env:"JWT_ISSUER" envDefault:"embedded-market-auth"`
	Audience       string        `env:"JWT_AUDIENCE" envDefault:"embedded-market"`
	AccessTTL      time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	RefreshTTL     time.Duration `env:"JWT_REFRESH_TTL" envDefault:"720h"`

	privateKey *rsa.PrivateKey
}

// PrivateKey returns the parsed RSA signing key.
func (j JWTConfig) PrivateKey() *rsa.PrivateKey { return j.privateKey }

// OAuthConfig holds per-provider client credentials.
type OAuthConfig struct {
	GitHubClientID     string `env:"OAUTH_GITHUB_CLIENT_ID"`
	GitHubClientSecret string `env:"OAUTH_GITHUB_CLIENT_SECRET"`
	GitHubRedirectURL  string `env:"OAUTH_GITHUB_REDIRECT_URL"`

	GoogleClientID     string `env:"OAUTH_GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"OAUTH_GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `env:"OAUTH_GOOGLE_REDIRECT_URL"`

	GitLabClientID     string `env:"OAUTH_GITLAB_CLIENT_ID"`
	GitLabClientSecret string `env:"OAUTH_GITLAB_CLIENT_SECRET"`
	GitLabRedirectURL  string `env:"OAUTH_GITLAB_REDIRECT_URL"`
}

// CookieConfig configures the refresh-token and OAuth-state cookies.
type CookieConfig struct {
	RefreshName  string        `env:"COOKIE_REFRESH_NAME" envDefault:"em_refresh_token"`
	StateName    string        `env:"COOKIE_OAUTH_STATE_NAME" envDefault:"em_oauth_state"`
	Domain       string        `env:"COOKIE_DOMAIN"`
	Secure       bool          `env:"COOKIE_SECURE" envDefault:"true"`
	OAuthStateTTL time.Duration `env:"COOKIE_OAUTH_STATE_TTL" envDefault:"10m"`
}

// OutboxConfig configures the transactional-outbox relay.
type OutboxConfig struct {
	PollInterval time.Duration `env:"OUTBOX_POLL_INTERVAL" envDefault:"2s"`
	BatchSize    int32         `env:"OUTBOX_BATCH_SIZE" envDefault:"100"`
}

// Load parses the environment and resolves derived values (the RSA private key).
func Load() (*Config, error) {
	cfg := &Config{}
	cfg.ServiceName = "auth-service"
	if err := baseconfig.Load(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "auth-service"
	}

	pemBytes, err := resolvePrivateKeyPEM(cfg.JWT)
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse JWT private key: %w", err)
	}
	cfg.JWT.privateKey = key
	return cfg, nil
}

func resolvePrivateKeyPEM(j JWTConfig) ([]byte, error) {
	if j.PrivateKeyPEM != "" {
		return []byte(j.PrivateKeyPEM), nil
	}
	if j.PrivateKeyFile != "" {
		data, err := os.ReadFile(j.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read JWT private key file: %w", err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("a JWT signing key is required: set JWT_PRIVATE_KEY_PEM or JWT_PRIVATE_KEY_FILE")
}
