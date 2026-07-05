// Package token provides the RS256 implementation of the domain TokenService: it signs
// access tokens with a private key, mints/hashes opaque refresh tokens, and publishes the
// corresponding public key as a JWKS so every other service can verify tokens locally.
package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	pkgtoken "github.com/embedded-market/backend/pkg/token"
	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// Config configures the RS256 token service.
type Config struct {
	PrivateKey *rsa.PrivateKey
	KeyID      string
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Service signs access tokens and manages refresh tokens.
type Service struct {
	cfg Config
}

// NewService builds a token Service.
func NewService(cfg Config) *Service { return &Service{cfg: cfg} }

var _ domain.TokenService = (*Service)(nil)

// IssueAccessToken signs a short-lived RS256 access token carrying the shared claims.
func (s *Service) IssueAccessToken(account *domain.Account, sessionID uuid.UUID, now time.Time) (domain.IssuedToken, error) {
	expiresAt := now.Add(s.cfg.AccessTTL)
	claims := pkgtoken.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   account.ID.String(),
			Audience:  jwt.ClaimStrings{s.cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		Email:     account.Email.String(),
		Username:  account.Username.String(),
		SessionID: sessionID.String(),
		TokenType: pkgtoken.AccessTokenType,
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tkn.Header["kid"] = s.cfg.KeyID
	signed, err := tkn.SignedString(s.cfg.PrivateKey)
	if err != nil {
		return domain.IssuedToken{}, fmt.Errorf("sign access token: %w", err)
	}
	return domain.IssuedToken{Value: signed, ExpiresAt: expiresAt}, nil
}

// GenerateRefreshToken returns a 256-bit opaque token (base64url) and its SHA-256 hash.
func (s *Service) GenerateRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	plaintext := base64.RawURLEncoding.EncodeToString(raw)
	return plaintext, s.HashRefreshToken(plaintext), nil
}

// HashRefreshToken returns the hex-encoded SHA-256 of a refresh token. SHA-256 (not a slow
// hash) is appropriate because the token is high-entropy and not human-chosen.
func (s *Service) HashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// RefreshTTL reports the configured refresh-token lifetime.
func (s *Service) RefreshTTL() time.Duration { return s.cfg.RefreshTTL }

// JSONWebKey is one entry of a JWKS document.
type JSONWebKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JSONWebKeySet is the document served at /.well-known/jwks.json.
type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

// JWKS renders the public half of the signing key as a JWKS document.
func (s *Service) JWKS() JSONWebKeySet {
	pub := s.cfg.PrivateKey.Public().(*rsa.PublicKey)
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	return JSONWebKeySet{Keys: []JSONWebKey{{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: s.cfg.KeyID,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}}}
}
