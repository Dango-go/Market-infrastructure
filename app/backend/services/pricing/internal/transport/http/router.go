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

	v1 := engine.Group("/api/v1/pricing")
	v1.GET("/prices", d.Handler.ListPrices)
	v1.GET("/prices/:product_id", d.Handler.GetPrice)
	v1.GET("/promotions", d.Handler.ListPromotions)

	write := v1.Group("")
	write.Use(middleware.Authentication(d.Verifier))
	write.POST("/prices", d.Handler.UpsertPrice)
	write.POST("/promotions", d.Handler.CreatePromotion)
	return engine
}
