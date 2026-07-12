package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/logger"
	pkgtoken "github.com/embedded-market/backend/pkg/token"
	"github.com/embedded-market/backend/services/api-gateway/internal/config"
	transport "github.com/embedded-market/backend/services/api-gateway/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	log, err := logger.New(cfg.ServiceName, cfg.LogLevel, cfg.IsDevelopment())
	if err != nil {
		panic("failed to build logger: " + err.Error())
	}
	defer func() { _ = log.Sync() }()

	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, log); err != nil {
		log.Fatal("gateway terminated", zap.Error(err))
	}
	log.Info("gateway stopped cleanly")
}

func run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {
	verifier := pkgtoken.NewVerifier(cfg.JWT.PublicKey(), cfg.JWT.Issuer, cfg.JWT.Audience)
	router := transport.NewRouter(transport.RouterDeps{
		System:         transport.NewSystemHandler(),
		Verifier:       verifier,
		Logger:         log,
		Auth:           cfg.Upstream.Auth,
		User:           cfg.Upstream.User,
		Catalog:        cfg.Upstream.Catalog,
		Inventory:      cfg.Upstream.Inventory,
		Pricing:        cfg.Upstream.Pricing,
		Cart:           cfg.Upstream.Cart,
		Order:          cfg.Upstream.Order,
		Shipping:       cfg.Upstream.Shipping,
		Payment:        cfg.Upstream.Payment,
		Notification:   cfg.Upstream.Notification,
		Wishlist:       cfg.Upstream.Wishlist,
		Search:         cfg.Upstream.Search,
		Review:         cfg.Upstream.Review,
		Recommendation: cfg.Upstream.Recommendation,
		Analytics:      cfg.Upstream.Analytics,
	})

	log.Info("starting api gateway", zap.Int("port", cfg.HTTPPort), zap.String("env", cfg.Environment))
	return httpx.RunServer(ctx, router, httpx.ServerConfig{
		Port:            cfg.HTTPPort,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})
}
