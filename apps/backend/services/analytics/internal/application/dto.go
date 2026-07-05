package application

import "github.com/google/uuid"

type OverviewView struct {
	Days           int32 `json:"days"`
	TotalEvents    int64 `json:"total_events"`
	UniqueSessions int64 `json:"unique_sessions"`
	ProductViews   int64 `json:"product_views"`
	Searches       int64 `json:"searches"`
	AddToCarts     int64 `json:"add_to_carts"`
	Checkouts      int64 `json:"checkouts"`
	Purchases      int64 `json:"purchases"`
}

type TopProductView struct {
	ProductID       uuid.UUID `json:"product_id"`
	Views           int64     `json:"views"`
	AddToCarts      int64     `json:"add_to_carts"`
	Purchases       int64     `json:"purchases"`
	ConversionScore float64   `json:"conversion_score"`
}
