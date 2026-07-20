package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Clock interface { Now() time.Time }

type IDGenerator interface { New() uuid.UUID }

type ProductProfileRepository interface {
	Upsert(ctx context.Context, item *ProductProfile) error
	GetByProductID(ctx context.Context, productID uuid.UUID) (*ProductProfile, error)
	Related(ctx context.Context, productID uuid.UUID, limit int32) ([]Recommendation, error)
	Trending(ctx context.Context, limit int32) ([]Recommendation, error)
}

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
}

type Store interface {
	Profiles() ProductProfileRepository
	Events() EventRepository
}
