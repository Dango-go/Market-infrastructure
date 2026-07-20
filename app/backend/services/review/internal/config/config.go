package config

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"

	baseconfig "github.com/embedded-market/backend/pkg/config"
)

type Config struct {
	baseconfig.Base
	Postgres baseconfig.Postgres
	JWT      JWTConfig
}

type JWTConfig struct {
	PublicKeyPEM  string `env:"JWT_PUBLIC_KEY_PEM"`
	PublicKeyFile string `env:"JWT_PUBLIC_KEY_FILE"`
	Issuer        string `env:"JWT_ISSUER" envDefault:"embedded-market-auth"`
	Audience      string `env:"JWT_AUDIENCE" envDefault:"embedded-market"`

	publicKey *rsa.PublicKey
}

func (j JWTConfig) PublicKey() *rsa.PublicKey { return j.publicKey }

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.ServiceName = "review-service"
	if err := baseconfig.Load(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "review-service"
	}
	pemBytes, err := resolvePublicKeyPEM(cfg.JWT)
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse JWT public key: %w", err)
	}
	cfg.JWT.publicKey = key
	return cfg, nil
}

func resolvePublicKeyPEM(j JWTConfig) ([]byte, error) {
	if j.PublicKeyPEM != "" {
		return []byte(j.PublicKeyPEM), nil
	}
	if j.PublicKeyFile != "" {
		data, err := os.ReadFile(j.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read JWT public key file: %w", err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("a JWT verification key is required: set JWT_PUBLIC_KEY_PEM or JWT_PUBLIC_KEY_FILE")
}
