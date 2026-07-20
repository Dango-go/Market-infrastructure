package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProductStatus string

const (
	StatusDraft     ProductStatus = "draft"
	StatusActive    ProductStatus = "active"
	StatusArchived  ProductStatus = "archived"
)

type Product struct {
	ID                uuid.UUID
	CategoryID        uuid.UUID
	BrandID           uuid.UUID
	Slug              string
	SKU               string
	Name              string
	ShortDescription  string
	Description       string
	DatasheetURL      string
	ImageURL          string
	Status            ProductStatus
	Featured          bool
	CreatedBy         uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Specs             []ProductSpec
	Media             []ProductMedia
	Compatibility     []CompatibilityRule
}

type ProductSpec struct {
	Key   string
	Value string
}

type ProductMedia struct {
	URL      string
	Type     string
	SortOrder int32
}

type CompatibilityRule struct {
	Kind  string
	Value string
}

func NewProduct(id, categoryID, brandID, createdBy uuid.UUID, slug, sku, name, shortDescription, description, datasheetURL, imageURL string, featured bool, status ProductStatus, now time.Time) *Product {
	if status == "" {
		status = StatusDraft
	}
	return &Product{
		ID:               id,
		CategoryID:       categoryID,
		BrandID:          brandID,
		Slug:             normalizeSlug(slug),
		SKU:              strings.TrimSpace(strings.ToUpper(sku)),
		Name:             strings.TrimSpace(name),
		ShortDescription: strings.TrimSpace(shortDescription),
		Description:      strings.TrimSpace(description),
		DatasheetURL:     strings.TrimSpace(datasheetURL),
		ImageURL:         strings.TrimSpace(imageURL),
		Status:           status,
		Featured:         featured,
		CreatedBy:        createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (p *Product) ReplaceSpecs(specs []ProductSpec) {
	p.Specs = normalizeSpecs(specs)
}

func (p *Product) ReplaceMedia(media []ProductMedia) {
	p.Media = normalizeMedia(media)
}

func (p *Product) ReplaceCompatibility(items []CompatibilityRule) {
	p.Compatibility = normalizeCompatibility(items)
}

func (p *Product) Update(categoryID, brandID *uuid.UUID, slug, sku, name, shortDescription, description, datasheetURL, imageURL *string, featured *bool, status *ProductStatus, now time.Time) {
	if categoryID != nil {
		p.CategoryID = *categoryID
	}
	if brandID != nil {
		p.BrandID = *brandID
	}
	if slug != nil {
		p.Slug = normalizeSlug(*slug)
	}
	if sku != nil {
		p.SKU = strings.TrimSpace(strings.ToUpper(*sku))
	}
	if name != nil {
		p.Name = strings.TrimSpace(*name)
	}
	if shortDescription != nil {
		p.ShortDescription = strings.TrimSpace(*shortDescription)
	}
	if description != nil {
		p.Description = strings.TrimSpace(*description)
	}
	if datasheetURL != nil {
		p.DatasheetURL = strings.TrimSpace(*datasheetURL)
	}
	if imageURL != nil {
		p.ImageURL = strings.TrimSpace(*imageURL)
	}
	if featured != nil {
		p.Featured = *featured
	}
	if status != nil {
		p.Status = *status
	}
	p.UpdatedAt = now
}

func normalizeSlug(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	v = strings.ReplaceAll(v, " ", "-")
	return v
}

func normalizeSpecs(items []ProductSpec) []ProductSpec {
	out := make([]ProductSpec, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		value := strings.TrimSpace(item.Value)
		if key == "" || value == "" {
			continue
		}
		out = append(out, ProductSpec{Key: key, Value: value})
	}
	return out
}

func normalizeMedia(items []ProductMedia) []ProductMedia {
	out := make([]ProductMedia, 0, len(items))
	for _, item := range items {
		url := strings.TrimSpace(item.URL)
		typ := strings.TrimSpace(item.Type)
		if url == "" || typ == "" {
			continue
		}
		out = append(out, ProductMedia{URL: url, Type: typ, SortOrder: item.SortOrder})
	}
	return out
}

func normalizeCompatibility(items []CompatibilityRule) []CompatibilityRule {
	out := make([]CompatibilityRule, 0, len(items))
	for _, item := range items {
		kind := strings.TrimSpace(item.Kind)
		value := strings.TrimSpace(item.Value)
		if kind == "" || value == "" {
			continue
		}
		out = append(out, CompatibilityRule{Kind: kind, Value: value})
	}
	return out
}
