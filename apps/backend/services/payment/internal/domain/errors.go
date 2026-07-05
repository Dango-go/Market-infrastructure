package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrPaymentNotFound = apperr.NotFound("payment_not_found", "payment not found")
)
