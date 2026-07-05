package http

import (
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "github.com/embedded-market/backend/pkg/apperr"
    "github.com/embedded-market/backend/pkg/httpx"
    pkgvalidator "github.com/embedded-market/backend/pkg/validator"
    "github.com/embedded-market/backend/services/search/internal/application"
    "github.com/embedded-market/backend/services/search/internal/domain"
)

type UseCases struct{ Search *application.SearchUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) Search(c *gin.Context) {
    page := httpx.ParsePageParams(c)
    filters, err := parseFilters(c)
    if err != nil { httpx.Fail(c, err); return }
    items, total, err := h.uc.Search.Search(c.Request.Context(), filters, page.Limit(), page.Offset())
    if err != nil { httpx.Fail(c, err); return }
    httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) Suggest(c *gin.Context) {
    limit, err := parseLimit(c.Query("limit"), 8, 20)
    if err != nil { httpx.Fail(c, err); return }
    items, err := h.uc.Search.Suggest(c.Request.Context(), c.Query("q"), int32(limit))
    if err != nil { httpx.Fail(c, err); return }
    httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) UpsertDocument(c *gin.Context) {
    var req upsertDocumentRequest
    if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }

    productID, err := uuid.Parse(req.ProductID)
    if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }

    var id *uuid.UUID
    if req.ID != nil && strings.TrimSpace(*req.ID) != "" {
        parsed, err := uuid.Parse(*req.ID)
        if err != nil { httpx.Fail(c, apperr.Invalid("invalid_document_id", "the document id is not a valid UUID")); return }
        id = &parsed
    }

    view, err := h.uc.Search.UpsertDocument(c.Request.Context(), application.UpsertDocumentInput{
        ID:               id,
        ProductID:        productID,
        Slug:             req.Slug,
        SKU:              req.SKU,
        Name:             req.Name,
        ShortDescription: req.ShortDescription,
        CategorySlug:     req.CategorySlug,
        BrandSlug:        req.BrandSlug,
        Tags:             req.Tags,
        SpecsText:        req.SpecsText,
        PriceCents:       req.PriceCents,
        Currency:         req.Currency,
        Available:        req.Available,
    })
    if err != nil { httpx.Fail(c, err); return }
    httpx.OK(c, view)
}

func (h *Handler) DeleteDocument(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil { httpx.Fail(c, apperr.Invalid("invalid_document_id", "the document id is not a valid UUID")); return }
    if err := h.uc.Search.DeleteDocument(c.Request.Context(), id); err != nil { httpx.Fail(c, err); return }
    httpx.NoContent(c)
}

func bindJSON(c *gin.Context, target any) error {
    if err := c.ShouldBindJSON(target); err != nil {
        return apperr.Invalid("invalid_json", "the request body is invalid").WithCause(err)
    }
    if err := pkgvalidator.Struct(target); err != nil {
        return err
    }
    return nil
}

func parseFilters(c *gin.Context) (domain.SearchFilters, error) {
    filters := domain.SearchFilters{
        Query:        strings.TrimSpace(c.Query("q")),
        CategorySlug: strings.TrimSpace(c.Query("category_slug")),
        BrandSlug:    strings.TrimSpace(c.Query("brand_slug")),
    }
    rawAvailable := strings.TrimSpace(c.Query("available"))
    if rawAvailable != "" {
        available, err := strconv.ParseBool(rawAvailable)
        if err != nil {
            return domain.SearchFilters{}, apperr.Invalid("invalid_available", "available must be a boolean")
        }
        filters.Available = &available
    }
    return filters, nil
}

func parseLimit(raw string, fallback, max int) (int, error) {
    if strings.TrimSpace(raw) == "" {
        return fallback, nil
    }
    value, err := strconv.Atoi(raw)
    if err != nil {
        return 0, apperr.Invalid("invalid_limit", "limit must be an integer")
    }
    if value < 1 {
        return 0, apperr.Invalid("invalid_limit", "limit must be greater than zero")
    }
    if value > max {
        value = max
    }
    return value, nil
}
