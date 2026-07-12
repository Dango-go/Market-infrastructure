package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/payment/internal/domain"
	"github.com/google/uuid"
)

type PaymentUseCase struct{ Deps }

func NewPaymentUseCase(d Deps) *PaymentUseCase { return &PaymentUseCase{Deps: d} }

type CreatePaymentInput struct {
	OrderID        uuid.UUID
	Provider       string
	Method         string
	Currency       string
	AmountCents    int64
	TransactionRef string
}

type ConfirmPaymentInput struct {
	TransactionRef string
}

type FailPaymentInput struct {
	Reason         string
	TransactionRef string
}

func (uc *PaymentUseCase) Create(ctx context.Context, accountID uuid.UUID, input CreatePaymentInput, req RequestContext) (PaymentView, error) {
	if strings.TrimSpace(input.Provider) == "" {
		return PaymentView{}, apperr.Invalid("invalid_provider", "provider is required")
	}
	if strings.TrimSpace(input.Method) == "" {
		return PaymentView{}, apperr.Invalid("invalid_method", "method is required")
	}
	if strings.TrimSpace(input.Currency) == "" {
		return PaymentView{}, apperr.Invalid("invalid_currency", "currency is required")
	}
	if input.AmountCents <= 0 {
		return PaymentView{}, apperr.Invalid("invalid_amount", "amount_cents must be greater than zero")
	}

	var out PaymentView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		now := uc.Clock.Now()
		payment := domain.NewPayment(
			uc.IDs.New(),
			input.OrderID,
			accountID,
			strings.TrimSpace(input.Provider),
			strings.TrimSpace(input.Method),
			strings.ToUpper(strings.TrimSpace(input.Currency)),
			input.AmountCents,
			strings.TrimSpace(input.TransactionRef),
			now,
		)
		if err := tx.Payments().Create(ctx, payment); err != nil {
			return err
		}
		env, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicPaymentCreated,
			uc.Source,
			payment.ID.String(),
			req.CorrelationID,
			now,
			events.PaymentCreated{
				PaymentID:      payment.ID.String(),
				OrderID:        payment.OrderID.String(),
				AccountID:      payment.AccountID.String(),
				Status:         string(payment.Status),
				Provider:       payment.Provider,
				Method:         payment.Method,
				Currency:       payment.Currency,
				AmountCents:    payment.AmountCents,
				TransactionRef: payment.TransactionRef,
			},
		)
		if err != nil {
			return fmt.Errorf("build payment created event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toPaymentView(payment)
		return nil
	})
	return out, err
}

func (uc *PaymentUseCase) GetByID(ctx context.Context, accountID, paymentID uuid.UUID) (PaymentView, error) {
	payment, err := uc.Store.Payments().GetByID(ctx, paymentID)
	if err != nil {
		return PaymentView{}, err
	}
	if payment.AccountID != accountID {
		return PaymentView{}, domain.ErrPaymentNotFound
	}
	return toPaymentView(payment), nil
}

func (uc *PaymentUseCase) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]PaymentView, int64, error) {
	items, total, err := uc.Store.Payments().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]PaymentView, 0, len(items))
	for i := range items {
		out = append(out, toPaymentView(&items[i]))
	}
	return out, total, nil
}

func (uc *PaymentUseCase) Confirm(ctx context.Context, accountID, paymentID uuid.UUID, input ConfirmPaymentInput, req RequestContext) (PaymentView, error) {
	var out PaymentView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		payment, err := tx.Payments().GetByID(ctx, paymentID)
		if err != nil { return err }
		if payment.AccountID != accountID { return domain.ErrPaymentNotFound }
		now := uc.Clock.Now()
		payment.Succeed(strings.TrimSpace(input.TransactionRef), now)
		if err := tx.Payments().Update(ctx, payment); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicPaymentSucceeded, uc.Source, payment.ID.String(), req.CorrelationID, now, events.PaymentSucceeded{PaymentID: payment.ID.String(), OrderID: payment.OrderID.String(), AccountID: payment.AccountID.String(), Status: string(payment.Status), AmountCents: payment.AmountCents, Currency: payment.Currency, TransactionRef: payment.TransactionRef})
		if err != nil { return fmt.Errorf("build payment succeeded event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toPaymentView(payment)
		return nil
	})
	return out, err
}

func (uc *PaymentUseCase) Fail(ctx context.Context, accountID, paymentID uuid.UUID, input FailPaymentInput, req RequestContext) (PaymentView, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return PaymentView{}, apperr.Invalid("invalid_reason", "reason is required")
	}
	var out PaymentView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		payment, err := tx.Payments().GetByID(ctx, paymentID)
		if err != nil { return err }
		if payment.AccountID != accountID { return domain.ErrPaymentNotFound }
		now := uc.Clock.Now()
		payment.Fail(strings.TrimSpace(input.Reason), strings.TrimSpace(input.TransactionRef), now)
		if err := tx.Payments().Update(ctx, payment); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicPaymentFailed, uc.Source, payment.ID.String(), req.CorrelationID, now, events.PaymentFailed{PaymentID: payment.ID.String(), OrderID: payment.OrderID.String(), AccountID: payment.AccountID.String(), Status: string(payment.Status), FailureReason: payment.FailureReason, TransactionRef: payment.TransactionRef})
		if err != nil { return fmt.Errorf("build payment failed event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toPaymentView(payment)
		return nil
	})
	return out, err
}

func (uc *PaymentUseCase) Refund(ctx context.Context, accountID, paymentID uuid.UUID, req RequestContext) (PaymentView, error) {
	var out PaymentView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		payment, err := tx.Payments().GetByID(ctx, paymentID)
		if err != nil { return err }
		if payment.AccountID != accountID { return domain.ErrPaymentNotFound }
		now := uc.Clock.Now()
		payment.Refund(now)
		if err := tx.Payments().Update(ctx, payment); err != nil { return err }
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicPaymentRefunded, uc.Source, payment.ID.String(), req.CorrelationID, now, events.PaymentRefunded{PaymentID: payment.ID.String(), OrderID: payment.OrderID.String(), AccountID: payment.AccountID.String(), Status: string(payment.Status), AmountCents: payment.AmountCents, Currency: payment.Currency, TransactionRef: payment.TransactionRef})
		if err != nil { return fmt.Errorf("build payment refunded event: %w", err) }
		if err := tx.Outbox().Enqueue(ctx, env); err != nil { return err }
		out = toPaymentView(payment)
		return nil
	})
	return out, err
}

func toPaymentView(item *domain.Payment) PaymentView {
	return PaymentView{ID: item.ID, OrderID: item.OrderID, AccountID: item.AccountID, Status: string(item.Status), Provider: item.Provider, Method: item.Method, Currency: item.Currency, AmountCents: item.AmountCents, TransactionRef: item.TransactionRef, FailureReason: item.FailureReason, PaidAt: item.PaidAt, RefundedAt: item.RefundedAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
