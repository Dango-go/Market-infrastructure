package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/review/internal/application"
)

type UseCases struct{ Review *application.ReviewUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListByProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Query("product_id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Review.ListByProductID(c.Request.Context(), productID, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) GetSummary(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	item, err := h.uc.Review.GetSummaryByProductID(c.Request.Context(), productID)
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) ListMine(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Review.ListByAccountID(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) CreateReview(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	var req createReviewRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID"))
		return
	}
	item, err := h.uc.Review.Create(c.Request.Context(), accountID, application.CreateReviewInput{ProductID: productID, Rating: req.Rating, Title: req.Title, Body: req.Body})
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) UpdateReview(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_review_id", "the review id is not a valid UUID"))
		return
	}
	var req updateReviewRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	item, err := h.uc.Review.Update(c.Request.Context(), accountID, reviewID, application.UpdateReviewInput{Rating: req.Rating, Title: req.Title, Body: req.Body})
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) DeleteReview(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_review_id", "the review id is not a valid UUID"))
		return
	}
	if err := h.uc.Review.Delete(c.Request.Context(), accountID, reviewID); err != nil { httpx.Fail(c, err); return }
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
