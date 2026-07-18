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
	engine.Use(middleware.RequestContext(), middleware.HTTPMetrics("inventory"), middleware.Logging(d.Logger), middleware.Recovery())
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("inventory"))

	v1 := engine.Group("/api/v1/inventory")
	v1.GET("/warehouses", d.Handler.ListWarehouses)
	v1.GET("/stock", d.Handler.ListStock)
	v1.GET("/reservations", d.Handler.ListReservations)

	write := v1.Group("")
	write.Use(middleware.Authentication(d.Verifier))
	write.POST("/warehouses", d.Handler.CreateWarehouse)
	write.POST("/stock/adjust", d.Handler.AdjustStock)
	write.POST("/reservations", d.Handler.ReserveStock)
	write.POST("/reservations/:id/release", d.Handler.ReleaseReservation)
	return engine
}
