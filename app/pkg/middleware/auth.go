package middleware

import (
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/token"
	"github.com/gin-gonic/gin"
)

// TokenVerifier is the minimal verification port the auth middleware depends on. The
// shared *token.Verifier satisfies it; tests can substitute a fake.
type TokenVerifier interface {
	Verify(tokenString string) (*token.Claims, error)
}

// Authentication requires a valid Bearer access token, storing the claims and account id
// on the context. Requests without a valid token are rejected with 401.
func Authentication(v TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := extract(c, v)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		c.Set(ctxClaims, claims)
		c.Set(ctxAccountID, claims.Subject)
		c.Next()
	}
}

// OptionalAuthentication attaches claims when a valid token is present but never rejects
// the request — used for endpoints whose response varies for authenticated callers.
func OptionalAuthentication(v TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, err := extract(c, v); err == nil {
			c.Set(ctxClaims, claims)
			c.Set(ctxAccountID, claims.Subject)
		}
		c.Next()
	}
}

func extract(c *gin.Context, v TokenVerifier) (*token.Claims, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return nil, apperr.Unauthorized("missing_token", "authorization header is required")
	}
	scheme, raw, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || raw == "" {
		return nil, apperr.Unauthorized("invalid_auth_scheme", "authorization header must be a Bearer token")
	}
	claims, err := v.Verify(strings.TrimSpace(raw))
	if err != nil {
		return nil, apperr.Unauthorized("invalid_token", "the access token is invalid or expired")
	}
	return claims, nil
}
