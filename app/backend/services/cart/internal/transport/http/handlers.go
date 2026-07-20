package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/cart/internal/application"
)

type UseCases struct{ Cart *application.CartUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) GetActiveCart(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok {
		return
	}
	item, err := h.uc.Cart.GetOrCreateActive(c.Request.Context(), accountID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *Handler) AddItem(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok {
		return
	}
	var req addItemRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	item, err := h.uc.Cart.AddItem(c.Request.Context(), accountID, application.AddItemInput{
		ProductID:      productID,
		Quantity:       req.Quantity,
		UnitPriceCents: req.UnitPriceCents,
	}, requestContext(c, accountID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok {
		return
	}
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	var req updateItemRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	item, err := h.uc.Cart.UpdateItem(c.Request.Context(), accountID, productID, req.Quantity, requestContext(c, accountID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *Handler) ClearCart(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok {
		return
	}
	item, err := h.uc.Cart.Clear(c.Request.Context(), accountID, requestContext(c, accountID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *Handler) Checkout(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok {
		return
	}
	item, err := h.uc.Cart.Checkout(c.Request.Context(), accountID, requestContext(c, accountID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, item)
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
	return application.RequestContext{
		CorrelationID: middleware.CorrelationID(c),
		AccountID:     accountID,
	}
}
