package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface { Now() time.Time }
type IDGenerator interface { New() uuid.UUID }

type StockFilters struct {
	ProductID   *uuid.UUID
	WarehouseID *uuid.UUID
}

type WarehouseRepository interface {
	Create(ctx context.Context, w *Warehouse) error
	List(ctx context.Context) ([]*Warehouse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Warehouse, error)
}

type StockRepository interface {
	CreateOrGet(ctx context.Context, item *StockItem) (*StockItem, error)
	GetByProductWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*StockItem, error)
	List(ctx context.Context, filters StockFilters, limit, offset int32) ([]*StockItem, int64, error)
	Update(ctx context.Context, item *StockItem) error
}

type ReservationRepository interface {
	Create(ctx context.Context, reservation *Reservation) error
	GetByID(ctx context.Context, id uuid.UUID) (*Reservation, error)
	Update(ctx context.Context, reservation *Reservation) error
	List(ctx context.Context, limit, offset int32) ([]*Reservation, int64, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Warehouses() WarehouseRepository
	Stock() StockRepository
	Reservations() ReservationRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
