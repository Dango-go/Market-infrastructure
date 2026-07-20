package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type CategoryView struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BrandView struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CountryCode string    `json:"country_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductSpecView struct { Key string `json:"key"`; Value string `json:"value"` }

type ProductMediaView struct { URL string `json:"url"`; Type string `json:"type"`; SortOrder int32 `json:"sort_order"` }

type CompatibilityView struct { Kind string `json:"kind"`; Value string `json:"value"` }

type ProductView struct {
	ID               uuid.UUID           `json:"id"`
	CategoryID       uuid.UUID           `json:"category_id"`
	BrandID          uuid.UUID           `json:"brand_id"`
	Slug             string              `json:"slug"`
	SKU              string              `json:"sku"`
	Name             string              `json:"name"`
	ShortDescription string              `json:"short_description"`
	Description      string              `json:"description"`
	DatasheetURL     string              `json:"datasheet_url"`
	ImageURL         string              `json:"image_url"`
	Status           string              `json:"status"`
	Featured         bool                `json:"featured"`
	CreatedBy        uuid.UUID           `json:"created_by"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Specs            []ProductSpecView   `json:"specs"`
	Media            []ProductMediaView  `json:"media"`
	Compatibility    []CompatibilityView `json:"compatibility"`
}
