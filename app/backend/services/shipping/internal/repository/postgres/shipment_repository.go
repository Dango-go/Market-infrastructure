package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/shipping/internal/domain"
	"github.com/google/uuid"
)

type shipmentRepository struct{ db pgxConn }

func (r *shipmentRepository) Create(ctx context.Context, item *domain.Shipment) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO shipments (id, order_id, account_id, status, carrier, service_level, tracking_number, destination_address, eta, shipped_at, delivered_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, item.ID, item.OrderID, item.AccountID, string(item.Status), item.Carrier, item.ServiceLevel, item.TrackingNumber, item.DestinationAddress, item.Eta, item.ShippedAt, item.DeliveredAt, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("insert shipment: %w", err) }
	return nil
}

func (r *shipmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	return r.getOne(ctx, `SELECT id, order_id, account_id, status, carrier, service_level, tracking_number, destination_address, eta, shipped_at, delivered_at, created_at, updated_at FROM shipments WHERE id = $1`, id)
}

func (r *shipmentRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]domain.Shipment, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, order_id, account_id, status, carrier, service_level, tracking_number, destination_address, eta, shipped_at, delivered_at, created_at, updated_at, COUNT(*) OVER() AS total_count FROM shipments WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list shipments: %w", err) }
	defer rows.Close()
	items := make([]domain.Shipment, 0)
	var total int64
	for rows.Next() {
		var item domain.Shipment
		var status string
		if err := rows.Scan(&item.ID, &item.OrderID, &item.AccountID, &status, &item.Carrier, &item.ServiceLevel, &item.TrackingNumber, &item.DestinationAddress, &item.Eta, &item.ShippedAt, &item.DeliveredAt, &item.CreatedAt, &item.UpdatedAt, &total); err != nil { return nil, 0, fmt.Errorf("scan shipment: %w", err) }
		item.Status = domain.ShipmentStatus(status)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate shipments: %w", err) }
	return items, total, nil
}

func (r *shipmentRepository) Update(ctx context.Context, item *domain.Shipment) error {
	tag, err := r.db.Exec(ctx, `UPDATE shipments SET status = $2, carrier = $3, service_level = $4, tracking_number = $5, destination_address = $6, eta = $7, shipped_at = $8, delivered_at = $9, updated_at = $10 WHERE id = $1`, item.ID, string(item.Status), item.Carrier, item.ServiceLevel, item.TrackingNumber, item.DestinationAddress, item.Eta, item.ShippedAt, item.DeliveredAt, item.UpdatedAt)
	if err != nil { return fmt.Errorf("update shipment: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrShipmentNotFound }
	return nil
}

func (r *shipmentRepository) getOne(ctx context.Context, query string, arg any) (*domain.Shipment, error) {
	var item domain.Shipment
	var status string
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.OrderID, &item.AccountID, &status, &item.Carrier, &item.ServiceLevel, &item.TrackingNumber, &item.DestinationAddress, &item.Eta, &item.ShippedAt, &item.DeliveredAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrShipmentNotFound }
		return nil, fmt.Errorf("get shipment: %w", err)
	}
	item.Status = domain.ShipmentStatus(status)
	return &item, nil
}
