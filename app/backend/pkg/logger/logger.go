// Package logger provides a zap-based structured logger and helpers to carry
// request-id / correlation-id through context, satisfying the platform-wide logging
// requirement that every log line is attributable to a request and a correlation chain.
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey int

const (
	loggerKey ctxKey = iota
	requestIDKey
	correlationIDKey
)

// Field aliases so callers need not import zap directly.
type Field = zap.Field

var (
	String = zap.String
	Int    = zap.Int
	Int64  = zap.Int64
	Error  = zap.Error
	Any    = zap.Any
	Bool   = zap.Bool
)

// New builds a production or development zap logger for the given service name and level.
func New(service, level string, development bool) (*zap.Logger, error) {
	var cfg zap.Config
	if development {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	if lvl, err := zapcore.ParseLevel(level); err == nil {
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}
	l, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, err
	}
	return l.With(zap.String("service", service)), nil
}

// WithContext stores a logger in the context.
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns the request-scoped logger, or a no-op logger if none is set.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.NewNop()
}

// WithRequestID / WithCorrelationID store identifiers used by middleware and propagated
// into Kafka event envelopes.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func CorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDKey).(string); ok {
		return v
	}
	return ""
}
