package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/user/internal/application"
)

// UseCases bundles the user-service use cases injected into transport.
type UseCases struct {
	Bootstrap   *application.BootstrapUseCase
	Profile     *application.ProfileUseCase
	Preferences *application.PreferencesUseCase
	Addresses   *application.AddressUseCase
}

type Handler struct {
	uc UseCases
}

func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) Me(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	view, err := h.uc.Profile.Get(c.Request.Context(), accountID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toProfileResponse(view))
}

func (h *Handler) UpdateMe(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	var req profileUpdateRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	view, err := h.uc.Profile.Update(c.Request.Context(), accountID, application.UpdateProfileInput{
		DisplayName:    req.DisplayName,
		Bio:            req.Bio,
		Phone:          req.Phone,
		AvatarURL:      req.AvatarURL,
		Locale:         req.Locale,
		Timezone:       req.Timezone,
		MarketingOptIn: req.MarketingOptIn,
	}, requestContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toProfileResponse(view))
}

func (h *Handler) GetPreferences(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	view, err := h.uc.Preferences.Get(c.Request.Context(), accountID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toPreferencesResponse(view))
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	var req preferencesUpdateRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	view, err := h.uc.Preferences.Update(c.Request.Context(), accountID, application.UpdatePreferencesInput{
		Currency:          req.Currency,
		Language:          req.Language,
		EmailNotifications: req.EmailNotifications,
		SMSNotifications:  req.SMSNotifications,
		PushNotifications:  req.PushNotifications,
		MarketingOptIn:    req.MarketingOptIn,
	}, requestContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toPreferencesResponse(view))
}

func (h *Handler) ListAddresses(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Addresses.List(c.Request.Context(), accountID, page.Limit(), page.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	views := make([]addressResponse, 0, len(items))
	for _, item := range items {
		views = append(views, toAddressResponse(item))
	}
	httpx.Page(c, views, page.BuildPagination(total))
}

func (h *Handler) CreateAddress(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	var req addressCreateRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	view, err := h.uc.Addresses.Create(c.Request.Context(), accountID, application.CreateAddressInput{
		Label:             req.Label,
		RecipientName:     req.RecipientName,
		Line1:             req.Line1,
		Line2:             req.Line2,
		City:              req.City,
		Region:            req.Region,
		PostalCode:        req.PostalCode,
		CountryCode:       req.CountryCode,
		Phone:             req.Phone,
		IsDefaultShipping: req.IsDefaultShipping,
		IsDefaultBilling:  req.IsDefaultBilling,
	}, requestContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, toAddressResponse(view))
}

func (h *Handler) UpdateAddress(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_address_id", "the address id is not a valid UUID"))
		return
	}
	var req addressUpdateRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	view, err := h.uc.Addresses.Update(c.Request.Context(), accountID, addressID, application.UpdateAddressInput{
		Label:             req.Label,
		RecipientName:     req.RecipientName,
		Line1:             req.Line1,
		Line2:             req.Line2,
		City:              req.City,
		Region:            req.Region,
		PostalCode:        req.PostalCode,
		CountryCode:       req.CountryCode,
		Phone:             req.Phone,
		IsDefaultShipping: req.IsDefaultShipping,
		IsDefaultBilling:  req.IsDefaultBilling,
	}, requestContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toAddressResponse(view))
}

func (h *Handler) DeleteAddress(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_address_id", "the address id is not a valid UUID"))
		return
	}
	if err := h.uc.Addresses.Delete(c.Request.Context(), accountID, addressID, requestContext(c)); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) SetDefaultShipping(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_address_id", "the address id is not a valid UUID"))
		return
	}
	if err := h.uc.Addresses.SetDefaultShipping(c.Request.Context(), accountID, addressID, requestContext(c)); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) SetDefaultBilling(c *gin.Context) {
	accountID, ok := currentAccountID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("missing_account_id", "authenticated account id is missing"))
		return
	}
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_address_id", "the address id is not a valid UUID"))
		return
	}
	if err := h.uc.Addresses.SetDefaultBilling(c.Request.Context(), accountID, addressID, requestContext(c)); err != nil {
		httpx.Fail(c, err)
		return
	}
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

func currentAccountID(c *gin.Context) (uuid.UUID, bool) {
	raw := middleware.AccountID(c)
	if raw == "" {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, false
	}
	return id, true
}

func requestContext(c *gin.Context) application.RequestContext {
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c)}
}

func toProfileResponse(v application.ProfileView) profileResponse {
	return profileResponse{
		AccountID:      v.AccountID.String(),
		Email:          v.Email,
		Username:       v.Username,
		DisplayName:    v.DisplayName,
		Bio:            v.Bio,
		Phone:          v.Phone,
		AvatarURL:      v.AvatarURL,
		Locale:         v.Locale,
		Timezone:       v.Timezone,
		MarketingOptIn: v.MarketingOptIn,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toPreferencesResponse(v application.PreferencesView) preferencesResponse {
	return preferencesResponse{
		AccountID:         v.AccountID.String(),
		Currency:          v.Currency,
		Language:          v.Language,
		EmailNotifications: v.EmailNotifications,
		SMSNotifications:  v.SMSNotifications,
		PushNotifications:  v.PushNotifications,
		MarketingOptIn:    v.MarketingOptIn,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}

func toAddressResponse(v application.AddressView) addressResponse {
	return addressResponse{
		ID:                v.ID.String(),
		AccountID:         v.AccountID.String(),
		Label:             v.Label,
		RecipientName:     v.RecipientName,
		Line1:             v.Line1,
		Line2:             v.Line2,
		City:              v.City,
		Region:            v.Region,
		PostalCode:        v.PostalCode,
		CountryCode:       v.CountryCode,
		Phone:             v.Phone,
		IsDefaultShipping: v.IsDefaultShipping,
		IsDefaultBilling:  v.IsDefaultBilling,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}
