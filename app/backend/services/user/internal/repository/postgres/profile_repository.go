package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type profileRepository struct{ db pgxConn }

func (r *profileRepository) CreateIfMissing(ctx context.Context, p *domain.Profile) (bool, error) {
	const q = `INSERT INTO profiles (account_id, email, username, display_name, bio, phone, avatar_url, locale, timezone, marketing_opt_in, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (account_id) DO NOTHING`
	tag, err := r.db.Exec(ctx, q, p.AccountID, p.Email, p.Username, p.DisplayName, p.Bio, p.Phone, p.AvatarURL, p.Locale, p.Timezone, p.MarketingOptIn, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return false, fmt.Errorf("insert profile: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *profileRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Profile, error) {
	const q = `SELECT account_id, email, username, display_name, bio, phone, avatar_url, locale, timezone, marketing_opt_in, created_at, updated_at, deleted_at FROM profiles WHERE account_id = $1 AND deleted_at IS NULL`
	row := r.db.QueryRow(ctx, q, accountID)
	return scanProfile(row)
}

func (r *profileRepository) Update(ctx context.Context, p *domain.Profile) error {
	const q = `UPDATE profiles SET email = $2, username = $3, display_name = $4, bio = $5, phone = $6, avatar_url = $7, locale = $8, timezone = $9, marketing_opt_in = $10, updated_at = $11, deleted_at = $12 WHERE account_id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, p.AccountID, p.Email, p.Username, p.DisplayName, p.Bio, p.Phone, p.AvatarURL, p.Locale, p.Timezone, p.MarketingOptIn, p.UpdatedAt, p.DeletedAt)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

func scanProfile(row interface{ Scan(dest ...any) error }) (*domain.Profile, error) {
	var p domain.Profile
	if err := row.Scan(&p.AccountID, &p.Email, &p.Username, &p.DisplayName, &p.Bio, &p.Phone, &p.AvatarURL, &p.Locale, &p.Timezone, &p.MarketingOptIn, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, fmt.Errorf("scan profile: %w", err)
	}
	return &p, nil
}
