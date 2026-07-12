package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
}

type WishlistView struct {
	AccountID uuid.UUID `json:"account_id"`
	ProductID uuid.UUID `json:"product_id"`
	AddedAt   time.Time `json:"added_at"`
}
