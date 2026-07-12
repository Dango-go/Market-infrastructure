package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type CartItemView struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	Quantity       int32     `json:"quantity"`
	UnitPriceCents int64     `json:"unit_price_cents"`
	LineTotalCents int64     `json:"line_total_cents"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CartView struct {
	ID            uuid.UUID      `json:"id"`
	AccountID     uuid.UUID      `json:"account_id"`
	Status        string         `json:"status"`
	Currency      string         `json:"currency"`
	SubtotalCents int64          `json:"subtotal_cents"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CheckedOutAt  *time.Time     `json:"checked_out_at,omitempty"`
	Items         []CartItemView `json:"items"`
}
