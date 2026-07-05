package system

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/user/internal/domain"
)

type Clock struct{}

func NewClock() Clock { return Clock{} }

var _ domain.Clock = Clock{}

func (Clock) Now() time.Time { return time.Now().UTC() }

type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator { return UUIDGenerator{} }

var _ domain.IDGenerator = UUIDGenerator{}

func (UUIDGenerator) New() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}

// ReadyProbe is a thin adapter used by the HTTP system handler.
type ReadyProbe struct {
	Ping func(ctx context.Context) error
}
