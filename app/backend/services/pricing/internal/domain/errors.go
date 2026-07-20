package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrPriceNotFound     = apperr.NotFound("price_not_found", "price not found")
	ErrPromotionNotFound = apperr.NotFound("promotion_not_found", "promotion not found")
)
