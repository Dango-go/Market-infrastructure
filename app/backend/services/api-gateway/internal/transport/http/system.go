package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler { return &SystemHandler{} }

func (h *SystemHandler) Healthz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }

func (h *SystemHandler) Readyz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) }
