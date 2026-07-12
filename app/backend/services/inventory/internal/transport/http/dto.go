package http

type warehouseCreateRequest struct {
	Code     string `json:"code" validate:"required,max=40"`
	Name     string `json:"name" validate:"required,max=160"`
	Location string `json:"location" validate:"omitempty,max=255"`
}

type adjustStockRequest struct {
	ProductID   string `json:"product_id" validate:"required,uuid"`
	WarehouseID string `json:"warehouse_id" validate:"required,uuid"`
	Delta       int64  `json:"delta" validate:"required"`
}

type reserveStockRequest struct {
	ProductID   string `json:"product_id" validate:"required,uuid"`
	WarehouseID string `json:"warehouse_id" validate:"required,uuid"`
	Reference   string `json:"reference" validate:"required,max=120"`
	Quantity    int64  `json:"quantity" validate:"required"`
}
