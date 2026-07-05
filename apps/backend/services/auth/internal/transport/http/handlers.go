package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/auth/internal/application"
	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// UseCases bundles the auth use cases injected into the transport layer.
type UseCases struct {
	Register *application.RegisterUseCase
	Login    *application.LoginUseCase
	Refresh  *application.RefreshUseCase
	Logout   *application.LogoutUseCase
	Sessions *application.SessionsUseCase
	Account  *application.AccountUseCase
	OAuth    *application.OAuthUseCase
}

// HandlerConfig configures cookie behaviour and the OAuth flow.
type HandlerConfig struct {
	RefreshCookieName    string
	OAuthStateCookie     string
	CookieDomain         string
	CookieSecure         bool
	RefreshTTL           time.Duration
	OAuthStateTTL        time.Duration
}

// Handler holds the use cases and renders HTTP responses.
type Handler struct {
	uc  UseCases
	cfg HandlerConfig
}

// NewHandler builds the auth HTTP handler.
func NewHandler(uc UseCases, cfg HandlerConfig) *Handler {
	return &Handler{uc: uc, cfg: cfg}
}

// Register creates a new account and returns tokens.
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	result, err := h.uc.Register.Execute(c.Request.Context(), application.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}, h.reqContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	httpx.Created(c, toAuthResponse(result))
}

// Login authenticates an account and returns tokens.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := bindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	result, err := h.uc.Login.Execute(c.Request.Context(), application.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	}, h.reqContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	httpx.OK(c, toAuthResponse(result))
}

// Refresh rotates a refresh token, accepting it from the JSON body or the cookie.
func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	_ = c.ShouldBindJSON(&req) // body is optional when the cookie is present
	token := h.refreshTokenFrom(c, req.RefreshToken)

	result, err := h.uc.Refresh.Execute(c.Request.Context(), application.RefreshInput{RefreshToken: token}, h.reqContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	httpx.OK(c, toAuthResponse(result))
}

// Logout revokes the session for the presented refresh token (idempotent).
func (h *Handler) Logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)
	token := h.refreshTokenFrom(c, req.RefreshToken)

	if err := h.uc.Logout.Execute(c.Request.Context(), application.RefreshInput{RefreshToken: token}); err != nil {
		httpx.Fail(c, err)
		return
	}
	h.clearRefreshCookie(c)
	httpx.NoContent(c)
}

// Me returns the authenticated account.
func (h *Handler) Me(c *gin.Context) {
	accountID, err := currentAccountID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	view, err := h.uc.Account.Get(c.Request.Context(), accountID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toAccountResponse(view))
}

// ListSessions returns the account's active sessions (paginated).
func (h *Handler) ListSessions(c *gin.Context) {
	accountID, err := currentAccountID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	currentSession := currentSessionID(c)
	page := httpx.ParsePageParams(c)

	views, total, err := h.uc.Sessions.List(c.Request.Context(), accountID, currentSession, page.Limit(), page.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	items := make([]sessionResponse, 0, len(views))
	for _, v := range views {
		items = append(items, toSessionResponse(v))
	}
	httpx.Page(c, items, page.BuildPagination(total))
}

// RevokeSession revokes one of the account's sessions.
func (h *Handler) RevokeSession(c *gin.Context) {
	accountID, err := currentAccountID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid_session_id", "the session id is not a valid UUID"))
		return
	}
	if err := h.uc.Sessions.Revoke(c.Request.Context(), accountID, sessionID); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.NoContent(c)
}

// RevokeAllSessions revokes every session for the account ("sign out everywhere").
func (h *Handler) RevokeAllSessions(c *gin.Context) {
	accountID, err := currentAccountID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if _, err := h.uc.Sessions.RevokeAll(c.Request.Context(), accountID); err != nil {
		httpx.Fail(c, err)
		return
	}
	h.clearRefreshCookie(c)
	httpx.NoContent(c)
}

// OAuthBegin starts the provider authorization-code flow: it issues a CSRF state, stores
// it in a short-lived cookie, and redirects to the provider's consent screen.
func (h *Handler) OAuthBegin(c *gin.Context) {
	provider := domain.OAuthProvider(c.Param("provider"))
	state := uuid.NewString()

	url, err := h.uc.OAuth.Begin(provider, state)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.setStateCookie(c, state)
	c.Redirect(http.StatusFound, url)
}

// OAuthCallback completes the flow: it validates the CSRF state against the cookie,
// exchanges the code, and returns tokens.
func (h *Handler) OAuthCallback(c *gin.Context) {
	provider := domain.OAuthProvider(c.Param("provider"))

	queryState := c.Query("state")
	cookieState, _ := c.Cookie(h.cfg.OAuthStateCookie)
	if queryState == "" || cookieState == "" || queryState != cookieState {
		httpx.Fail(c, domain.ErrOAuthStateMismatch)
		return
	}
	h.clearStateCookie(c)

	code := c.Query("code")
	if code == "" {
		httpx.Fail(c, apperr.Invalid("missing_code", "the authorization code is missing"))
		return
	}

	result, err := h.uc.OAuth.Complete(c.Request.Context(), provider, code, h.reqContext(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	httpx.OK(c, toAuthResponse(result))
}

// --- helpers --------------------------------------------------------------------------

func bindJSON(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return apperr.Invalid("invalid_request_body", "the request body is not valid JSON").WithCause(err)
	}
	return pkgvalidator.Struct(dst)
}

func (h *Handler) reqContext(c *gin.Context) application.RequestContext {
	return application.RequestContext{
		CorrelationID: middleware.CorrelationID(c),
		UserAgent:     c.Request.UserAgent(),
		IPAddress:     c.ClientIP(),
	}
}

func (h *Handler) refreshTokenFrom(c *gin.Context, bodyToken string) string {
	if bodyToken != "" {
		return bodyToken
	}
	if cookie, err := c.Cookie(h.cfg.RefreshCookieName); err == nil {
		return cookie
	}
	return ""
}

func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cfg.RefreshCookieName, token, int(h.cfg.RefreshTTL.Seconds()), "/api/v1/auth", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cfg.RefreshCookieName, "", -1, "/api/v1/auth", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}

func (h *Handler) setStateCookie(c *gin.Context, state string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cfg.OAuthStateCookie, state, int(h.cfg.OAuthStateTTL.Seconds()), "/api/v1/auth/oauth", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}

func (h *Handler) clearStateCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cfg.OAuthStateCookie, "", -1, "/api/v1/auth/oauth", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}

func currentAccountID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(middleware.AccountID(c))
	if err != nil {
		return uuid.Nil, apperr.Unauthorized("invalid_subject", "the token subject is not a valid account id")
	}
	return id, nil
}

func currentSessionID(c *gin.Context) uuid.UUID {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return uuid.Nil
	}
	id, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return uuid.Nil
	}
	return id
}
