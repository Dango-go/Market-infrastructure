// Package redis provides a go-redis client factory configured from shared config.
package redis

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/config"
	"github.com/redis/go-redis/v9"
)

// Client is the concrete Redis client type used across services.
type Client = redis.Client

// Connect builds and verifies a Redis client.
func Connect(ctx context.Context, cfg config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
