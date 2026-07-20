package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationStatus string

type NotificationChannel string

const (
	NotificationDraft NotificationStatus = "draft"
	NotificationSent  NotificationStatus = "sent"
	NotificationRead  NotificationStatus = "read"

	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
	ChannelInApp NotificationChannel = "in_app"
)

type Template struct {
	ID        uuid.UUID
	Code      string
	Channel   NotificationChannel
	Subject   string
	Body      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Notification struct {
	ID            uuid.UUID
	AccountID     uuid.UUID
	TemplateID    *uuid.UUID
	Channel       NotificationChannel
	Status        NotificationStatus
	Subject       string
	Body          string
	MetadataJSON  string
	SentAt        *time.Time
	ReadAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewTemplate(id uuid.UUID, code string, channel NotificationChannel, subject, body string, now time.Time) *Template {
	return &Template{ID: id, Code: code, Channel: channel, Subject: subject, Body: body, Active: true, CreatedAt: now, UpdatedAt: now}
}

func NewNotification(id, accountID uuid.UUID, templateID *uuid.UUID, channel NotificationChannel, subject, body, metadataJSON string, now time.Time) *Notification {
	return &Notification{ID: id, AccountID: accountID, TemplateID: templateID, Channel: channel, Status: NotificationDraft, Subject: subject, Body: body, MetadataJSON: metadataJSON, CreatedAt: now, UpdatedAt: now}
}

func (n *Notification) MarkSent(now time.Time) {
	n.Status = NotificationSent
	sentAt := now
	n.SentAt = &sentAt
	n.UpdatedAt = now
}

func (n *Notification) MarkRead(now time.Time) {
	n.Status = NotificationRead
	readAt := now
	n.ReadAt = &readAt
	n.UpdatedAt = now
}
