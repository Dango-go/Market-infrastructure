// Package middleware holds Gin middleware shared by every service: request-id /
// correlation-id propagation, structured request logging, panic recovery, rate limiting
// and JWT authentication.
package middleware

import (
	"github.com/embedded-market/backend/pkg/token"
	"github.com/gin-gonic/gin"
)

// Gin context keys and propagated header names.
const (
	HeaderRequestID     = "X-Request-Id"
	HeaderCorrelationID = "X-Correlation-Id"

	ctxRequestID     = "request_id"
	ctxCorrelationID = "correlation_id"
	ctxClaims        = "auth_claims"
	ctxAccountID     = "account_id"
)

// ClaimsFromContext returns the authenticated claims, if the request passed auth.
func ClaimsFromContext(c *gin.Context) (*token.Claims, bool) {
	v, ok := c.Get(ctxClaims)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*token.Claims)
	return claims, ok
}

// AccountID returns the authenticated account id (subject), or "" if unauthenticated.
func AccountID(c *gin.Context) string {
	if v, ok := c.Get(ctxAccountID); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// RequestID returns the current request id.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxRequestID); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// CorrelationID returns the current correlation id, propagated into emitted events.
func CorrelationID(c *gin.Context) string {
	if v, ok := c.Get(ctxCorrelationID); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
