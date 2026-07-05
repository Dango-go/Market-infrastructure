package middleware

import (
	"time"

	"github.com/embedded-market/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestContext assigns/propagates request-id and correlation-id, echoes them as
// response headers, and stores them on both the Gin context and the request context so
// the logger and Kafka envelopes can reference them.
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		correlationID := c.GetHeader(HeaderCorrelationID)
		if correlationID == "" {
			correlationID = requestID
		}

		c.Set(ctxRequestID, requestID)
		c.Set(ctxCorrelationID, correlationID)
		c.Header(HeaderRequestID, requestID)
		c.Header(HeaderCorrelationID, correlationID)

		ctx := c.Request.Context()
		ctx = logger.WithRequestID(ctx, requestID)
		ctx = logger.WithCorrelationID(ctx, correlationID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Logging binds a request-scoped logger (tagged with ids) into the request context and
// emits one structured access log per request.
func Logging(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := RequestID(c)
		corrID, _ := c.Get(ctxCorrelationID)

		l := base.With(
			zap.String("request_id", reqID),
			zap.Any("correlation_id", corrID),
		)
		c.Request = c.Request.WithContext(logger.WithContext(c.Request.Context(), l))

		c.Next()

		l.Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// Recovery converts panics into a 500 and logs the stack, keeping the process alive.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.FromContext(c.Request.Context()).Error("panic recovered", zap.Any("panic", r))
				c.AbortWithStatusJSON(500, gin.H{"error": gin.H{
					"code":    "internal_error",
					"message": "an internal error occurred",
				}})
			}
		}()
		c.Next()
	}
}
