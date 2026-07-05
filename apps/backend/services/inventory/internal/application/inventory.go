package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/inventory/internal/domain"
	"github.com/google/uuid"
)

type InventoryUseCase struct{ Deps }

func NewInventoryUseCase(d Deps) *InventoryUseCase { return &InventoryUseCase{Deps: d} }

type CreateWarehouseInput struct {
	Code     string
	Name     string
	Location string
}

type AdjustStockInput struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Delta       int64
}

type ReserveStockInput struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Reference   string
	Quantity    int64
}

func (uc *InventoryUseCase) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (WarehouseView, error) {
	item := domain.NewWarehouse(uc.IDs.New(), input.Code, input.Name, input.Location, uc.Clock.Now())
	if err := uc.Store.Warehouses().Create(ctx, item); err != nil { return WarehouseView{}, err }
	return toWarehouseView(item), nil
}

func (uc *InventoryUseCase) ListWarehouses(ctx context.Context) ([]WarehouseView, error) {
	items, err := uc.Store.Warehouses().List(ctx)
	if err != nil { return nil, err }
	out := make([]WarehouseView, 0, len(items))
	for _, item := range items { out = append(out, toWarehouseView(item)) }
	return out, nil
}

func (uc *InventoryUseCase) ListStock(ctx context.Context, filters domain.StockFilters, limit, offset int32) ([]StockView, int64, error) {
	items, total, err := uc.Store.Stock().List(ctx, filters, limit, offset)
	if err != nil { return nil, 0, err }
	out := make([]StockView, 0, len(items))
	for _, item := range items { out = append(out, toStockView(item)) }
	return out, total, nil
}

func (uc *InventoryUseCase) AdjustStock(ctx context.Context, input AdjustStockInput, req RequestContext) (StockView, error) {
	var out StockView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if _, err := tx.Warehouses().GetByID(ctx, input.WarehouseID); err != nil { return err }
		stock, err := tx.Stock().CreateOrGet(ctx, domain.NewStockItem(uc.IDs.New(), input.ProductID, input.WarehouseID, 0, uc.Clock.Now()))
		if err != nil { return err }
		stock.Adjust(input.Delta, uc.Clock.Now())
		if err := tx.Stock().Update(ctx, stock); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicStockAdjusted, uc.Source, stock.ID.String(), req.CorrelationID, uc.Clock.Now(), events.StockAdjusted{ProductID: stock.ProductID.String(), WarehouseID: stock.WarehouseID.String(), OnHand: stock.OnHand, Reserved: stock.Reserved, Available: stock.Available(), Delta: input.Delta})
		if err != nil { return fmt.Errorf("build stock adjusted event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toStockView(stock)
		return nil
	})
	return out, err
}

func (uc *InventoryUseCase) ReserveStock(ctx context.Context, input ReserveStockInput, req RequestContext) (ReservationView, error) {
	var out ReservationView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if _, err := tx.Warehouses().GetByID(ctx, input.WarehouseID); err != nil { return err }
		stock, err := tx.Stock().CreateOrGet(ctx, domain.NewStockItem(uc.IDs.New(), input.ProductID, input.WarehouseID, 0, uc.Clock.Now()))
		if err != nil { return err }
		if err := stock.Reserve(input.Quantity, uc.Clock.Now()); err != nil { return err }
		if err := tx.Stock().Update(ctx, stock); err != nil { return err }
		reservation := domain.NewReservation(uc.IDs.New(), input.ProductID, input.WarehouseID, input.Reference, input.Quantity, uc.Clock.Now())
		if err := tx.Reservations().Create(ctx, reservation); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicStockReserved, uc.Source, reservation.ID.String(), req.CorrelationID, uc.Clock.Now(), events.StockReserved{ReservationID: reservation.ID.String(), ProductID: reservation.ProductID.String(), WarehouseID: reservation.WarehouseID.String(), Reference: reservation.Reference, Quantity: reservation.Quantity})
		if err != nil { return fmt.Errorf("build stock reserved event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toReservationView(reservation)
		return nil
	})
	return out, err
}

func (uc *InventoryUseCase) ReleaseReservation(ctx context.Context, id uuid.UUID, req RequestContext) (ReservationView, error) {
	var out ReservationView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		reservation, err := tx.Reservations().GetByID(ctx, id)
		if err != nil { return err }
		if reservation.Status == domain.ReservationReleased {
			out = toReservationView(reservation)
			return nil
		}
		stock, err := tx.Stock().GetByProductWarehouse(ctx, reservation.ProductID, reservation.WarehouseID)
		if err != nil { return err }
		stock.Release(reservation.Quantity, uc.Clock.Now())
		if err := tx.Stock().Update(ctx, stock); err != nil { return err }
		reservation.Release(uc.Clock.Now())
		if err := tx.Reservations().Update(ctx, reservation); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicStockReleased, uc.Source, reservation.ID.String(), req.CorrelationID, uc.Clock.Now(), events.StockReleased{ReservationID: reservation.ID.String(), ProductID: reservation.ProductID.String(), WarehouseID: reservation.WarehouseID.String(), Reference: reservation.Reference, Quantity: reservation.Quantity})
		if err != nil { return fmt.Errorf("build stock released event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toReservationView(reservation)
		return nil
	})
	return out, err
}

func (uc *InventoryUseCase) ListReservations(ctx context.Context, limit, offset int32) ([]ReservationView, int64, error) {
	items, total, err := uc.Store.Reservations().List(ctx, limit, offset)
	if err != nil { return nil, 0, err }
	out := make([]ReservationView, 0, len(items))
	for _, item := range items { out = append(out, toReservationView(item)) }
	return out, total, nil
}

func toWarehouseView(w *domain.Warehouse) WarehouseView {
	return WarehouseView{ID: w.ID, Code: w.Code, Name: w.Name, Location: w.Location, Active: w.Active, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt}
}

func toStockView(s *domain.StockItem) StockView {
	return StockView{ID: s.ID, ProductID: s.ProductID, WarehouseID: s.WarehouseID, OnHand: s.OnHand, Reserved: s.Reserved, Available: s.Available(), CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func toReservationView(r *domain.Reservation) ReservationView {
	return ReservationView{ID: r.ID, ProductID: r.ProductID, WarehouseID: r.WarehouseID, Reference: r.Reference, Quantity: r.Quantity, Status: string(r.Status), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
