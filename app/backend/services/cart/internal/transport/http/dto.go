package http

type addItemRequest struct {
	ProductID      string `json:"product_id" validate:"required,uuid"`
	Quantity       int32  `json:"quantity" validate:"required"`
	UnitPriceCents int64  `json:"unit_price_cents" validate:"required"`
}

type updateItemRequest struct {
	Quantity int32 `json:"quantity" validate:"required"`
}
