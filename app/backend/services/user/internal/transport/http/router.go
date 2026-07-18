package http

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/embedded-market/backend/pkg/middleware"
)

type RouterDeps struct {
	Handler  *Handler
	System   *SystemHandler
	Verifier middleware.TokenVerifier
	Logger   *zap.Logger
	Redis    *redis.Client
}

func NewRouter(d RouterDeps) *gin.Engine {
	engine := gin.New()
	engine.Use(
		middleware.RequestContext(),
		middleware.HTTPMetrics("user"),
		middleware.Logging(d.Logger),
		middleware.Recovery(),
	)

	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("user"))

	v1 := engine.Group("/api/v1/users")
	authed := v1.Group("/me")
	authed.Use(middleware.Authentication(d.Verifier))
	{
		authed.GET("", d.Handler.Me)
		authed.PUT("", d.Handler.UpdateMe)

		authed.GET("/preferences", d.Handler.GetPreferences)
		authed.PUT("/preferences", d.Handler.UpdatePreferences)

		authed.GET("/addresses", d.Handler.ListAddresses)
		authed.POST("/addresses", d.Handler.CreateAddress)
		authed.PATCH("/addresses/:id", d.Handler.UpdateAddress)
		authed.DELETE("/addresses/:id", d.Handler.DeleteAddress)
		authed.PUT("/addresses/:id/default-shipping", d.Handler.SetDefaultShipping)
		authed.PUT("/addresses/:id/default-billing", d.Handler.SetDefaultBilling)
	}

	_ = d.Redis
	return engine
}
