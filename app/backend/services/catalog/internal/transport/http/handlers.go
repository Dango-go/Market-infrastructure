package http

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/catalog/internal/application"
	"github.com/embedded-market/backend/services/catalog/internal/domain"
)

type UseCases struct { Catalog *application.CatalogUseCase }

type Handler struct { uc UseCases }
func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListProducts(c *gin.Context) {
	page := httpx.ParsePageParams(c)
	filters := domain.ProductFilters{Query: c.Query("q"), CategorySlug: c.Query("category_slug"), BrandSlug: c.Query("brand_slug"), Status: c.Query("status")}
	if raw := strings.TrimSpace(c.Query("featured")); raw != "" {
		v := strings.EqualFold(raw, "true") || raw == "1"
		filters.Featured = &v
	}
	items, total, err := h.uc.Catalog.ListProducts(c.Request.Context(), filters, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) GetProduct(c *gin.Context) {
	item, err := h.uc.Catalog.GetProduct(c.Request.Context(), c.Param("slug"))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req productCreateRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_category_id", "the category id is not a valid UUID")); return }
	brandID, err := uuid.Parse(req.BrandID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_brand_id", "the brand id is not a valid UUID")); return }
	item, err := h.uc.Catalog.CreateProduct(c.Request.Context(), application.CreateProductInput{CategoryID: categoryID, BrandID: brandID, Slug: req.Slug, SKU: req.SKU, Name: req.Name, ShortDescription: req.ShortDescription, Description: req.Description, DatasheetURL: req.DatasheetURL, ImageURL: req.ImageURL, Status: req.Status, Featured: req.Featured, Specs: toSpecs(req.Specs), Media: toMedia(req.Media), Compatibility: toCompatibility(req.Compatibility)}, requestContext(c))
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
	var req productUpdateRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	var categoryID *uuid.UUID
	if req.CategoryID != nil { v, err := uuid.Parse(*req.CategoryID); if err != nil { httpx.Fail(c, apperr.Invalid("invalid_category_id", "the category id is not a valid UUID")); return }; categoryID = &v }
	var brandID *uuid.UUID
	if req.BrandID != nil { v, err := uuid.Parse(*req.BrandID); if err != nil { httpx.Fail(c, apperr.Invalid("invalid_brand_id", "the brand id is not a valid UUID")); return }; brandID = &v }
	item, err := h.uc.Catalog.UpdateProduct(c.Request.Context(), id, application.UpdateProductInput{CategoryID: categoryID, BrandID: brandID, Slug: req.Slug, SKU: req.SKU, Name: req.Name, ShortDescription: req.ShortDescription, Description: req.Description, DatasheetURL: req.DatasheetURL, ImageURL: req.ImageURL, Status: req.Status, Featured: req.Featured, Specs: toSpecs(req.Specs), Media: toMedia(req.Media), Compatibility: toCompatibility(req.Compatibility), ReplaceSpecs: req.ReplaceSpecs, ReplaceMedia: req.ReplaceMedia, ReplaceCompatibility: req.ReplaceCompatibility}, requestContext(c))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.uc.Catalog.ListCategories(c.Request.Context())
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req categoryCreateRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	item, err := h.uc.Catalog.CreateCategory(c.Request.Context(), application.CreateCategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description})
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) ListBrands(c *gin.Context) {
	items, err := h.uc.Catalog.ListBrands(c.Request.Context())
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) CreateBrand(c *gin.Context) {
	var req brandCreateRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	item, err := h.uc.Catalog.CreateBrand(c.Request.Context(), application.CreateBrandInput{Name: req.Name, Slug: req.Slug, Description: req.Description, CountryCode: req.CountryCode})
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func bindJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil { return apperr.Invalid("invalid_json", "the request body is invalid").WithCause(err) }
	if err := pkgvalidator.Struct(target); err != nil { return err }
	return nil
}

func requestContext(c *gin.Context) application.RequestContext {
	var accountID uuid.UUID
	if raw := middleware.AccountID(c); raw != "" { accountID, _ = uuid.Parse(raw) }
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c), AccountID: accountID}
}

func toSpecs(items []specRequest) []domain.ProductSpec {
	out := make([]domain.ProductSpec, 0, len(items))
	for _, item := range items { out = append(out, domain.ProductSpec{Key: item.Key, Value: item.Value}) }
	return out
}
func toMedia(items []mediaRequest) []domain.ProductMedia {
	out := make([]domain.ProductMedia, 0, len(items))
	for _, item := range items { out = append(out, domain.ProductMedia{URL: item.URL, Type: item.Type, SortOrder: item.SortOrder}) }
	return out
}
func toCompatibility(items []compatibilityRequest) []domain.CompatibilityRule {
	out := make([]domain.CompatibilityRule, 0, len(items))
	for _, item := range items { out = append(out, domain.CompatibilityRule{Kind: item.Kind, Value: item.Value}) }
	return out
}
