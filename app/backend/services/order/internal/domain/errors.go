package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrOrderNotFound = apperr.NotFound("order_not_found", "order not found")
)
