package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SystemHandler serves operational endpoints: liveness, readiness, and the JWKS document
// used by other services to verify access tokens.
type SystemHandler struct {
	jwks  func() any
	ready func(ctx context.Context) error
}

// NewSystemHandler builds the system handler from a JWKS provider and a readiness probe.
func NewSystemHandler(jwks func() any, ready func(ctx context.Context) error) *SystemHandler {
	return &SystemHandler{jwks: jwks, ready: ready}
}

// Healthz reports process liveness.
func (h *SystemHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz reports whether dependencies are reachable.
func (h *SystemHandler) Readyz(c *gin.Context) {
	if err := h.ready(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// JWKS serves the JSON Web Key Set for the access-token signing key.
func (h *SystemHandler) JWKS(c *gin.Context) {
	c.JSON(http.StatusOK, h.jwks())
}
