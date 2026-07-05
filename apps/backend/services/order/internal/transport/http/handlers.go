package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/order/internal/application"
)

type UseCases struct{ Order *application.OrderUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListOrders(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Order.ListByAccountID(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) GetOrder(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_order_id", "the order id is not a valid UUID")); return }
	item, err := h.uc.Order.GetByID(c.Request.Context(), accountID, orderID)
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	var req createOrderRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	var cartID *uuid.UUID
	if req.CartID != nil && *req.CartID != "" {
		parsed, err := uuid.Parse(*req.CartID)
		if err != nil { httpx.Fail(c, apperr.Invalid("invalid_cart_id", "the cart id is not a valid UUID")); return }
		cartID = &parsed
	}
	items := make([]application.CreateOrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
		items = append(items, application.CreateOrderItemInput{ProductID: productID, Quantity: item.Quantity, UnitPriceCents: item.UnitPriceCents})
	}
	view, err := h.uc.Order.Create(c.Request.Context(), accountID, application.CreateOrderInput{
		CartID:          cartID,
		Currency:        req.Currency,
		ShippingCents:   req.ShippingCents,
		DeliveryMethod:  req.DeliveryMethod,
		DeliveryAddress: req.DeliveryAddress,
		CustomerNote:    req.CustomerNote,
		Items:           items,
	}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, view)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_order_id", "the order id is not a valid UUID")); return }
	var req updateStatusRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	view, err := h.uc.Order.UpdateStatus(c.Request.Context(), accountID, orderID, application.UpdateStatusInput{Status: req.Status}, requestContext(c, accountID))
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
