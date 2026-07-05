package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentSucceeded PaymentStatus = "succeeded"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	AccountID      uuid.UUID
	Status         PaymentStatus
	Provider       string
	Method         string
	Currency       string
	AmountCents    int64
	TransactionRef string
	FailureReason  string
	PaidAt         *time.Time
	RefundedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewPayment(id, orderID, accountID uuid.UUID, provider, method, currency string, amountCents int64, transactionRef string, now time.Time) *Payment {
	return &Payment{ID: id, OrderID: orderID, AccountID: accountID, Status: PaymentPending, Provider: provider, Method: method, Currency: currency, AmountCents: amountCents, TransactionRef: transactionRef, CreatedAt: now, UpdatedAt: now}
}

func (p *Payment) Succeed(transactionRef string, now time.Time) {
	p.Status = PaymentSucceeded
	if transactionRef != "" { p.TransactionRef = transactionRef }
	p.FailureReason = ""
	paidAt := now
	p.PaidAt = &paidAt
	p.UpdatedAt = now
}

func (p *Payment) Fail(reason, transactionRef string, now time.Time) {
	p.Status = PaymentFailed
	if transactionRef != "" { p.TransactionRef = transactionRef }
	p.FailureReason = reason
	p.UpdatedAt = now
}

func (p *Payment) Refund(now time.Time) {
	p.Status = PaymentRefunded
	refundedAt := now
	p.RefundedAt = &refundedAt
	p.UpdatedAt = now
}
