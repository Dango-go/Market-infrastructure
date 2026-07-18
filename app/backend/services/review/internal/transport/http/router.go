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
	engine.Use(middleware.RequestContext(), middleware.HTTPMetrics("review"), middleware.Logging(d.Logger), middleware.Recovery())
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("review"))

	v1 := engine.Group("/api/v1/reviews")
	v1.GET("", d.Handler.ListByProduct)
	v1.GET("/summary/:productID", d.Handler.GetSummary)

	protected := v1.Group("")
	protected.Use(middleware.Authentication(d.Verifier))
	protected.GET("/me", d.Handler.ListMine)
	protected.POST("", d.Handler.CreateReview)
	protected.PUT("/:id", d.Handler.UpdateReview)
	protected.DELETE("/:id", d.Handler.DeleteReview)
	return engine
}
