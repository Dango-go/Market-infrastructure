package domain

import (
	"time"

	"github.com/google/uuid"
)

type CartStatus string

const (
	CartActive    CartStatus = "active"
	CartCheckedOut CartStatus = "checked_out"
)

type Cart struct {
	ID           uuid.UUID
	AccountID    uuid.UUID
	Status       CartStatus
	Currency     string
	SubtotalCents int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CheckedOutAt *time.Time
	Items        []CartItem
}

type CartItem struct {
	ID            uuid.UUID
	CartID        uuid.UUID
	ProductID     uuid.UUID
	Quantity      int32
	UnitPriceCents int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewCart(id, accountID uuid.UUID, currency string, now time.Time) *Cart {
	return &Cart{ID: id, AccountID: accountID, Status: CartActive, Currency: currency, CreatedAt: now, UpdatedAt: now}
}

func NewCartItem(id, cartID, productID uuid.UUID, quantity int32, unitPriceCents int64, now time.Time) *CartItem {
	return &CartItem{ID: id, CartID: cartID, ProductID: productID, Quantity: quantity, UnitPriceCents: unitPriceCents, CreatedAt: now, UpdatedAt: now}
}

func (c *Cart) Recalculate(items []CartItem, now time.Time) {
	c.Items = items
	var subtotal int64
	for _, item := range items {
		subtotal += int64(item.Quantity) * item.UnitPriceCents
	}
	c.SubtotalCents = subtotal
	c.UpdatedAt = now
}

func (c *Cart) Checkout(now time.Time) {
	c.Status = CartCheckedOut
	c.CheckedOutAt = &now
	c.UpdatedAt = now
}
