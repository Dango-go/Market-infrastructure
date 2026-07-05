// Package config provides shared configuration primitives parsed from the environment
// (12-factor). Each service embeds Base and adds its own fields, parsing the combined
// struct with Load.
package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Base holds configuration common to every service.
type Base struct {
	ServiceName string `env:"SERVICE_NAME"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	HTTPPort    int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
}

// IsDevelopment reports whether the service runs in a development environment.
func (b Base) IsDevelopment() bool { return b.Environment == "development" }

// Postgres holds connection settings for a PostgreSQL database.
type Postgres struct {
	DSN             string        `env:"POSTGRES_DSN,required"`
	MaxConns        int32         `env:"POSTGRES_MAX_CONNS" envDefault:"10"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS" envDefault:"2"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"1h"`
	ConnectTimeout  time.Duration `env:"POSTGRES_CONNECT_TIMEOUT" envDefault:"5s"`
}

// Redis holds connection settings for a Redis instance.
type Redis struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

// Kafka holds broker connection and consumer-group settings.
type Kafka struct {
	Brokers       []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	ConsumerGroup string   `env:"KAFKA_CONSUMER_GROUP"`
}

// Load parses environment variables into the provided config struct pointer.
func Load[T any](cfg *T) error {
	return env.Parse(cfg)
}
