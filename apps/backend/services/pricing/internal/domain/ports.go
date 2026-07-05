package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type PriceRepository interface {
	Create(ctx context.Context, p *Price) error
	Update(ctx context.Context, p *Price) error
	GetByProductID(ctx context.Context, productID uuid.UUID) (*Price, error)
	List(ctx context.Context, limit, offset int32) ([]*Price, int64, error)
}

type PromotionRepository interface {
	Create(ctx context.Context, p *Promotion) error
	List(ctx context.Context, limit, offset int32) ([]*Promotion, int64, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Prices() PriceRepository
	Promotions() PromotionRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
