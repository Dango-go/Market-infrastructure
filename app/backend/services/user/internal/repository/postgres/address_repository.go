package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type addressRepository struct{ db pgxConn }

func (r *addressRepository) Create(ctx context.Context, a *domain.Address) error {
	const q = `INSERT INTO addresses (id, account_id, label, recipient_name, line1, line2, city, region, postal_code, country_code, phone, is_default_shipping, is_default_billing, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	_, err := r.db.Exec(ctx, q, a.ID, a.AccountID, a.Label, a.RecipientName, a.Line1, a.Line2, a.City, a.Region, a.PostalCode, a.CountryCode, a.Phone, a.IsDefaultShipping, a.IsDefaultBilling, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert address: %w", err)
	}
	return nil
}

func (r *addressRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Address, error) {
	const q = `SELECT id, account_id, label, recipient_name, line1, line2, city, region, postal_code, country_code, phone, is_default_shipping, is_default_billing, created_at, updated_at, deleted_at FROM addresses WHERE id = $1 AND deleted_at IS NULL`
	row := r.db.QueryRow(ctx, q, id)
	return scanAddress(row)
}

func (r *addressRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]*domain.Address, int64, error) {
	const countQ = `SELECT COUNT(*) FROM addresses WHERE account_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, accountID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count addresses: %w", err)
	}
	const q = `SELECT id, account_id, label, recipient_name, line1, line2, city, region, postal_code, country_code, phone, is_default_shipping, is_default_billing, created_at, updated_at, deleted_at FROM addresses WHERE account_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, accountID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list addresses: %w", err)
	}
	defer rows.Close()
	items := make([]*domain.Address, 0)
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate addresses: %w", err)
	}
	return items, total, nil
}

func (r *addressRepository) Update(ctx context.Context, a *domain.Address) error {
	const q = `UPDATE addresses SET label = $2, recipient_name = $3, line1 = $4, line2 = $5, city = $6, region = $7, postal_code = $8, country_code = $9, phone = $10, is_default_shipping = $11, is_default_billing = $12, updated_at = $13, deleted_at = $14 WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, a.ID, a.Label, a.RecipientName, a.Line1, a.Line2, a.City, a.Region, a.PostalCode, a.CountryCode, a.Phone, a.IsDefaultShipping, a.IsDefaultBilling, a.UpdatedAt, a.DeletedAt)
	if err != nil {
		return fmt.Errorf("update address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepository) Delete(ctx context.Context, id uuid.UUID, deletedAt time.Time) error {
	const q = `UPDATE addresses SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id, deletedAt)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepository) ClearDefaultShipping(ctx context.Context, accountID uuid.UUID) error {
	const q = `UPDATE addresses SET is_default_shipping = FALSE, updated_at = now() WHERE account_id = $1 AND deleted_at IS NULL AND is_default_shipping = TRUE`
	if _, err := r.db.Exec(ctx, q, accountID); err != nil {
		return fmt.Errorf("clear default shipping: %w", err)
	}
	return nil
}

func (r *addressRepository) ClearDefaultBilling(ctx context.Context, accountID uuid.UUID) error {
	const q = `UPDATE addresses SET is_default_billing = FALSE, updated_at = now() WHERE account_id = $1 AND deleted_at IS NULL AND is_default_billing = TRUE`
	if _, err := r.db.Exec(ctx, q, accountID); err != nil {
		return fmt.Errorf("clear default billing: %w", err)
	}
	return nil
}

func scanAddress(row interface{ Scan(dest ...any) error }) (*domain.Address, error) {
	var a domain.Address
	if err := row.Scan(&a.ID, &a.AccountID, &a.Label, &a.RecipientName, &a.Line1, &a.Line2, &a.City, &a.Region, &a.PostalCode, &a.CountryCode, &a.Phone, &a.IsDefaultShipping, &a.IsDefaultBilling, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("scan address: %w", err)
	}
	return &a, nil
}
