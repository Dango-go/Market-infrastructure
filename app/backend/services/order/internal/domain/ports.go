package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]Order, int64, error)
	Update(ctx context.Context, order *Order) error
}

type OrderItemRepository interface {
	ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)
	Create(ctx context.Context, item *OrderItem) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Orders() OrderRepository
	Items() OrderItemRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
