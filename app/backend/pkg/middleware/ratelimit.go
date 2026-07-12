package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig configures a fixed-window Redis-backed limiter.
type RateLimitConfig struct {
	Requests int           // permitted requests per window
	Window   time.Duration // window length
	Prefix   string        // redis key namespace, e.g. "auth"
}

// RateLimit enforces a per-identity fixed-window limit using Redis INCR + EXPIRE. The
// identity is the authenticated account id when present, else the client IP. If Redis is
// unavailable the request is allowed (fail-open) to avoid coupling availability to it.
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.Requests <= 0 || cfg.Window <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		identity := AccountID(c)
		if identity == "" {
			identity = c.ClientIP()
		}
		key := "ratelimit:" + cfg.Prefix + ":" + identity

		count, err := incrementWindow(c.Request.Context(), rdb, key, cfg.Window)
		if err != nil {
			c.Next()
			return
		}
		remaining := cfg.Requests - int(count)
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(remaining, 0)))

		if count > int64(cfg.Requests) {
			httpx.Fail(c, apperr.RateLimited("rate_limited", "too many requests, slow down"))
			return
		}
		c.Next()
	}
}

func incrementWindow(ctx context.Context, rdb *redis.Client, key string, window time.Duration) (int64, error) {
	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Only stamp the TTL on the first request of a window so the window is fixed,
	// not perpetually extended by subsequent calls.
	if count == 1 {
		if err := rdb.Expire(ctx, key, window).Err(); err != nil {
			return count, err
		}
	}
	return count, nil
}
