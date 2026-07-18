package http

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/embedded-market/backend/pkg/middleware"
)

// RouterDeps are the dependencies needed to assemble the auth HTTP router.
type RouterDeps struct {
	Handler   *Handler
	System    *SystemHandler
	Verifier  middleware.TokenVerifier
	Logger    *zap.Logger
	Redis     *redis.Client
	RateLimit middleware.RateLimitConfig
}

// NewRouter builds the Gin engine with platform middleware and the auth routes.
func NewRouter(d RouterDeps) *gin.Engine {
	engine := gin.New()
	engine.Use(
		middleware.RequestContext(),
		middleware.HTTPMetrics("auth"),
		middleware.Logging(d.Logger),
		middleware.Recovery(),
	)

	// Operational endpoints (unauthenticated, not rate limited).
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("auth"))
	engine.GET("/.well-known/jwks.json", d.System.JWKS)

	v1 := engine.Group("/api/v1/auth")
	if d.Redis != nil {
		v1.Use(middleware.RateLimit(d.Redis, d.RateLimit))
	}

	// Public, credential-bearing endpoints.
	v1.POST("/register", d.Handler.Register)
	v1.POST("/login", d.Handler.Login)
	v1.POST("/refresh", d.Handler.Refresh)
	v1.POST("/logout", d.Handler.Logout)

	// OAuth authorization-code flow.
	v1.GET("/oauth/:provider", d.Handler.OAuthBegin)
	v1.GET("/oauth/:provider/callback", d.Handler.OAuthCallback)

	// Authenticated endpoints.
	authed := v1.Group("")
	authed.Use(middleware.Authentication(d.Verifier))
	authed.GET("/me", d.Handler.Me)
	authed.GET("/sessions", d.Handler.ListSessions)
	authed.DELETE("/sessions/:id", d.Handler.RevokeSession)
	authed.DELETE("/sessions", d.Handler.RevokeAllSessions)

	return engine
}
