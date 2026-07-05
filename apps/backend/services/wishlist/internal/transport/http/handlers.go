package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	"github.com/embedded-market/backend/services/wishlist/internal/application"
)

type UseCases struct{ Wishlist *application.WishlistUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListWishlist(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Wishlist.List(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	views := make([]wishlistResponse, 0, len(items))
	for _, item := range items {
		views = append(views, wishlistResponse{AccountID: item.AccountID.String(), ProductID: item.ProductID.String(), AddedAt: item.AddedAt})
	}
	httpx.Page(c, views, page.BuildPagination(total))
}

func (h *Handler) AddWishlist(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	if err := h.uc.Wishlist.Add(c.Request.Context(), accountID, productID, requestContext(c)); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, gin.H{"status": "added"})
}

func (h *Handler) RemoveWishlist(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	if err := h.uc.Wishlist.Remove(c.Request.Context(), accountID, productID, requestContext(c)); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.NoContent(c)
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

func requestContext(c *gin.Context) application.RequestContext {
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c)}
}
