// Package httpx contains transport helpers shared by all Gin services: a uniform
// response envelope, application-error -> HTTP mapping, and pagination plumbing.
package httpx

import (
	"net/http"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// errorBody is the stable error envelope returned to clients.
type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// statusFor maps an application error Kind to an HTTP status code. This is the single
// place where transport semantics meet the domain taxonomy.
func statusFor(kind apperr.Kind) int {
	switch kind {
	case apperr.KindInvalid:
		return http.StatusUnprocessableEntity
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindRateLimited:
		return http.StatusTooManyRequests
	case apperr.KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// OK writes a 200 with the given payload.
func OK(c *gin.Context, payload any) { c.JSON(http.StatusOK, payload) }

// Created writes a 201 with the given payload.
func Created(c *gin.Context, payload any) { c.JSON(http.StatusCreated, payload) }

// NoContent writes a 204.
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// Page writes a 200 with a paginated envelope.
func Page[T any](c *gin.Context, items []T, p Pagination) {
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "pagination": p})
}

// Fail converts any error into the standard error envelope, logging internals and
// hiding their cause from the client. It always terminates the handler chain.
func Fail(c *gin.Context, err error) {
	ae := apperr.From(err)
	status := statusFor(ae.Kind)

	if status >= http.StatusInternalServerError {
		logger.FromContext(c.Request.Context()).Error("request failed",
			logger.String("code", ae.Code), logger.Error(ae))
		c.AbortWithStatusJSON(status, errorBody{Error: errorPayload{
			Code:    ae.Code,
			Message: "an internal error occurred",
		}})
		return
	}

	c.AbortWithStatusJSON(status, errorBody{Error: errorPayload{
		Code:    ae.Code,
		Message: ae.Message,
		Details: ae.Details,
	}})
}
