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

	pkgevents "github.com/embedded-market/backend/pkg/events"
	userdb "github.com/embedded-market/backend/services/user/db"
	"github.com/embedded-market/backend/services/user/internal/application"
	"github.com/embedded-market/backend/services/user/internal/config"
	infraevents "github.com/embedded-market/backend/services/user/internal/infrastructure/events"
	"github.com/embedded-market/backend/services/user/internal/infrastructure/system"
	userrepo "github.com/embedded-market/backend/services/user/internal/repository/postgres"
	transport "github.com/embedded-market/backend/services/user/internal/transport/http"
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

	if err := postgres.Migrate(startupCtx, pool, userdb.Migrations()); err != nil {
		return err
	}
	log.Info("database migrations applied")

	producer := pkgkafka.NewProducer(cfg.Kafka.Brokers)
	defer func() { _ = producer.Close() }()

	consumer := pkgkafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, pkgevents.TopicUserRegistered)
	defer func() { _ = consumer.Close() }()

	clock := system.NewClock()
	ids := system.NewUUIDGenerator()
	store := userrepo.NewStore(pool)

	deps := application.Deps{Store: store, Clock: clock, IDs: ids, Source: cfg.ServiceName}
	useCases := transport.UseCases{
		Bootstrap:   application.NewBootstrapUseCase(deps),
		Profile:     application.NewProfileUseCase(deps),
		Preferences: application.NewPreferencesUseCase(deps),
		Addresses:   application.NewAddressUseCase(deps),
	}

	verifier := pkgtoken.NewVerifier(cfg.JWT.PublicKey(), cfg.JWT.Issuer, cfg.JWT.Audience)
	handler := transport.NewHandler(useCases)
	systemHandler := transport.NewSystemHandler(func(ctx context.Context) error { return pool.Ping(ctx) })

	router := transport.NewRouter(transport.RouterDeps{
		Handler:  handler,
		System:   systemHandler,
		Verifier: verifier,
		Logger:   log,
	})

	relay := infraevents.NewRelay(store, producer, clock, log, cfg.Outbox.PollInterval, cfg.Outbox.BatchSize)
	go relay.Run(ctx)
	go runConsumerLoop(ctx, consumer, useCases.Bootstrap, log)

	log.Info("starting user service", zap.Int("port", cfg.HTTPPort), zap.String("env", cfg.Environment))
	return httpx.RunServer(ctx, router, httpx.ServerConfig{
		Port:            cfg.HTTPPort,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})
}

func runConsumerLoop(ctx context.Context, consumer *pkgkafka.Consumer, bootstrap *application.BootstrapUseCase, log *zap.Logger) {
	if bootstrap == nil {
		return
	}
	for ctx.Err() == nil {
		err := consumer.Run(ctx, func(ctx context.Context, env pkgevents.Envelope) error {
			return bootstrap.Handle(ctx, env)
		})
		if err == nil || ctx.Err() != nil {
			return
		}
		log.Warn("user registered consumer stopped; retrying", zap.Error(err))
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
