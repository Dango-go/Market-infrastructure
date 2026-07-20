package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/payment/internal/domain"
	"github.com/google/uuid"
)

type paymentRepository struct{ db pgxConn }

func (r *paymentRepository) Create(ctx context.Context, item *domain.Payment) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO payments (id, order_id, account_id, status, provider, method, currency, amount_cents, transaction_ref, failure_reason, paid_at, refunded_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, item.ID, item.OrderID, item.AccountID, string(item.Status), item.Provider, item.Method, item.Currency, item.AmountCents, item.TransactionRef, item.FailureReason, item.PaidAt, item.RefundedAt, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("insert payment: %w", err) }
	return nil
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return r.getOne(ctx, `SELECT id, order_id, account_id, status, provider, method, currency, amount_cents, transaction_ref, failure_reason, paid_at, refunded_at, created_at, updated_at FROM payments WHERE id = $1`, id)
}

func (r *paymentRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]domain.Payment, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, order_id, account_id, status, provider, method, currency, amount_cents, transaction_ref, failure_reason, paid_at, refunded_at, created_at, updated_at, COUNT(*) OVER() AS total_count FROM payments WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list payments: %w", err) }
	defer rows.Close()
	items := make([]domain.Payment, 0)
	var total int64
	for rows.Next() {
		var item domain.Payment
		var status string
		if err := rows.Scan(&item.ID, &item.OrderID, &item.AccountID, &status, &item.Provider, &item.Method, &item.Currency, &item.AmountCents, &item.TransactionRef, &item.FailureReason, &item.PaidAt, &item.RefundedAt, &item.CreatedAt, &item.UpdatedAt, &total); err != nil { return nil, 0, fmt.Errorf("scan payment: %w", err) }
		item.Status = domain.PaymentStatus(status)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate payments: %w", err) }
	return items, total, nil
}

func (r *paymentRepository) Update(ctx context.Context, item *domain.Payment) error {
	tag, err := r.db.Exec(ctx, `UPDATE payments SET status = $2, provider = $3, method = $4, currency = $5, amount_cents = $6, transaction_ref = $7, failure_reason = $8, paid_at = $9, refunded_at = $10, updated_at = $11 WHERE id = $1`, item.ID, string(item.Status), item.Provider, item.Method, item.Currency, item.AmountCents, item.TransactionRef, item.FailureReason, item.PaidAt, item.RefundedAt, item.UpdatedAt)
	if err != nil { return fmt.Errorf("update payment: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrPaymentNotFound }
	return nil
}

func (r *paymentRepository) getOne(ctx context.Context, query string, arg any) (*domain.Payment, error) {
	var item domain.Payment
	var status string
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.OrderID, &item.AccountID, &status, &item.Provider, &item.Method, &item.Currency, &item.AmountCents, &item.TransactionRef, &item.FailureReason, &item.PaidAt, &item.RefundedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrPaymentNotFound }
		return nil, fmt.Errorf("get payment: %w", err)
	}
	item.Status = domain.PaymentStatus(status)
	return &item, nil
}
