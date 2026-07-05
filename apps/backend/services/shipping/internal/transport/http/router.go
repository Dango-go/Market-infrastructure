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

	v1 := engine.Group("/api/v1/shipping")
	v1.Use(middleware.Authentication(d.Verifier))
	v1.GET("/shipments", d.Handler.ListShipments)
	v1.GET("/shipments/:id", d.Handler.GetShipment)
	v1.POST("/shipments", d.Handler.CreateShipment)
	v1.PUT("/shipments/:id/status", d.Handler.UpdateStatus)
	return engine
}
