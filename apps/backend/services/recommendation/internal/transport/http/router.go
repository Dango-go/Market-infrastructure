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

	v1 := engine.Group("/api/v1/recommendations")
	v1.GET("/related/:productID", d.Handler.Related)
	v1.GET("/trending", d.Handler.Trending)
	v1.POST("/events", d.Handler.TrackEvent)

	protected := v1.Group("")
	protected.Use(middleware.Authentication(d.Verifier))
	protected.POST("/profiles", d.Handler.UpsertProfile)
	return engine
}
