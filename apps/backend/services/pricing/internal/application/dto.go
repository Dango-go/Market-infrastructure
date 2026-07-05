package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type PriceView struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	Currency       string    `json:"currency"`
	AmountCents    int64     `json:"amount_cents"`
	CompareAtCents int64     `json:"compare_at_cents"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PromotionView struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	DiscountType string     `json:"discount_type"`
	ValueCents   int64      `json:"value_cents"`
	PercentOff   int        `json:"percent_off"`
	Active       bool       `json:"active"`
	StartsAt     time.Time  `json:"starts_at"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
