package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type WishlistRepository interface {
	Add(ctx context.Context, item *WishlistItem) (bool, error)
	Remove(ctx context.Context, accountID, productID uuid.UUID) error
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]*WishlistItem, int64, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Wishlist() WishlistRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
