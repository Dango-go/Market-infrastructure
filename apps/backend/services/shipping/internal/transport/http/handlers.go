package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/shipping/internal/application"
)

type UseCases struct{ Shipping *application.ShippingUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListShipments(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Shipping.ListByAccountID(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) GetShipment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	shipmentID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_shipment_id", "the shipment id is not a valid UUID")); return }
	item, err := h.uc.Shipping.GetByID(c.Request.Context(), accountID, shipmentID)
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) CreateShipment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	var req createShipmentRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_order_id", "the order id is not a valid UUID")); return }
	eta, err := parseOptionalTime(req.Eta, "eta")
	if err != nil { httpx.Fail(c, err); return }
	view, err := h.uc.Shipping.Create(c.Request.Context(), accountID, application.CreateShipmentInput{OrderID: orderID, Carrier: req.Carrier, ServiceLevel: req.ServiceLevel, TrackingNumber: req.TrackingNumber, DestinationAddress: req.DestinationAddress, Eta: eta}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, view)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	shipmentID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_shipment_id", "the shipment id is not a valid UUID")); return }
	var req updateShipmentStatusRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	eta, err := parseOptionalTime(req.Eta, "eta")
	if err != nil { httpx.Fail(c, err); return }
	view, err := h.uc.Shipping.UpdateStatus(c.Request.Context(), accountID, shipmentID, application.UpdateShipmentStatusInput{Status: req.Status, Carrier: req.Carrier, TrackingNumber: req.TrackingNumber, Eta: eta}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, view)
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

func parseOptionalTime(raw *string, field string) (*time.Time, error) {
	if raw == nil || *raw == "" { return nil, nil }
	parsed, err := time.Parse(time.RFC3339, *raw)
	if err != nil { return nil, apperr.Invalid("invalid_"+field, field+" must be RFC3339") }
	return &parsed, nil
}

func requireAccountID(c *gin.Context) (uuid.UUID, bool) {
	raw := middleware.AccountID(c)
	if raw == "" {
		httpx.Fail(c, apperr.Unauthorized("authentication_required", "authentication is required"))
		return uuid.Nil, false
	}
	accountID, err := uuid.Parse(raw)
	if err != nil {
		httpx.Fail(c, apperr.Unauthorized("invalid_account_id", "authenticated account id is invalid"))
		return uuid.Nil, false
	}
	return accountID, true
}

func requestContext(c *gin.Context, accountID uuid.UUID) application.RequestContext {
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c), AccountID: accountID}
}
