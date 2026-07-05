package domain

import "github.com/embedded-market/backend/pkg/apperr"

var ErrInvalidEventType = apperr.Invalid("invalid_event_type", "event type must be one of: page_view, product_view, search, add_to_cart, begin_checkout, purchase")
