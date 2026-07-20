package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/inventory/internal/domain"
	"github.com/google/uuid"
)

type reservationRepository struct{ db pgxConn }

func (r *reservationRepository) Create(ctx context.Context, reservation *domain.Reservation) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO reservations (id, product_id, warehouse_id, reference, quantity, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, reservation.ID, reservation.ProductID, reservation.WarehouseID, reservation.Reference, reservation.Quantity, string(reservation.Status), reservation.CreatedAt, reservation.UpdatedAt); err != nil { return fmt.Errorf("insert reservation: %w", err) }
	return nil
}

func (r *reservationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Reservation, error) {
	var item domain.Reservation
	var status string
	if err := r.db.QueryRow(ctx, `SELECT id, product_id, warehouse_id, reference, quantity, status, created_at, updated_at FROM reservations WHERE id = $1`, id).Scan(&item.ID, &item.ProductID, &item.WarehouseID, &item.Reference, &item.Quantity, &status, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrReservationNotFound }
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	item.Status = domain.ReservationStatus(status)
	return &item, nil
}

func (r *reservationRepository) Update(ctx context.Context, reservation *domain.Reservation) error {
	tag, err := r.db.Exec(ctx, `UPDATE reservations SET status = $2, updated_at = $3 WHERE id = $1`, reservation.ID, string(reservation.Status), reservation.UpdatedAt)
	if err != nil { return fmt.Errorf("update reservation: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrReservationNotFound }
	return nil
}

func (r *reservationRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Reservation, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reservations`).Scan(&total); err != nil { return nil, 0, fmt.Errorf("count reservations: %w", err) }
	rows, err := r.db.Query(ctx, `SELECT id, product_id, warehouse_id, reference, quantity, status, created_at, updated_at FROM reservations ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list reservations: %w", err) }
	defer rows.Close()
	items := make([]*domain.Reservation, 0)
	for rows.Next() {
		var item domain.Reservation
		var status string
		if err := rows.Scan(&item.ID, &item.ProductID, &item.WarehouseID, &item.Reference, &item.Quantity, &status, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, 0, fmt.Errorf("scan reservation: %w", err) }
		item.Status = domain.ReservationStatus(status)
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate reservations: %w", err) }
	return items, total, nil
}
