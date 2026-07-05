package application

import (
	"context"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/services/search/internal/domain"
	"github.com/google/uuid"
)

type SearchUseCase struct{ Deps }

func NewSearchUseCase(d Deps) *SearchUseCase { return &SearchUseCase{Deps: d} }

type UpsertDocumentInput struct {
	ID               *uuid.UUID
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
}

func (uc *SearchUseCase) UpsertDocument(ctx context.Context, input UpsertDocumentInput) (DocumentView, error) {
	if strings.TrimSpace(input.Slug) == "" { return DocumentView{}, apperr.Invalid("invalid_slug", "slug is required") }
	if strings.TrimSpace(input.Name) == "" { return DocumentView{}, apperr.Invalid("invalid_name", "name is required") }
	if strings.TrimSpace(input.Currency) == "" { return DocumentView{}, apperr.Invalid("invalid_currency", "currency is required") }
	if input.PriceCents < 0 { return DocumentView{}, apperr.Invalid("invalid_price", "price_cents must be zero or greater") }

	now := uc.Clock.Now()
	id := uc.IDs.New()
	if input.ID != nil { id = *input.ID }
	doc, err := uc.Store.Documents().GetByID(ctx, id)
	if err != nil && err != domain.ErrDocumentNotFound { return DocumentView{}, err }
	if err == domain.ErrDocumentNotFound {
		doc = domain.NewDocument(id, input.ProductID, strings.TrimSpace(input.Slug), strings.TrimSpace(input.SKU), strings.TrimSpace(input.Name), strings.TrimSpace(input.ShortDescription), strings.TrimSpace(input.CategorySlug), strings.TrimSpace(input.BrandSlug), cleanTags(input.Tags), strings.TrimSpace(input.SpecsText), input.PriceCents, strings.ToUpper(strings.TrimSpace(input.Currency)), input.Available, now)
	} else {
		doc.Update(strings.TrimSpace(input.Slug), strings.TrimSpace(input.SKU), strings.TrimSpace(input.Name), strings.TrimSpace(input.ShortDescription), strings.TrimSpace(input.CategorySlug), strings.TrimSpace(input.BrandSlug), cleanTags(input.Tags), strings.TrimSpace(input.SpecsText), input.PriceCents, strings.ToUpper(strings.TrimSpace(input.Currency)), input.Available, now)
	}
	if err := uc.Store.Documents().Upsert(ctx, doc); err != nil { return DocumentView{}, err }
	return toDocumentView(doc), nil
}

func (uc *SearchUseCase) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return uc.Store.Documents().Delete(ctx, id)
}

func (uc *SearchUseCase) Search(ctx context.Context, filters domain.SearchFilters, limit, offset int32) ([]DocumentView, int64, error) {
	items, total, err := uc.Store.Documents().Search(ctx, filters, limit, offset)
	if err != nil { return nil, 0, err }
	out := make([]DocumentView, 0, len(items))
	for i := range items { out = append(out, toDocumentView(&items[i])) }
	return out, total, nil
}

func (uc *SearchUseCase) Suggest(ctx context.Context, query string, limit int32) ([]string, error) {
	if strings.TrimSpace(query) == "" { return []string{}, nil }
	return uc.Store.Documents().Suggest(ctx, query, limit)
}

func cleanTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if v := strings.TrimSpace(tag); v != "" { out = append(out, v) }
	}
	return out
}

func toDocumentView(item *domain.Document) DocumentView {
	return DocumentView{ID: item.ID, ProductID: item.ProductID, Slug: item.Slug, SKU: item.SKU, Name: item.Name, ShortDescription: item.ShortDescription, CategorySlug: item.CategorySlug, BrandSlug: item.BrandSlug, Tags: item.Tags, SpecsText: item.SpecsText, PriceCents: item.PriceCents, Currency: item.Currency, Available: item.Available, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
