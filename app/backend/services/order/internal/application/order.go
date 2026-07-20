package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/order/internal/domain"
	"github.com/google/uuid"
)

type OrderUseCase struct{ Deps }

func NewOrderUseCase(d Deps) *OrderUseCase { return &OrderUseCase{Deps: d} }

type CreateOrderItemInput struct {
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
}

type CreateOrderInput struct {
	CartID          *uuid.UUID
	Currency        string
	ShippingCents   int64
	DeliveryMethod  string
	DeliveryAddress string
	CustomerNote    string
	Items           []CreateOrderItemInput
}

type UpdateStatusInput struct {
	Status string
}

func (uc *OrderUseCase) Create(ctx context.Context, accountID uuid.UUID, input CreateOrderInput, req RequestContext) (OrderView, error) {
	if len(input.Items) == 0 {
		return OrderView{}, apperr.Invalid("empty_order", "order must contain at least one item")
	}
	if strings.TrimSpace(input.Currency) == "" {
		return OrderView{}, apperr.Invalid("invalid_currency", "currency is required")
	}
	if strings.TrimSpace(input.DeliveryMethod) == "" {
		return OrderView{}, apperr.Invalid("invalid_delivery_method", "delivery_method is required")
	}
	if strings.TrimSpace(input.DeliveryAddress) == "" {
		return OrderView{}, apperr.Invalid("invalid_delivery_address", "delivery_address is required")
	}
	if input.ShippingCents < 0 {
		return OrderView{}, apperr.Invalid("invalid_shipping_cents", "shipping_cents must be zero or greater")
	}
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return OrderView{}, apperr.Invalid("invalid_quantity", "each item quantity must be greater than zero")
		}
		if item.UnitPriceCents < 0 {
			return OrderView{}, apperr.Invalid("invalid_unit_price", "each item unit_price_cents must be zero or greater")
		}
	}

	var out OrderView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		now := uc.Clock.Now()
		order := domain.NewOrder(
			uc.IDs.New(),
			accountID,
			input.CartID,
			strings.ToUpper(strings.TrimSpace(input.Currency)),
			strings.TrimSpace(input.DeliveryMethod),
			strings.TrimSpace(input.DeliveryAddress),
			strings.TrimSpace(input.CustomerNote),
			input.ShippingCents,
			now,
		)
		if err := tx.Orders().Create(ctx, order); err != nil {
			return err
		}
		for _, itemInput := range input.Items {
			item := domain.NewOrderItem(uc.IDs.New(), order.ID, itemInput.ProductID, itemInput.Quantity, itemInput.UnitPriceCents, now)
			if err := tx.Items().Create(ctx, item); err != nil {
				return err
			}
		}
		if err := refreshOrder(ctx, tx, order); err != nil {
			return err
		}
		env, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicOrderCreated,
			uc.Source,
			order.ID.String(),
			req.CorrelationID,
			now,
			buildOrderCreatedPayload(order),
		)
		if err != nil {
			return fmt.Errorf("build order created event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toOrderView(order)
		return nil
	})
	return out, err
}

func (uc *OrderUseCase) GetByID(ctx context.Context, accountID, orderID uuid.UUID) (OrderView, error) {
	order, err := uc.Store.Orders().GetByID(ctx, orderID)
	if err != nil {
		return OrderView{}, err
	}
	if order.AccountID != accountID {
		return OrderView{}, domain.ErrOrderNotFound
	}
	if err := refreshOrder(ctx, uc.Store, order); err != nil {
		return OrderView{}, err
	}
	return toOrderView(order), nil
}

func (uc *OrderUseCase) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]OrderView, int64, error) {
	orders, total, err := uc.Store.Orders().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderView, 0, len(orders))
	for i := range orders {
		order := &orders[i]
		if err := refreshOrder(ctx, uc.Store, order); err != nil {
			return nil, 0, err
		}
		out = append(out, toOrderView(order))
	}
	return out, total, nil
}

func (uc *OrderUseCase) UpdateStatus(ctx context.Context, accountID, orderID uuid.UUID, input UpdateStatusInput, req RequestContext) (OrderView, error) {
	status, err := parseStatus(input.Status)
	if err != nil {
		return OrderView{}, err
	}

	var out OrderView
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		order, err := tx.Orders().GetByID(ctx, orderID)
		if err != nil {
			return err
		}
		if order.AccountID != accountID {
			return domain.ErrOrderNotFound
		}
		if err := refreshOrder(ctx, tx, order); err != nil {
			return err
		}
		order.SetStatus(status, uc.Clock.Now())
		if err := tx.Orders().Update(ctx, order); err != nil {
			return err
		}
		env, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicOrderStatusUpdated,
			uc.Source,
			order.ID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.OrderStatusUpdated{OrderID: order.ID.String(), AccountID: order.AccountID.String(), Status: string(order.Status)},
		)
		if err != nil {
			return fmt.Errorf("build order status event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toOrderView(order)
		return nil
	})
	return out, err
}

func refreshOrder(ctx context.Context, store domain.Store, order *domain.Order) error {
	items, err := store.Items().ListByOrderID(ctx, order.ID)
	if err != nil {
		return err
	}
	order.Recalculate(items, order.UpdatedAt)
	return nil
}

func parseStatus(raw string) (domain.OrderStatus, error) {
	status := domain.OrderStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch status {
	case domain.OrderPending, domain.OrderPaid, domain.OrderProcessing, domain.OrderShipped, domain.OrderCompleted, domain.OrderCancelled:
		return status, nil
	default:
		return "", apperr.Invalid("invalid_status", "status must be one of: pending, paid, processing, shipped, completed, cancelled")
	}
}

func buildOrderCreatedPayload(order *domain.Order) events.OrderCreated {
	items := make([]events.OrderCreatedItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, events.OrderCreatedItem{
			ProductID:      item.ProductID.String(),
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		})
	}
	payload := events.OrderCreated{
		OrderID:         order.ID.String(),
		AccountID:       order.AccountID.String(),
		Status:          string(order.Status),
		Currency:        order.Currency,
		SubtotalCents:   order.SubtotalCents,
		ShippingCents:   order.ShippingCents,
		TotalCents:      order.TotalCents,
		DeliveryMethod:  order.DeliveryMethod,
		DeliveryAddress: order.DeliveryAddress,
		Items:           items,
	}
	if order.CartID != nil {
		payload.CartID = order.CartID.String()
	}
	return payload
}

func toOrderView(order *domain.Order) OrderView {
	items := make([]OrderItemView, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, OrderItemView{
			ID:             item.ID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: int64(item.Quantity) * item.UnitPriceCents,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return OrderView{
		ID:              order.ID,
		AccountID:       order.AccountID,
		CartID:          order.CartID,
		Status:          string(order.Status),
		Currency:        order.Currency,
		SubtotalCents:   order.SubtotalCents,
		ShippingCents:   order.ShippingCents,
		TotalCents:      order.TotalCents,
		DeliveryMethod:  order.DeliveryMethod,
		DeliveryAddress: order.DeliveryAddress,
		CustomerNote:    order.CustomerNote,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           items,
	}
}
