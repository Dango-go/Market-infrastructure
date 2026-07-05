package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type PaymentView struct {
	ID             uuid.UUID  `json:"id"`
	OrderID        uuid.UUID  `json:"order_id"`
	AccountID      uuid.UUID  `json:"account_id"`
	Status         string     `json:"status"`
	Provider       string     `json:"provider"`
	Method         string     `json:"method"`
	Currency       string     `json:"currency"`
	AmountCents    int64      `json:"amount_cents"`
	TransactionRef string     `json:"transaction_ref,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	RefundedAt     *time.Time `json:"refunded_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
