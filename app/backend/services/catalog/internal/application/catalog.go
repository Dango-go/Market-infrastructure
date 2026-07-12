package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/catalog/internal/domain"
	"github.com/google/uuid"
)

type CatalogUseCase struct{ Deps }

func NewCatalogUseCase(d Deps) *CatalogUseCase { return &CatalogUseCase{Deps: d} }

type CreateCategoryInput struct { Name, Slug, Description string }
type CreateBrandInput struct { Name, Slug, Description, CountryCode string }

type CreateProductInput struct {
	CategoryID       uuid.UUID
	BrandID          uuid.UUID
	Slug             string
	SKU              string
	Name             string
	ShortDescription string
	Description      string
	DatasheetURL     string
	ImageURL         string
	Status           string
	Featured         bool
	Specs            []domain.ProductSpec
	Media            []domain.ProductMedia
	Compatibility    []domain.CompatibilityRule
}

type UpdateProductInput struct {
	CategoryID       *uuid.UUID
	BrandID          *uuid.UUID
	Slug             *string
	SKU              *string
	Name             *string
	ShortDescription *string
	Description      *string
	DatasheetURL     *string
	ImageURL         *string
	Status           *string
	Featured         *bool
	Specs            []domain.ProductSpec
	Media            []domain.ProductMedia
	Compatibility    []domain.CompatibilityRule
	ReplaceSpecs         bool
	ReplaceMedia         bool
	ReplaceCompatibility bool
}

func (uc *CatalogUseCase) CreateCategory(ctx context.Context, input CreateCategoryInput) (CategoryView, error) {
	cat := domain.NewCategory(uc.IDs.New(), input.Name, input.Slug, input.Description, uc.Clock.Now())
	if err := uc.Store.Categories().Create(ctx, cat); err != nil { return CategoryView{}, err }
	return CategoryView{ID: cat.ID, Name: cat.Name, Slug: cat.Slug, Description: cat.Description, CreatedAt: cat.CreatedAt, UpdatedAt: cat.UpdatedAt}, nil
}

func (uc *CatalogUseCase) ListCategories(ctx context.Context) ([]CategoryView, error) {
	items, err := uc.Store.Categories().List(ctx)
	if err != nil { return nil, err }
	out := make([]CategoryView, 0, len(items))
	for _, item := range items { out = append(out, CategoryView{ID: item.ID, Name: item.Name, Slug: item.Slug, Description: item.Description, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}) }
	return out, nil
}

func (uc *CatalogUseCase) CreateBrand(ctx context.Context, input CreateBrandInput) (BrandView, error) {
	brand := domain.NewBrand(uc.IDs.New(), input.Name, input.Slug, input.Description, input.CountryCode, uc.Clock.Now())
	if err := uc.Store.Brands().Create(ctx, brand); err != nil { return BrandView{}, err }
	return BrandView{ID: brand.ID, Name: brand.Name, Slug: brand.Slug, Description: brand.Description, CountryCode: brand.CountryCode, CreatedAt: brand.CreatedAt, UpdatedAt: brand.UpdatedAt}, nil
}

func (uc *CatalogUseCase) ListBrands(ctx context.Context) ([]BrandView, error) {
	items, err := uc.Store.Brands().List(ctx)
	if err != nil { return nil, err }
	out := make([]BrandView, 0, len(items))
	for _, item := range items { out = append(out, BrandView{ID: item.ID, Name: item.Name, Slug: item.Slug, Description: item.Description, CountryCode: item.CountryCode, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}) }
	return out, nil
}

func (uc *CatalogUseCase) CreateProduct(ctx context.Context, input CreateProductInput, req RequestContext) (ProductView, error) {
	var out ProductView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if _, err := tx.Categories().GetByID(ctx, input.CategoryID); err != nil { return err }
		if _, err := tx.Brands().GetByID(ctx, input.BrandID); err != nil { return err }
		status := domain.ProductStatus(strings.TrimSpace(input.Status))
		product := domain.NewProduct(uc.IDs.New(), input.CategoryID, input.BrandID, req.AccountID, input.Slug, input.SKU, input.Name, input.ShortDescription, input.Description, input.DatasheetURL, input.ImageURL, input.Featured, status, uc.Clock.Now())
		product.ReplaceSpecs(input.Specs)
		product.ReplaceMedia(input.Media)
		product.ReplaceCompatibility(input.Compatibility)
		if err := tx.Products().Create(ctx, product); err != nil { return err }
		if err := emitProductCreated(ctx, tx, uc, product, req); err != nil { return err }
		if len(product.Compatibility) > 0 {
			if err := emitCompatibilityUpdated(ctx, tx, uc, product, req); err != nil { return err }
		}
		out = toProductView(product)
		return nil
	})
	return out, err
}

func (uc *CatalogUseCase) UpdateProduct(ctx context.Context, id uuid.UUID, input UpdateProductInput, req RequestContext) (ProductView, error) {
	var out ProductView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		product, err := tx.Products().GetByID(ctx, id)
		if err != nil { return err }
		if input.CategoryID != nil {
			if _, err := tx.Categories().GetByID(ctx, *input.CategoryID); err != nil { return err }
		}
		if input.BrandID != nil {
			if _, err := tx.Brands().GetByID(ctx, *input.BrandID); err != nil { return err }
		}
		var status *domain.ProductStatus
		if input.Status != nil {
			v := domain.ProductStatus(strings.TrimSpace(*input.Status))
			status = &v
		}
		product.Update(input.CategoryID, input.BrandID, input.Slug, input.SKU, input.Name, input.ShortDescription, input.Description, input.DatasheetURL, input.ImageURL, input.Featured, status, uc.Clock.Now())
		if input.ReplaceSpecs { product.ReplaceSpecs(input.Specs) }
		if input.ReplaceMedia { product.ReplaceMedia(input.Media) }
		if input.ReplaceCompatibility { product.ReplaceCompatibility(input.Compatibility) }
		if err := tx.Products().Update(ctx, product); err != nil { return err }
		if input.ReplaceSpecs {
			if err := tx.Products().ReplaceSpecs(ctx, product.ID, product.Specs); err != nil { return err }
		}
		if input.ReplaceMedia {
			if err := tx.Products().ReplaceMedia(ctx, product.ID, product.Media); err != nil { return err }
		}
		if input.ReplaceCompatibility {
			if err := tx.Products().ReplaceCompatibility(ctx, product.ID, product.Compatibility); err != nil { return err }
			if err := emitCompatibilityUpdated(ctx, tx, uc, product, req); err != nil { return err }
		}
		if err := emitProductUpdated(ctx, tx, uc, product, req); err != nil { return err }
		out = toProductView(product)
		return nil
	})
	return out, err
}

func (uc *CatalogUseCase) GetProduct(ctx context.Context, slug string) (ProductView, error) {
	product, err := uc.Store.Products().GetBySlug(ctx, slug)
	if err != nil { return ProductView{}, err }
	return toProductView(product), nil
}

func (uc *CatalogUseCase) ListProducts(ctx context.Context, filters domain.ProductFilters, limit, offset int32) ([]ProductView, int64, error) {
	items, total, err := uc.Store.Products().List(ctx, filters, limit, offset)
	if err != nil { return nil, 0, err }
	out := make([]ProductView, 0, len(items))
	for _, item := range items { out = append(out, toProductView(item)) }
	return out, total, nil
}

func toProductView(p *domain.Product) ProductView {
	specs := make([]ProductSpecView, 0, len(p.Specs))
	for _, item := range p.Specs { specs = append(specs, ProductSpecView{Key: item.Key, Value: item.Value}) }
	media := make([]ProductMediaView, 0, len(p.Media))
	for _, item := range p.Media { media = append(media, ProductMediaView{URL: item.URL, Type: item.Type, SortOrder: item.SortOrder}) }
	compatibility := make([]CompatibilityView, 0, len(p.Compatibility))
	for _, item := range p.Compatibility { compatibility = append(compatibility, CompatibilityView{Kind: item.Kind, Value: item.Value}) }
	return ProductView{ID: p.ID, CategoryID: p.CategoryID, BrandID: p.BrandID, Slug: p.Slug, SKU: p.SKU, Name: p.Name, ShortDescription: p.ShortDescription, Description: p.Description, DatasheetURL: p.DatasheetURL, ImageURL: p.ImageURL, Status: string(p.Status), Featured: p.Featured, CreatedBy: p.CreatedBy, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, Specs: specs, Media: media, Compatibility: compatibility}
}

func emitProductCreated(ctx context.Context, tx domain.Store, uc *CatalogUseCase, product *domain.Product, req RequestContext) error {
	env, err := events.NewEnvelope(uc.IDs.New(), events.TopicProductCreated, uc.Source, product.ID.String(), req.CorrelationID, uc.Clock.Now(), events.ProductCreated{ProductID: product.ID.String(), CategoryID: product.CategoryID.String(), BrandID: product.BrandID.String(), Slug: product.Slug, SKU: product.SKU, Name: product.Name, Status: string(product.Status), Featured: product.Featured})
	if err != nil { return fmt.Errorf("build product created event: %w", err) }
	return tx.Outbox().Enqueue(ctx, env)
}

func emitProductUpdated(ctx context.Context, tx domain.Store, uc *CatalogUseCase, product *domain.Product, req RequestContext) error {
	env, err := events.NewEnvelope(uc.IDs.New(), events.TopicProductUpdated, uc.Source, product.ID.String(), req.CorrelationID, uc.Clock.Now(), events.ProductUpdated{ProductID: product.ID.String(), CategoryID: product.CategoryID.String(), BrandID: product.BrandID.String(), Slug: product.Slug, SKU: product.SKU, Name: product.Name, Status: string(product.Status), Featured: product.Featured})
	if err != nil { return fmt.Errorf("build product updated event: %w", err) }
	return tx.Outbox().Enqueue(ctx, env)
}

func emitCompatibilityUpdated(ctx context.Context, tx domain.Store, uc *CatalogUseCase, product *domain.Product, req RequestContext) error {
	rules := make([]events.CompatibilityRulePayload, 0, len(product.Compatibility))
	for _, item := range product.Compatibility { rules = append(rules, events.CompatibilityRulePayload{Kind: item.Kind, Value: item.Value}) }
	env, err := events.NewEnvelope(uc.IDs.New(), events.TopicProductCompatibilityUpdated, uc.Source, product.ID.String(), req.CorrelationID, uc.Clock.Now(), events.ProductCompatibilityUpdated{ProductID: product.ID.String(), Slug: product.Slug, Rules: rules})
	if err != nil { return fmt.Errorf("build compatibility updated event: %w", err) }
	return tx.Outbox().Enqueue(ctx, env)
}
