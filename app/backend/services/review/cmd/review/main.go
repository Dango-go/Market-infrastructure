package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/logger"
	"github.com/embedded-market/backend/pkg/postgres"
	pkgtoken "github.com/embedded-market/backend/pkg/token"

	reviewdb "github.com/embedded-market/backend/services/review/db"
	"github.com/embedded-market/backend/services/review/internal/application"
	"github.com/embedded-market/backend/services/review/internal/config"
	"github.com/embedded-market/backend/services/review/internal/infrastructure/system"
	reviewrepo "github.com/embedded-market/backend/services/review/internal/repository/postgres"
	transport "github.com/embedded-market/backend/services/review/internal/transport/http"
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
		log.Fatal("service terminated", zap.Error(err))
	}
	log.Info("service stopped cleanly")
}

func run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {
	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pool, err := postgres.Connect(startupCtx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := postgres.Migrate(startupCtx, pool, reviewdb.Migrations()); err != nil {
		return err
	}
	log.Info("database migrations applied")

	clock := system.NewClock()
	ids := system.NewUUIDGenerator()
	store := reviewrepo.NewStore(pool)
	deps := application.Deps{Store: store, Clock: clock, IDs: ids}

	useCases := transport.UseCases{Review: application.NewReviewUseCase(deps)}
	verifier := pkgtoken.NewVerifier(cfg.JWT.PublicKey(), cfg.JWT.Issuer, cfg.JWT.Audience)
	handler := transport.NewHandler(useCases)
	systemHandler := transport.NewSystemHandler(func(ctx context.Context) error { return pool.Ping(ctx) })
	router := transport.NewRouter(transport.RouterDeps{Handler: handler, System: systemHandler, Verifier: verifier, Logger: log})

	log.Info("starting review service", zap.Int("port", cfg.HTTPPort), zap.String("env", cfg.Environment))
	return httpx.RunServer(ctx, router, httpx.ServerConfig{
		Port:            cfg.HTTPPort,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})
}
