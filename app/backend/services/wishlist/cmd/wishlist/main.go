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
	pkgkafka "github.com/embedded-market/backend/pkg/kafka"
	"github.com/embedded-market/backend/pkg/logger"
	"github.com/embedded-market/backend/pkg/postgres"
	pkgtoken "github.com/embedded-market/backend/pkg/token"

	wishlistdb "github.com/embedded-market/backend/services/wishlist/db"
	"github.com/embedded-market/backend/services/wishlist/internal/application"
	"github.com/embedded-market/backend/services/wishlist/internal/config"
	infraevents "github.com/embedded-market/backend/services/wishlist/internal/infrastructure/events"
	"github.com/embedded-market/backend/services/wishlist/internal/infrastructure/system"
	wishlistrepo "github.com/embedded-market/backend/services/wishlist/internal/repository/postgres"
	transport "github.com/embedded-market/backend/services/wishlist/internal/transport/http"
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

	if err := postgres.Migrate(startupCtx, pool, wishlistdb.Migrations()); err != nil {
		return err
	}
	log.Info("database migrations applied")

	producer := pkgkafka.NewProducer(cfg.Kafka.Brokers)
	defer func() { _ = producer.Close() }()

	clock := system.NewClock()
	ids := system.NewUUIDGenerator()
	store := wishlistrepo.NewStore(pool)
	deps := application.Deps{Store: store, Clock: clock, IDs: ids, Source: cfg.ServiceName}

	useCases := transport.UseCases{Wishlist: application.NewWishlistUseCase(deps)}
	verifier := pkgtoken.NewVerifier(cfg.JWT.PublicKey(), cfg.JWT.Issuer, cfg.JWT.Audience)
	handler := transport.NewHandler(useCases)
	systemHandler := transport.NewSystemHandler(func(ctx context.Context) error { return pool.Ping(ctx) })
	router := transport.NewRouter(transport.RouterDeps{Handler: handler, System: systemHandler, Verifier: verifier, Logger: log})

	relay := infraevents.NewRelay(store, producer, clock, log, cfg.Outbox.PollInterval, cfg.Outbox.BatchSize)
	go relay.Run(ctx)

	log.Info("starting wishlist service", zap.Int("port", cfg.HTTPPort), zap.String("env", cfg.Environment))
	return httpx.RunServer(ctx, router, httpx.ServerConfig{
		Port:            cfg.HTTPPort,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})
}
