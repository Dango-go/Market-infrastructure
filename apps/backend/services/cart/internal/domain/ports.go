package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type CartRepository interface {
	Create(ctx context.Context, cart *Cart) error
	GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) (*Cart, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Cart, error)
	Update(ctx context.Context, cart *Cart) error
}

type CartItemRepository interface {
	ListByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error)
	GetByCartProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error)
	Create(ctx context.Context, item *CartItem) error
	Update(ctx context.Context, item *CartItem) error
	Delete(ctx context.Context, itemID uuid.UUID) error
	DeleteByCartID(ctx context.Context, cartID uuid.UUID) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Carts() CartRepository
	Items() CartItemRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
