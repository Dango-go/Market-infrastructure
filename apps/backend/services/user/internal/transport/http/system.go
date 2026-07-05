package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	ready func(ctx context.Context) error
}

func NewSystemHandler(ready func(ctx context.Context) error) *SystemHandler {
	return &SystemHandler{ready: ready}
}

func (h *SystemHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *SystemHandler) Readyz(c *gin.Context) {
	if err := h.ready(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
