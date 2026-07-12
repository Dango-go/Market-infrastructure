// Package system provides the trivial infrastructure adapters for time and identifier
// generation, satisfying the domain Clock and IDGenerator ports.
package system

import (
	"time"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// Clock is the real wall-clock implementation of domain.Clock.
type Clock struct{}

func NewClock() Clock { return Clock{} }

var _ domain.Clock = Clock{}

// Now returns the current UTC time.
func (Clock) Now() time.Time { return time.Now().UTC() }

// UUIDGenerator produces time-ordered UUIDv7 identifiers, which keep primary-key inserts
// index-friendly. It falls back to UUIDv4 if v7 generation ever fails.
type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator { return UUIDGenerator{} }

var _ domain.IDGenerator = UUIDGenerator{}

// New returns a fresh identifier.
func (UUIDGenerator) New() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}
