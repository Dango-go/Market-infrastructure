package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/notification/internal/application"
)

type UseCases struct{ Notification *application.NotificationUseCase }

type Handler struct{ uc UseCases }

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListTemplates(c *gin.Context) {
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Notification.ListTemplates(c.Request.Context(), page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req createTemplateRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	item, err := h.uc.Notification.CreateTemplate(c.Request.Context(), application.CreateTemplateInput{Code: req.Code, Channel: req.Channel, Subject: req.Subject, Body: req.Body})
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) ListNotifications(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Notification.ListNotifications(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) GetNotification(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_notification_id", "the notification id is not a valid UUID")); return }
	item, err := h.uc.Notification.GetNotification(c.Request.Context(), accountID, notificationID)
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) CreateNotification(c *gin.Context) {
	var req createNotificationRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_account_id", "the account id is not a valid UUID")); return }
	var templateID *uuid.UUID
	if req.TemplateID != nil && *req.TemplateID != "" {
		parsed, err := uuid.Parse(*req.TemplateID)
		if err != nil { httpx.Fail(c, apperr.Invalid("invalid_template_id", "the template id is not a valid UUID")); return }
		templateID = &parsed
	}
	item, err := h.uc.Notification.CreateNotification(c.Request.Context(), application.CreateNotificationInput{AccountID: accountID, TemplateID: templateID, Channel: req.Channel, Subject: req.Subject, Body: req.Body, MetadataJSON: req.MetadataJSON}, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) MarkSent(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_notification_id", "the notification id is not a valid UUID")); return }
	item, err := h.uc.Notification.MarkSent(c.Request.Context(), accountID, notificationID, requestContext(c, accountID))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) MarkRead(c *gin.Context) {
	accountID, ok := requireAccountID(c)
	if !ok { return }
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_notification_id", "the notification id is not a valid UUID")); return }
	item, err := h.uc.Notification.MarkRead(c.Request.Context(), accountID, notificationID)
	if err != nil { httpx.Fail(c, err); return }
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
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c), AccountID: accountID}
}
