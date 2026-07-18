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
	engine.Use(middleware.RequestContext(), middleware.HTTPMetrics("order"), middleware.Logging(d.Logger), middleware.Recovery())
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("order"))

	v1 := engine.Group("/api/v1/orders")
	v1.Use(middleware.Authentication(d.Verifier))
	v1.GET("", d.Handler.ListOrders)
	v1.GET("/:id", d.Handler.GetOrder)
	v1.POST("", d.Handler.CreateOrder)
	v1.PUT("/:id/status", d.Handler.UpdateStatus)
	return engine
}
