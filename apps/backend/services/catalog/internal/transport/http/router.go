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

	v1 := engine.Group("/api/v1/catalog")
	v1.GET("/products", d.Handler.ListProducts)
	v1.GET("/products/:slug", d.Handler.GetProduct)
	v1.GET("/categories", d.Handler.ListCategories)
	v1.GET("/brands", d.Handler.ListBrands)

	write := v1.Group("")
	write.Use(middleware.Authentication(d.Verifier))
	write.POST("/products", d.Handler.CreateProduct)
	write.PATCH("/products/:id", d.Handler.UpdateProduct)
	write.POST("/categories", d.Handler.CreateCategory)
	write.POST("/brands", d.Handler.CreateBrand)
	return engine
}
