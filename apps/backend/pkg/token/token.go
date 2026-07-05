// Package token defines the shared JWT claims contract and an RS256 verifier used by the
// API gateway and every service to validate access tokens locally — without calling the
// auth service per request. The auth service holds the private key and issues tokens;
// everyone else verifies with the public key (PEM or JWKS).
package token

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the access-token payload shared platform-wide. Standard fields live in
// RegisteredClaims (sub = account id, iss, aud, exp, iat); custom fields carry identity
// details so downstream services avoid an extra lookup.
type Claims struct {
	jwt.RegisteredClaims
	Email     string `json:"email"`
	Username  string `json:"username"`
	SessionID string `json:"sid"`
	TokenType string `json:"typ"`
}

// AccessTokenType marks tokens intended for resource access (vs. refresh).
const AccessTokenType = "access"

// ErrInvalidToken is returned for any verification failure.
var ErrInvalidToken = errors.New("invalid token")

// Verifier validates signed access tokens against an RSA public key.
type Verifier struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

// NewVerifier builds a Verifier from a parsed RSA public key.
func NewVerifier(pub *rsa.PublicKey, issuer, audience string) *Verifier {
	return &Verifier{publicKey: pub, issuer: issuer, audience: audience}
}

// NewVerifierFromPEM parses a PEM-encoded RSA public key and builds a Verifier.
func NewVerifierFromPEM(pemBytes []byte, issuer, audience string) (*Verifier, error) {
	pub, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa public key: %w", err)
	}
	return NewVerifier(pub, issuer, audience), nil
}

// Verify parses and validates a token string, enforcing RS256, issuer, audience and the
// access token type. It returns the typed claims on success.
func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)
	tkn, err := parser.ParseWithClaims(tokenString, claims, func(*jwt.Token) (any, error) {
		return v.publicKey, nil
	})
	if err != nil || !tkn.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != AccessTokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
