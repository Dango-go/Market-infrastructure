package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrWarehouseNotFound   = apperr.NotFound("warehouse_not_found", "warehouse not found")
	ErrStockItemNotFound   = apperr.NotFound("stock_item_not_found", "stock item not found")
	ErrReservationNotFound = apperr.NotFound("reservation_not_found", "reservation not found")
	ErrInsufficientStock   = apperr.Conflict("insufficient_stock", "insufficient available stock")
)
