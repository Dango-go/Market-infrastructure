package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrCartNotFound     = apperr.NotFound("cart_not_found", "cart not found")
	ErrCartItemNotFound = apperr.NotFound("cart_item_not_found", "cart item not found")
)
