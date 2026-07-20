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

	v1 := engine.Group("/api/v1/cart")
	v1.Use(middleware.Authentication(d.Verifier))
	v1.GET("", d.Handler.GetActiveCart)
	v1.POST("/items", d.Handler.AddItem)
	v1.PUT("/items/:product_id", d.Handler.UpdateItem)
	v1.DELETE("/items", d.Handler.ClearCart)
	v1.POST("/checkout", d.Handler.Checkout)
	return engine
}
