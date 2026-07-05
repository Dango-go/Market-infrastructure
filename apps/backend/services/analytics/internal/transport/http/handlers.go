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
	"github.com/embedded-market/backend/services/analytics/internal/application"
)

type UseCases struct{ Analytics *application.AnalyticsUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) TrackEvent(c *gin.Context) {
	var req trackEventRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	var productID *uuid.UUID
	if req.ProductID != nil && strings.TrimSpace(*req.ProductID) != "" {
		parsed, err := uuid.Parse(*req.ProductID)
		if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
		productID = &parsed
	}
	var accountID *uuid.UUID
	if raw := strings.TrimSpace(middleware.AccountID(c)); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err == nil { accountID = &parsed }
	}
	if err := h.uc.Analytics.TrackEvent(c.Request.Context(), application.TrackEventInput{AccountID: accountID, SessionID: req.SessionID, ProductID: productID, EventType: req.EventType, Path: req.Path, Referrer: req.Referrer, Query: req.Query, UserAgent: req.UserAgent}); err != nil { httpx.Fail(c, err); return }
	httpx.NoContent(c)
}

func (h *Handler) Overview(c *gin.Context) {
	days, err := parseDays(c.Query("days"), 7)
	if err != nil { httpx.Fail(c, err); return }
	item, err := h.uc.Analytics.Overview(c.Request.Context(), int32(days))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) TopProducts(c *gin.Context) {
	days, err := parseDays(c.Query("days"), 7)
	if err != nil { httpx.Fail(c, err); return }
	limit, err := parseLimit(c.Query("limit"), 10, 50)
	if err != nil { httpx.Fail(c, err); return }
	items, err := h.uc.Analytics.TopProducts(c.Request.Context(), int32(days), int32(limit))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, gin.H{"items": items})
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

func parseDays(raw string, fallback int) (int, error) {
	if strings.TrimSpace(raw) == "" { return fallback, nil }
	value, err := strconv.Atoi(raw)
	if err != nil { return 0, apperr.Invalid("invalid_days", "days must be an integer") }
	return value, nil
}

func parseLimit(raw string, fallback, max int) (int, error) {
	if strings.TrimSpace(raw) == "" { return fallback, nil }
	value, err := strconv.Atoi(raw)
	if err != nil { return 0, apperr.Invalid("invalid_limit", "limit must be an integer") }
	if value < 1 { return 0, apperr.Invalid("invalid_limit", "limit must be greater than zero") }
	if value > max { value = max }
	return value, nil
}
