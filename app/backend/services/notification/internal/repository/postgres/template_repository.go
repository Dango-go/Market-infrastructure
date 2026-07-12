package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/notification/internal/domain"
	"github.com/google/uuid"
)

type templateRepository struct{ db pgxConn }

func (r *templateRepository) Create(ctx context.Context, item *domain.Template) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO notification_templates (id, code, channel, subject, body, active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, item.ID, item.Code, string(item.Channel), item.Subject, item.Body, item.Active, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("insert template: %w", err) }
	return nil
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Template, error) {
	var item domain.Template
	var channel string
	if err := r.db.QueryRow(ctx, `SELECT id, code, channel, subject, body, active, created_at, updated_at FROM notification_templates WHERE id = $1`, id).Scan(&item.ID, &item.Code, &channel, &item.Subject, &item.Body, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrTemplateNotFound }
		return nil, fmt.Errorf("get template: %w", err)
	}
	item.Channel = domain.NotificationChannel(channel)
	return &item, nil
}

func (r *templateRepository) List(ctx context.Context, limit, offset int32) ([]domain.Template, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, code, channel, subject, body, active, created_at, updated_at, COUNT(*) OVER() AS total_count FROM notification_templates ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list templates: %w", err) }
	defer rows.Close()
	items := make([]domain.Template, 0)
	var total int64
	for rows.Next() {
		var item domain.Template
		var channel string
		if err := rows.Scan(&item.ID, &item.Code, &channel, &item.Subject, &item.Body, &item.Active, &item.CreatedAt, &item.UpdatedAt, &total); err != nil { return nil, 0, fmt.Errorf("scan template: %w", err) }
		item.Channel = domain.NotificationChannel(channel)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate templates: %w", err) }
	return items, total, nil
}
