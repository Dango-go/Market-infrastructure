package domain

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID               uuid.UUID
	ProductID        uuid.UUID
	Slug             string
	SKU              string
	Name             string
	ShortDescription string
	CategorySlug     string
	BrandSlug        string
	Tags             []string
	SpecsText        string
	PriceCents       int64
	Currency         string
	Available        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SearchFilters struct {
	Query        string
	CategorySlug string
	BrandSlug    string
	Available    *bool
}

func NewDocument(id, productID uuid.UUID, slug, sku, name, shortDescription, categorySlug, brandSlug string, tags []string, specsText string, priceCents int64, currency string, available bool, now time.Time) *Document {
	return &Document{ID: id, ProductID: productID, Slug: slug, SKU: sku, Name: name, ShortDescription: shortDescription, CategorySlug: categorySlug, BrandSlug: brandSlug, Tags: tags, SpecsText: specsText, PriceCents: priceCents, Currency: currency, Available: available, CreatedAt: now, UpdatedAt: now}
}

func (d *Document) Update(slug, sku, name, shortDescription, categorySlug, brandSlug string, tags []string, specsText string, priceCents int64, currency string, available bool, now time.Time) {
	d.Slug = slug
	d.SKU = sku
	d.Name = name
	d.ShortDescription = shortDescription
	d.CategorySlug = categorySlug
	d.BrandSlug = brandSlug
	d.Tags = tags
	d.SpecsText = specsText
	d.PriceCents = priceCents
	d.Currency = currency
	d.Available = available
	d.UpdatedAt = now
}
