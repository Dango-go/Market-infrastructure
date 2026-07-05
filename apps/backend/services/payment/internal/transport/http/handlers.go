package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/payment/internal/application"
)

type UseCases struct{ Payment *application.PaymentUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListPayments(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Payment.ListByAccountID(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) GetPayment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_payment_id", "the payment id is not a valid UUID")); return }
	item, err := h.uc.Payment.GetByID(c.Request.Context(), accountID, paymentID)
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) CreatePayment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	var req createPaymentRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_order_id", "the order id is not a valid UUID")); return }
	view, err := h.uc.Payment.Create(c.Request.Context(), accountID, application.CreatePaymentInput{OrderID: orderID, Provider: req.Provider, Method: req.Method, Currency: req.Currency, AmountCents: req.AmountCents, TransactionRef: req.TransactionRef}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, view)
}

func (h *Handler) ConfirmPayment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_payment_id", "the payment id is not a valid UUID")); return }
	var req confirmPaymentRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	view, err := h.uc.Payment.Confirm(c.Request.Context(), accountID, paymentID, application.ConfirmPaymentInput{TransactionRef: req.TransactionRef}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, view)
}

func (h *Handler) FailPayment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_payment_id", "the payment id is not a valid UUID")); return }
	var req failPaymentRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	view, err := h.uc.Payment.Fail(c.Request.Context(), accountID, paymentID, application.FailPaymentInput{Reason: req.Reason, TransactionRef: req.TransactionRef}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, view)
}

func (h *Handler) RefundPayment(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_payment_id", "the payment id is not a valid UUID")); return }
	view, err := h.uc.Payment.Refund(c.Request.Context(), accountID, paymentID, requestContext(c, accountID))
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
