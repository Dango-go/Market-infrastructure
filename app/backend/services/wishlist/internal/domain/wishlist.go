package domain

import (
	"time"

	"github.com/google/uuid"
)

// WishlistItem tracks a saved product reference.
type WishlistItem struct {
	AccountID uuid.UUID
	ProductID uuid.UUID
	AddedAt   time.Time
}

func NewWishlistItem(accountID, productID uuid.UUID, now time.Time) *WishlistItem {
	return &WishlistItem{AccountID: accountID, ProductID: productID, AddedAt: now}
}
