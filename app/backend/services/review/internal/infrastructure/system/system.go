package system

import (
	"time"

	"github.com/google/uuid"
)

type Clock struct{}

func NewClock() Clock { return Clock{} }
func (Clock) Now() time.Time { return time.Now().UTC() }

type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator { return UUIDGenerator{} }
func (UUIDGenerator) New() uuid.UUID { return uuid.New() }
