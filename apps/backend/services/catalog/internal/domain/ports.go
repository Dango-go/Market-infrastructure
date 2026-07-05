package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface { Now() time.Time }

type IDGenerator interface { New() uuid.UUID }

type ProductFilters struct {
	Query        string
	CategorySlug string
	BrandSlug    string
	Featured     *bool
	Status       string
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetBySlug(ctx context.Context, slug string) (*Product, error)
	List(ctx context.Context, filters ProductFilters, limit, offset int32) ([]*Product, int64, error)
	Update(ctx context.Context, p *Product) error
	ReplaceSpecs(ctx context.Context, productID uuid.UUID, specs []ProductSpec) error
	ReplaceMedia(ctx context.Context, productID uuid.UUID, media []ProductMedia) error
	ReplaceCompatibility(ctx context.Context, productID uuid.UUID, items []CompatibilityRule) error
}

type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	List(ctx context.Context) ([]*Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
}

type BrandRepository interface {
	Create(ctx context.Context, b *Brand) error
	List(ctx context.Context) ([]*Brand, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Brand, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Products() ProductRepository
	Categories() CategoryRepository
	Brands() BrandRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
