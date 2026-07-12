package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	New() uuid.UUID
}

type ProfileRepository interface {
	CreateIfMissing(ctx context.Context, p *Profile) (bool, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*Profile, error)
	Update(ctx context.Context, p *Profile) error
}

type AddressRepository interface {
	Create(ctx context.Context, a *Address) error
	GetByID(ctx context.Context, id uuid.UUID) (*Address, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]*Address, int64, error)
	Update(ctx context.Context, a *Address) error
	Delete(ctx context.Context, id uuid.UUID, deletedAt time.Time) error
	ClearDefaultShipping(ctx context.Context, accountID uuid.UUID) error
	ClearDefaultBilling(ctx context.Context, accountID uuid.UUID) error
}

type PreferencesRepository interface {
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*Preferences, error)
	CreateIfMissing(ctx context.Context, p *Preferences) (bool, error)
	Update(ctx context.Context, p *Preferences) error
}

type ProcessedEventRepository interface {
	Mark(ctx context.Context, eventID uuid.UUID, topic string, processedAt time.Time) (bool, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Profiles() ProfileRepository
	Addresses() AddressRepository
	Preferences() PreferencesRepository
	ProcessedEvents() ProcessedEventRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
