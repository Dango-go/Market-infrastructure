package http

import "time"

type wishlistResponse struct {
	AccountID string    `json:"account_id"`
	ProductID string    `json:"product_id"`
	AddedAt   time.Time `json:"added_at"`
}
