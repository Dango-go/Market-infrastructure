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
	engine.Use(middleware.RequestContext(), middleware.HTTPMetrics("payment"), middleware.Logging(d.Logger), middleware.Recovery())
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("payment"))

	v1 := engine.Group("/api/v1/payments")
	v1.Use(middleware.Authentication(d.Verifier))
	v1.GET("", d.Handler.ListPayments)
	v1.GET("/:id", d.Handler.GetPayment)
	v1.POST("", d.Handler.CreatePayment)
	v1.POST("/:id/confirm", d.Handler.ConfirmPayment)
	v1.POST("/:id/fail", d.Handler.FailPayment)
	v1.POST("/:id/refund", d.Handler.RefundPayment)
	return engine
}
