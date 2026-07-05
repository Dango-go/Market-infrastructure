// Command auth is the composition root of the auth service: it loads configuration, wires
// every layer's concrete dependencies behind their domain interfaces, applies database
// migrations, starts the transactional-outbox relay, and serves the HTTP API with
// graceful shutdown.
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
	"github.com/embedded-market/backend/pkg/middleware"
	"github.com/embedded-market/backend/pkg/postgres"
	pkgredis "github.com/embedded-market/backend/pkg/redis"
	pkgtoken "github.com/embedded-market/backend/pkg/token"

	authdb "github.com/embedded-market/backend/services/auth/db"
	"github.com/embedded-market/backend/services/auth/internal/application"
	"github.com/embedded-market/backend/services/auth/internal/config"
	"github.com/embedded-market/backend/services/auth/internal/infrastructure/crypto"
	infraevents "github.com/embedded-market/backend/services/auth/internal/infrastructure/events"
	infraoauth "github.com/embedded-market/backend/services/auth/internal/infrastructure/oauth"
	"github.com/embedded-market/backend/services/auth/internal/infrastructure/system"
	infratoken "github.com/embedded-market/backend/services/auth/internal/infrastructure/token"
	authrepo "github.com/embedded-market/backend/services/auth/internal/repository/postgres"
	transport "github.com/embedded-market/backend/services/auth/internal/transport/http"
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
	// --- Infrastructure: datastores & messaging ---
	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pool, err := postgres.Connect(startupCtx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := postgres.Migrate(startupCtx, pool, authdb.Migrations()); err != nil {
		return err
	}
	log.Info("database migrations applied")

	rdb := connectRedis(startupCtx, cfg, log)
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
	}

	producer := pkgkafka.NewProducer(cfg.Kafka.Brokers)
	defer func() { _ = producer.Close() }()

	// --- Infrastructure: ports implementations ---
	clock := system.NewClock()
	ids := system.NewUUIDGenerator()
	hasher := crypto.NewArgon2Hasher(crypto.DefaultArgon2Params())

	tokenSvc := infratoken.NewService(infratoken.Config{
		PrivateKey: cfg.JWT.PrivateKey(),
		KeyID:      cfg.JWT.KeyID,
		Issuer:     cfg.JWT.Issuer,
		Audience:   cfg.JWT.Audience,
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
	})

	oauthGateway := infraoauth.NewGateway(infraoauth.Config{
		GitHub: infraoauth.ProviderCredentials{ClientID: cfg.OAuth.GitHubClientID, ClientSecret: cfg.OAuth.GitHubClientSecret, RedirectURL: cfg.OAuth.GitHubRedirectURL},
		Google: infraoauth.ProviderCredentials{ClientID: cfg.OAuth.GoogleClientID, ClientSecret: cfg.OAuth.GoogleClientSecret, RedirectURL: cfg.OAuth.GoogleRedirectURL},
		GitLab: infraoauth.ProviderCredentials{ClientID: cfg.OAuth.GitLabClientID, ClientSecret: cfg.OAuth.GitLabClientSecret, RedirectURL: cfg.OAuth.GitLabRedirectURL},
	})

	store := authrepo.NewStore(pool)

	// --- Application: use cases ---
	deps := application.Deps{
		Store:  store,
		Hasher: hasher,
		Tokens: tokenSvc,
		OAuth:  oauthGateway,
		Clock:  clock,
		IDs:    ids,
		Source: cfg.ServiceName,
	}
	useCases := transport.UseCases{
		Register: application.NewRegisterUseCase(deps),
		Login:    application.NewLoginUseCase(deps),
		Refresh:  application.NewRefreshUseCase(deps),
		Logout:   application.NewLogoutUseCase(deps),
		Sessions: application.NewSessionsUseCase(deps),
		Account:  application.NewAccountUseCase(deps),
		OAuth:    application.NewOAuthUseCase(deps),
	}

	// --- Transport: verifier, handlers, router ---
	verifier := pkgtoken.NewVerifier(&cfg.JWT.PrivateKey().PublicKey, cfg.JWT.Issuer, cfg.JWT.Audience)

	handler := transport.NewHandler(useCases, transport.HandlerConfig{
		RefreshCookieName: cfg.Cookie.RefreshName,
		OAuthStateCookie:  cfg.Cookie.StateName,
		CookieDomain:      cfg.Cookie.Domain,
		CookieSecure:      cfg.Cookie.Secure,
		RefreshTTL:        cfg.JWT.RefreshTTL,
		OAuthStateTTL:     cfg.Cookie.OAuthStateTTL,
	})

	systemHandler := transport.NewSystemHandler(
		func() any { return tokenSvc.JWKS() },
		func(ctx context.Context) error { return pool.Ping(ctx) },
	)

	router := transport.NewRouter(transport.RouterDeps{
		Handler:  handler,
		System:   systemHandler,
		Verifier: verifier,
		Logger:   log,
		Redis:    rdb,
		RateLimit: middleware.RateLimitConfig{
			Requests: cfg.RateLimitRequests,
			Window:   cfg.RateLimitWindow,
			Prefix:   "auth",
		},
	})

	// --- Background: outbox relay ---
	relay := infraevents.NewRelay(store, producer, clock, log, cfg.Outbox.PollInterval, cfg.Outbox.BatchSize)
	go relay.Run(ctx)

	log.Info("starting auth service", zap.Int("port", cfg.HTTPPort), zap.String("env", cfg.Environment))
	return httpx.RunServer(ctx, router, httpx.ServerConfig{
		Port:            cfg.HTTPPort,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})
}

// connectRedis attempts a Redis connection for rate limiting. Redis is optional for the
// auth service: on failure we log and continue with rate limiting disabled rather than
// coupling auth availability to a cache.
func connectRedis(ctx context.Context, cfg *config.Config, log *zap.Logger) *pkgredis.Client {
	if cfg.Redis.Addr == "" {
		return nil
	}
	rdb, err := pkgredis.Connect(ctx, cfg.Redis)
	if err != nil {
		log.Warn("redis unavailable; rate limiting disabled", zap.Error(err))
		return nil
	}
	return rdb
}
