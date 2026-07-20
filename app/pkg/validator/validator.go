// Package validator wraps go-playground/validator and translates validation failures
// into apperr.Error values with field-level details, so every service validates and
// reports request errors identically.
package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/go-playground/validator/v10"
)

var (
	once sync.Once
	v    *validator.Validate
)

func instance() *validator.Validate {
	once.Do(func() {
		v = validator.New(validator.WithRequiredStructEnabled())
		// Report the JSON tag name rather than the Go struct field name.
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return fld.Name
			}
			return name
		})
	})
	return v
}

// Struct validates s and returns an apperr.Invalid with per-field messages on failure.
func Struct(s any) error {
	if err := instance().Struct(s); err != nil {
		var ve validator.ValidationErrors
		if ok := asValidationErrors(err, &ve); ok {
			details := make(map[string]string, len(ve))
			for _, fe := range ve {
				details[fe.Field()] = message(fe)
			}
			return apperr.Invalid("validation_failed", "request validation failed").WithDetails(details)
		}
		return apperr.Invalid("validation_failed", "request validation failed").WithCause(err)
	}
	return nil
}

func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "uuid", "uuid4", "uuid7":
		return "must be a valid UUID"
	case "alphanum":
		return "must contain only letters and digits"
	case "url":
		return "must be a valid URL"
	default:
		return "is invalid"
	}
}
