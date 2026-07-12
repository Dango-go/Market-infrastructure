package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type TemplateView struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Channel   string    `json:"channel"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationView struct {
	ID           uuid.UUID  `json:"id"`
	AccountID    uuid.UUID  `json:"account_id"`
	TemplateID   *uuid.UUID `json:"template_id,omitempty"`
	Channel      string     `json:"channel"`
	Status       string     `json:"status"`
	Subject      string     `json:"subject"`
	Body         string     `json:"body"`
	MetadataJSON string     `json:"metadata_json,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	ReadAt       *time.Time `json:"read_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
