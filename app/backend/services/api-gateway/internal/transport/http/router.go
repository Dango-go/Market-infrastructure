package http

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/embedded-market/backend/pkg/middleware"
)

type RouterDeps struct {
	System   *SystemHandler
	Verifier middleware.TokenVerifier
	Logger   *zap.Logger

	Auth           *url.URL
	User           *url.URL
	Catalog        *url.URL
	Inventory      *url.URL
	Pricing        *url.URL
	Cart           *url.URL
	Order          *url.URL
	Shipping       *url.URL
	Payment        *url.URL
	Notification   *url.URL
	Wishlist       *url.URL
	Search         *url.URL
	Review         *url.URL
	Recommendation *url.URL
	Analytics      *url.URL
}

func NewRouter(d RouterDeps) *gin.Engine {
	engine := gin.New()
	engine.Use(corsMiddleware())
	engine.Use(middleware.RequestContext(), middleware.HTTPMetrics("api-gateway"), middleware.Logging(d.Logger), middleware.Recovery())
	engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "route not found"}})
	})
	engine.GET("/healthz", d.System.Healthz)
	engine.GET("/readyz", d.System.Readyz)
	engine.GET("/metrics", middleware.PrometheusHandler("api-gateway"))

	authProxy := NewProxyHandler(d.Auth)
	userProxy := NewProxyHandler(d.User)
	catalogProxy := NewProxyHandler(d.Catalog)
	inventoryProxy := NewProxyHandler(d.Inventory)
	pricingProxy := NewProxyHandler(d.Pricing)
	cartProxy := NewProxyHandler(d.Cart)
	orderProxy := NewProxyHandler(d.Order)
	shippingProxy := NewProxyHandler(d.Shipping)
	paymentProxy := NewProxyHandler(d.Payment)
	notificationProxy := NewProxyHandler(d.Notification)
	wishlistProxy := NewProxyHandler(d.Wishlist)
	searchProxy := NewProxyHandler(d.Search)
	reviewProxy := NewProxyHandler(d.Review)
	recommendationProxy := NewProxyHandler(d.Recommendation)
	analyticsProxy := NewProxyHandler(d.Analytics)

	registerProxy(engine, "/.well-known/jwks.json", authProxy.Serve)
	registerProxy(engine, "/api/v1/auth", authProxy.Serve)
	registerProxy(engine, "/api/v1/users", userProxy.Serve)
	registerProxy(engine, "/api/v1/catalog", catalogProxy.Serve)
	registerProxy(engine, "/api/v1/inventory", inventoryProxy.Serve)
	registerProxy(engine, "/api/v1/pricing", pricingProxy.Serve)
	registerProxy(engine, "/api/v1/search", searchProxy.Serve)
	registerProxy(engine, "/api/v1/reviews", reviewProxy.Serve)
	registerProxy(engine, "/api/v1/recommendations", recommendationProxy.Serve)
	registerProxy(engine, "/api/v1/analytics", analyticsProxy.Serve)
	registerProtectedProxy(engine, "/api/v1/cart", d.Verifier, cartProxy.Serve)
	registerProtectedProxy(engine, "/api/v1/orders", d.Verifier, orderProxy.Serve)
	registerProtectedProxy(engine, "/api/v1/shipping", d.Verifier, shippingProxy.Serve)
	registerProtectedProxy(engine, "/api/v1/payments", d.Verifier, paymentProxy.Serve)
	registerProtectedProxy(engine, "/api/v1/notifications", d.Verifier, notificationProxy.Serve)
	registerProtectedProxy(engine, "/api/v1/wishlist", d.Verifier, wishlistProxy.Serve)
	registerProxy(engine, "/api/v1/notification-templates", notificationProxy.Serve)

	return engine
}

func registerProxy(engine *gin.Engine, basePath string, handler gin.HandlerFunc) {
	engine.Any(basePath, handler)
	engine.Any(basePath+"/*proxyPath", handler)
}

func registerProtectedProxy(engine *gin.Engine, basePath string, verifier middleware.TokenVerifier, handler gin.HandlerFunc) {
	group := engine.Group(basePath)
	group.Use(middleware.Authentication(verifier))
	group.Any("", handler)
	group.Any("/*proxyPath", handler)
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		}

		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		c.Next()
	}
}
