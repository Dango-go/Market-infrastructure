package system

import (
	"time"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/payment/internal/domain"
)

type Clock struct{}
func NewClock() Clock { return Clock{} }
var _ domain.Clock = Clock{}
func (Clock) Now() time.Time { return time.Now().UTC() }

type UUIDGenerator struct{}
func NewUUIDGenerator() UUIDGenerator { return UUIDGenerator{} }
var _ domain.IDGenerator = UUIDGenerator{}
func (UUIDGenerator) New() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil { return id }
	return uuid.New()
}
