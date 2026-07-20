package http

type createOrderItemRequest struct {
	ProductID      string `json:"product_id" validate:"required,uuid"`
	Quantity       int32  `json:"quantity" validate:"required"`
	UnitPriceCents int64  `json:"unit_price_cents" validate:"required"`
}

type createOrderRequest struct {
	CartID          *string                  `json:"cart_id"`
	Currency        string                   `json:"currency" validate:"required,max=8"`
	ShippingCents   int64                    `json:"shipping_cents"`
	DeliveryMethod  string                   `json:"delivery_method" validate:"required,max=80"`
	DeliveryAddress string                   `json:"delivery_address" validate:"required,max=500"`
	CustomerNote    string                   `json:"customer_note" validate:"omitempty,max=500"`
	Items           []createOrderItemRequest `json:"items" validate:"required,min=1"`
}

type updateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid processing shipped completed cancelled"`
}
