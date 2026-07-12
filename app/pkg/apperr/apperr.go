// Package apperr defines a transport-agnostic error taxonomy shared by every service.
//
// Domain and application layers return *Error values carrying a Kind. The transport
// layer (and only the transport layer) maps a Kind to an HTTP status code, keeping the
// inner layers free of any knowledge about HTTP or Gin.
package apperr

import (
	"errors"
	"fmt"
)

// Kind classifies an error independently of any transport.
type Kind string

const (
	KindInvalid      Kind = "invalid"       // malformed input / failed validation
	KindNotFound     Kind = "not_found"     // requested resource does not exist
	KindConflict     Kind = "conflict"      // uniqueness / state conflict
	KindUnauthorized Kind = "unauthorized"  // missing or invalid credentials
	KindForbidden    Kind = "forbidden"     // authenticated but not permitted
	KindRateLimited  Kind = "rate_limited"  // too many requests
	KindUnavailable  Kind = "unavailable"   // a dependency is down
	KindInternal     Kind = "internal"      // unexpected server failure
)

// Error is the canonical application error. It carries a machine-readable Kind, a stable
// Code for clients, a human message, optional field-level Details, and a wrapped cause.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Details map[string]string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// WithDetails attaches field-level details (e.g. validation failures) and returns e.
func (e *Error) WithDetails(d map[string]string) *Error {
	e.Details = d
	return e
}

// WithCause wraps an underlying error for logging without leaking it to clients.
func (e *Error) WithCause(cause error) *Error {
	e.cause = cause
	return e
}

func newf(kind Kind, code, format string, args ...any) *Error {
	return &Error{Kind: kind, Code: code, Message: fmt.Sprintf(format, args...)}
}

// Constructors — one per Kind. Code is a stable, client-facing identifier.

func Invalid(code, format string, args ...any) *Error      { return newf(KindInvalid, code, format, args...) }
func NotFound(code, format string, args ...any) *Error     { return newf(KindNotFound, code, format, args...) }
func Conflict(code, format string, args ...any) *Error     { return newf(KindConflict, code, format, args...) }
func Unauthorized(code, format string, args ...any) *Error { return newf(KindUnauthorized, code, format, args...) }
func Forbidden(code, format string, args ...any) *Error    { return newf(KindForbidden, code, format, args...) }
func RateLimited(code, format string, args ...any) *Error  { return newf(KindRateLimited, code, format, args...) }
func Unavailable(code, format string, args ...any) *Error  { return newf(KindUnavailable, code, format, args...) }
func Internal(code, format string, args ...any) *Error     { return newf(KindInternal, code, format, args...) }

// From normalizes an arbitrary error into an *Error, defaulting unknown errors to
// KindInternal so they are never silently exposed to clients.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return Internal("internal_error", "an unexpected error occurred").WithCause(err)
}

// KindOf extracts the Kind of any error, treating non-application errors as internal.
func KindOf(err error) Kind {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Kind
	}
	return KindInternal
}
