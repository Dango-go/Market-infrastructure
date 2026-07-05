package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderPaid       OrderStatus = "paid"
	OrderProcessing OrderStatus = "processing"
	OrderShipped    OrderStatus = "shipped"
	OrderCompleted  OrderStatus = "completed"
	OrderCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID               uuid.UUID
	AccountID        uuid.UUID
	CartID           *uuid.UUID
	Status           OrderStatus
	Currency         string
	SubtotalCents    int64
	ShippingCents    int64
	TotalCents       int64
	DeliveryMethod   string
	DeliveryAddress  string
	CustomerNote     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Items            []OrderItem
}

type OrderItem struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewOrder(id, accountID uuid.UUID, cartID *uuid.UUID, currency, deliveryMethod, deliveryAddress, customerNote string, shippingCents int64, now time.Time) *Order {
	return &Order{
		ID:              id,
		AccountID:       accountID,
		CartID:          cartID,
		Status:          OrderPending,
		Currency:        currency,
		ShippingCents:   shippingCents,
		DeliveryMethod:  deliveryMethod,
		DeliveryAddress: deliveryAddress,
		CustomerNote:    customerNote,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func NewOrderItem(id, orderID, productID uuid.UUID, quantity int32, unitPriceCents int64, now time.Time) *OrderItem {
	return &OrderItem{ID: id, OrderID: orderID, ProductID: productID, Quantity: quantity, UnitPriceCents: unitPriceCents, CreatedAt: now, UpdatedAt: now}
}

func (o *Order) Recalculate(items []OrderItem, now time.Time) {
	o.Items = items
	var subtotal int64
	for _, item := range items {
		subtotal += int64(item.Quantity) * item.UnitPriceCents
	}
	o.SubtotalCents = subtotal
	o.TotalCents = subtotal + o.ShippingCents
	o.UpdatedAt = now
}

func (o *Order) SetStatus(status OrderStatus, now time.Time) {
	o.Status = status
	o.UpdatedAt = now
}
