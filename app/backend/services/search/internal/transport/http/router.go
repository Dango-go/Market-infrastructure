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
	engine.Use(middleware.RequestContext(), middleware.HTTPMetrics("search"), middleware.Logging(d.Logger), middleware.Recovery())
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("search"))

	v1 := engine.Group("/api/v1/search")
	v1.GET("", d.Handler.Search)
	v1.GET("/suggestions", d.Handler.Suggest)

	protected := v1.Group("")
	protected.Use(middleware.Authentication(d.Verifier))
	protected.POST("/documents", d.Handler.UpsertDocument)
	protected.DELETE("/documents/:id", d.Handler.DeleteDocument)
	return engine
}
