package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrWishlistNotFound = apperr.NotFound("wishlist_item_not_found", "wishlist item not found")
)
