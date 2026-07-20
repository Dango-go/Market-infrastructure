package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/embedded-market/backend/pkg/middleware"
)

type RouterDeps struct {
	Handler  *Handler
	System   *SystemHandler
	Verifier middleware.TokenVerifier
	Logger   *zap.Logger
}

func NewRouter(d RouterDeps) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.RequestContext(), middleware.Logging(d.Logger), middleware.Recovery())
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)

	v1 := engine.Group("/api/v1/notifications")
	v1.Use(middleware.Authentication(d.Verifier))
	v1.GET("", d.Handler.ListNotifications)
	v1.GET("/:id", d.Handler.GetNotification)
	v1.POST("", d.Handler.CreateNotification)
	v1.POST("/:id/sent", d.Handler.MarkSent)
	v1.POST("/:id/read", d.Handler.MarkRead)

	templates := engine.Group("/api/v1/notification-templates")
	templates.GET("", d.Handler.ListTemplates)
	protectedTemplates := templates.Group("")
	protectedTemplates.Use(middleware.Authentication(d.Verifier))
	protectedTemplates.POST("", d.Handler.CreateTemplate)
	return engine
}
