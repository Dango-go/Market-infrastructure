package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/pricing/internal/application"
)

type UseCases struct{ Pricing *application.PricingUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) GetPrice(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	item, err := h.uc.Pricing.GetPrice(c.Request.Context(), productID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *Handler) ListPrices(c *gin.Context) {
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Pricing.ListPrices(c.Request.Context(), page.Limit(), page.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) UpsertPrice(c *gin.Context) {
	var req upsertPriceRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	item, err := h.uc.Pricing.UpsertPrice(c.Request.Context(), application.UpsertPriceInput{
		ProductID:      productID,
		Currency:       req.Currency,
		AmountCents:    req.AmountCents,
		CompareAtCents: req.CompareAtCents,
		Active:         req.Active,
	}, requestContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *Handler) ListPromotions(c *gin.Context) {
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Pricing.ListPromotions(c.Request.Context(), page.Limit(), page.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) CreatePromotion(c *gin.Context) {
	var req createPromotionRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_starts_at", "starts_at must be RFC3339"))
		return
	}
	var endsAt *time.Time
	if req.EndsAt != nil && *req.EndsAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			httpx.Fail(c, apperr.Invalid("invalid_ends_at", "ends_at must be RFC3339"))
			return
		}
		endsAt = &parsed
	}
	item, err := h.uc.Pricing.CreatePromotion(c.Request.Context(), application.CreatePromotionInput{
		Name:         req.Name,
		Code:         req.Code,
		DiscountType: req.DiscountType,
		ValueCents:   req.ValueCents,
		PercentOff:   req.PercentOff,
		Active:       req.Active,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
	}, requestContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, item)
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

func requestContext(c *gin.Context) application.RequestContext {
	var accountID uuid.UUID
	if raw := middleware.AccountID(c); raw != "" {
		accountID, _ = uuid.Parse(raw)
	}
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c), AccountID: accountID}
}
