package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/notification/internal/domain"
	"github.com/google/uuid"
)

type notificationRepository struct{ db pgxConn }

func (r *notificationRepository) Create(ctx context.Context, item *domain.Notification) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO notifications (id, account_id, template_id, channel, status, subject, body, metadata_json, sent_at, read_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, item.ID, item.AccountID, item.TemplateID, string(item.Channel), string(item.Status), item.Subject, item.Body, item.MetadataJSON, item.SentAt, item.ReadAt, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("insert notification: %w", err) }
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return r.getOne(ctx, `SELECT id, account_id, template_id, channel, status, subject, body, metadata_json, sent_at, read_at, created_at, updated_at FROM notifications WHERE id = $1`, id)
}

func (r *notificationRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]domain.Notification, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, account_id, template_id, channel, status, subject, body, metadata_json, sent_at, read_at, created_at, updated_at, COUNT(*) OVER() AS total_count FROM notifications WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list notifications: %w", err) }
	defer rows.Close()
	items := make([]domain.Notification, 0)
	var total int64
	for rows.Next() {
		var item domain.Notification
		var channel, status string
		if err := rows.Scan(&item.ID, &item.AccountID, &item.TemplateID, &channel, &status, &item.Subject, &item.Body, &item.MetadataJSON, &item.SentAt, &item.ReadAt, &item.CreatedAt, &item.UpdatedAt, &total); err != nil { return nil, 0, fmt.Errorf("scan notification: %w", err) }
		item.Channel = domain.NotificationChannel(channel)
		item.Status = domain.NotificationStatus(status)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate notifications: %w", err) }
	return items, total, nil
}

func (r *notificationRepository) Update(ctx context.Context, item *domain.Notification) error {
	tag, err := r.db.Exec(ctx, `UPDATE notifications SET channel = $2, status = $3, subject = $4, body = $5, metadata_json = $6, sent_at = $7, read_at = $8, updated_at = $9 WHERE id = $1`, item.ID, string(item.Channel), string(item.Status), item.Subject, item.Body, item.MetadataJSON, item.SentAt, item.ReadAt, item.UpdatedAt)
	if err != nil { return fmt.Errorf("update notification: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrNotificationNotFound }
	return nil
}

func (r *notificationRepository) getOne(ctx context.Context, query string, arg any) (*domain.Notification, error) {
	var item domain.Notification
	var channel, status string
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.AccountID, &item.TemplateID, &channel, &status, &item.Subject, &item.Body, &item.MetadataJSON, &item.SentAt, &item.ReadAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrNotificationNotFound }
		return nil, fmt.Errorf("get notification: %w", err)
	}
	item.Channel = domain.NotificationChannel(channel)
	item.Status = domain.NotificationStatus(status)
	return &item, nil
}
