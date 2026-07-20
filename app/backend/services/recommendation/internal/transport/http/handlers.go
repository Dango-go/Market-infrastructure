package http

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/recommendation/internal/application"
)

type UseCases struct{ Recommendation *application.RecommendationUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) Related(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
	limit, err := parseLimit(c.Query("limit"), 8, 24)
	if err != nil { httpx.Fail(c, err); return }
	items, err := h.uc.Recommendation.Related(c.Request.Context(), productID, int32(limit))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) Trending(c *gin.Context) {
	limit, err := parseLimit(c.Query("limit"), 8, 24)
	if err != nil { httpx.Fail(c, err); return }
	items, err := h.uc.Recommendation.Trending(c.Request.Context(), int32(limit))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) UpsertProfile(c *gin.Context) {
	var req upsertProfileRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	productID, err := uuid.Parse(req.ProductID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
	item, err := h.uc.Recommendation.UpsertProfile(c.Request.Context(), application.UpsertProfileInput{ProductID: productID, Slug: req.Slug, CategorySlug: req.CategorySlug, BrandSlug: req.BrandSlug, Tags: req.Tags, PriceCents: req.PriceCents, Available: req.Available})
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) TrackEvent(c *gin.Context) {
	var req trackEventRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	productID, err := uuid.Parse(req.ProductID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
	var accountID *uuid.UUID
	if raw := strings.TrimSpace(middleware.AccountID(c)); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err == nil {
			accountID = &parsed
		}
	}
	if err := h.uc.Recommendation.TrackEvent(c.Request.Context(), application.TrackEventInput{ProductID: productID, AccountID: accountID, Type: req.Type}); err != nil { httpx.Fail(c, err); return }
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

func parseLimit(raw string, fallback, max int) (int, error) {
	if strings.TrimSpace(raw) == "" { return fallback, nil }
	value, err := strconv.Atoi(raw)
	if err != nil { return 0, apperr.Invalid("invalid_limit", "limit must be an integer") }
	if value < 1 { return 0, apperr.Invalid("invalid_limit", "limit must be greater than zero") }
	if value > max { value = max }
	return value, nil
}
