package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/notification/internal/domain"
	"github.com/google/uuid"
)

type NotificationUseCase struct{ Deps }

func NewNotificationUseCase(d Deps) *NotificationUseCase { return &NotificationUseCase{Deps: d} }

type CreateTemplateInput struct {
	Code    string
	Channel string
	Subject string
	Body    string
}

type CreateNotificationInput struct {
	AccountID    uuid.UUID
	TemplateID   *uuid.UUID
	Channel      string
	Subject      string
	Body         string
	MetadataJSON string
}

func (uc *NotificationUseCase) CreateTemplate(ctx context.Context, input CreateTemplateInput) (TemplateView, error) {
	channel, err := parseChannel(input.Channel)
	if err != nil { return TemplateView{}, err }
	if strings.TrimSpace(input.Code) == "" { return TemplateView{}, apperr.Invalid("invalid_code", "code is required") }
	if strings.TrimSpace(input.Subject) == "" { return TemplateView{}, apperr.Invalid("invalid_subject", "subject is required") }
	if strings.TrimSpace(input.Body) == "" { return TemplateView{}, apperr.Invalid("invalid_body", "body is required") }
	item := domain.NewTemplate(uc.IDs.New(), strings.TrimSpace(input.Code), channel, strings.TrimSpace(input.Subject), strings.TrimSpace(input.Body), uc.Clock.Now())
	if err := uc.Store.Templates().Create(ctx, item); err != nil { return TemplateView{}, err }
	return toTemplateView(item), nil
}

func (uc *NotificationUseCase) ListTemplates(ctx context.Context, limit, offset int32) ([]TemplateView, int64, error) {
	items, total, err := uc.Store.Templates().List(ctx, limit, offset)
	if err != nil { return nil, 0, err }
	out := make([]TemplateView, 0, len(items))
	for i := range items { out = append(out, toTemplateView(&items[i])) }
	return out, total, nil
}

func (uc *NotificationUseCase) CreateNotification(ctx context.Context, input CreateNotificationInput, req RequestContext) (NotificationView, error) {
	channel, err := parseChannel(input.Channel)
	if err != nil { return NotificationView{}, err }
	if strings.TrimSpace(input.Subject) == "" { return NotificationView{}, apperr.Invalid("invalid_subject", "subject is required") }
	if strings.TrimSpace(input.Body) == "" { return NotificationView{}, apperr.Invalid("invalid_body", "body is required") }

	var out NotificationView
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if input.TemplateID != nil {
			if _, err := tx.Templates().GetByID(ctx, *input.TemplateID); err != nil { return err }
		}
		now := uc.Clock.Now()
		item := domain.NewNotification(uc.IDs.New(), input.AccountID, input.TemplateID, channel, strings.TrimSpace(input.Subject), strings.TrimSpace(input.Body), strings.TrimSpace(input.MetadataJSON), now)
		if err := tx.Notifications().Create(ctx, item); err != nil { return err }
		out = toNotificationView(item)
		return nil
	})
	return out, err
}

func (uc *NotificationUseCase) GetNotification(ctx context.Context, accountID, notificationID uuid.UUID) (NotificationView, error) {
	item, err := uc.Store.Notifications().GetByID(ctx, notificationID)
	if err != nil { return NotificationView{}, err }
	if item.AccountID != accountID { return NotificationView{}, domain.ErrNotificationNotFound }
	return toNotificationView(item), nil
}

func (uc *NotificationUseCase) ListNotifications(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]NotificationView, int64, error) {
	items, total, err := uc.Store.Notifications().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil { return nil, 0, err }
	out := make([]NotificationView, 0, len(items))
	for i := range items { out = append(out, toNotificationView(&items[i])) }
	return out, total, nil
}

func (uc *NotificationUseCase) MarkSent(ctx context.Context, accountID, notificationID uuid.UUID, req RequestContext) (NotificationView, error) {
	var out NotificationView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		item, err := tx.Notifications().GetByID(ctx, notificationID)
		if err != nil { return err }
		if item.AccountID != accountID { return domain.ErrNotificationNotFound }
		now := uc.Clock.Now()
		item.MarkSent(now)
		if err := tx.Notifications().Update(ctx, item); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicNotificationSent, uc.Source, item.ID.String(), req.CorrelationID, now, events.NotificationSent{NotificationID: item.ID.String(), RecipientID: item.AccountID.String(), Channel: string(item.Channel)})
		if err != nil { return fmt.Errorf("build notification sent event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toNotificationView(item)
		return nil
	})
	return out, err
}

func (uc *NotificationUseCase) MarkRead(ctx context.Context, accountID, notificationID uuid.UUID) (NotificationView, error) {
	var out NotificationView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		item, err := tx.Notifications().GetByID(ctx, notificationID)
		if err != nil { return err }
		if item.AccountID != accountID { return domain.ErrNotificationNotFound }
		item.MarkRead(uc.Clock.Now())
		if err := tx.Notifications().Update(ctx, item); err != nil { return err }
		out = toNotificationView(item)
		return nil
	})
	return out, err
}

func parseChannel(raw string) (domain.NotificationChannel, error) {
	channel := domain.NotificationChannel(strings.ToLower(strings.TrimSpace(raw)))
	switch channel {
	case domain.ChannelEmail, domain.ChannelSMS, domain.ChannelPush, domain.ChannelInApp:
		return channel, nil
	default:
		return "", apperr.Invalid("invalid_channel", "channel must be one of: email, sms, push, in_app")
	}
}

func toTemplateView(item *domain.Template) TemplateView {
	return TemplateView{ID: item.ID, Code: item.Code, Channel: string(item.Channel), Subject: item.Subject, Body: item.Body, Active: item.Active, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func toNotificationView(item *domain.Notification) NotificationView {
	return NotificationView{ID: item.ID, AccountID: item.AccountID, TemplateID: item.TemplateID, Channel: string(item.Channel), Status: string(item.Status), Subject: item.Subject, Body: item.Body, MetadataJSON: item.MetadataJSON, SentAt: item.SentAt, ReadAt: item.ReadAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
