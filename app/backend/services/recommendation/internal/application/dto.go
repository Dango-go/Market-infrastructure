package application

import "github.com/google/uuid"

type RecommendationView struct {
	ProductID uuid.UUID `json:"product_id"`
	Score     float64   `json:"score"`
	Reason    string    `json:"reason"`
}

type ProductProfileView struct {
	ProductID    uuid.UUID `json:"product_id"`
	Slug         string    `json:"slug"`
	CategorySlug string    `json:"category_slug"`
	BrandSlug    string    `json:"brand_slug"`
	Tags         []string  `json:"tags"`
	PriceCents   int64     `json:"price_cents"`
	Available    bool      `json:"available"`
}
