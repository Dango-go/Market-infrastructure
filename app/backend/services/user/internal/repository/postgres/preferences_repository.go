package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type preferencesRepository struct{ db pgxConn }

func (r *preferencesRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Preferences, error) {
	const q = `SELECT account_id, currency, language, email_notifications, sms_notifications, push_notifications, marketing_opt_in, created_at, updated_at FROM preferences WHERE account_id = $1`
	row := r.db.QueryRow(ctx, q, accountID)
	var p domain.Preferences
	if err := row.Scan(&p.AccountID, &p.Currency, &p.Language, &p.EmailNotifications, &p.SMSNotifications, &p.PushNotifications, &p.MarketingOptIn, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrPreferencesNotFound
		}
		return nil, fmt.Errorf("scan preferences: %w", err)
	}
	return &p, nil
}

func (r *preferencesRepository) CreateIfMissing(ctx context.Context, p *domain.Preferences) (bool, error) {
	const q = `INSERT INTO preferences (account_id, currency, language, email_notifications, sms_notifications, push_notifications, marketing_opt_in, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (account_id) DO NOTHING`
	tag, err := r.db.Exec(ctx, q, p.AccountID, p.Currency, p.Language, p.EmailNotifications, p.SMSNotifications, p.PushNotifications, p.MarketingOptIn, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return false, fmt.Errorf("insert preferences: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *preferencesRepository) Update(ctx context.Context, p *domain.Preferences) error {
	const q = `UPDATE preferences SET currency = $2, language = $3, email_notifications = $4, sms_notifications = $5, push_notifications = $6, marketing_opt_in = $7, updated_at = $8 WHERE account_id = $1`
	tag, err := r.db.Exec(ctx, q, p.AccountID, p.Currency, p.Language, p.EmailNotifications, p.SMSNotifications, p.PushNotifications, p.MarketingOptIn, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update preferences: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPreferencesNotFound
	}
	return nil
}
