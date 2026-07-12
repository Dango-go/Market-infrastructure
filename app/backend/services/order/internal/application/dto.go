package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type OrderItemView struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	Quantity       int32     `json:"quantity"`
	UnitPriceCents int64     `json:"unit_price_cents"`
	LineTotalCents int64     `json:"line_total_cents"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OrderView struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	CartID          *uuid.UUID      `json:"cart_id,omitempty"`
	Status          string          `json:"status"`
	Currency        string          `json:"currency"`
	SubtotalCents   int64           `json:"subtotal_cents"`
	ShippingCents   int64           `json:"shipping_cents"`
	TotalCents      int64           `json:"total_cents"`
	DeliveryMethod  string          `json:"delivery_method"`
	DeliveryAddress string          `json:"delivery_address"`
	CustomerNote    string          `json:"customer_note,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Items           []OrderItemView `json:"items"`
}
