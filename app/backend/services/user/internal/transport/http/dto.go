package http

import (
	"time"
)

// --- Requests -------------------------------------------------------------------------

type profileUpdateRequest struct {
	DisplayName    *string `json:"display_name" validate:"omitempty,max=100"`
	Bio            *string `json:"bio" validate:"omitempty,max=500"`
	Phone          *string `json:"phone" validate:"omitempty,max=32"`
	AvatarURL      *string `json:"avatar_url" validate:"omitempty,url,max=512"`
	Locale         *string `json:"locale" validate:"omitempty,max=20"`
	Timezone       *string `json:"timezone" validate:"omitempty,max=64"`
	MarketingOptIn *bool   `json:"marketing_opt_in"`
}

type addressCreateRequest struct {
	Label             string `json:"label" validate:"required,max=64"`
	RecipientName     string `json:"recipient_name" validate:"required,max=100"`
	Line1             string `json:"line1" validate:"required,max=120"`
	Line2             string `json:"line2" validate:"omitempty,max=120"`
	City              string `json:"city" validate:"required,max=80"`
	Region            string `json:"region" validate:"omitempty,max=80"`
	PostalCode        string `json:"postal_code" validate:"required,max=20"`
	CountryCode       string `json:"country_code" validate:"required,len=2"`
	Phone             string `json:"phone" validate:"omitempty,max=32"`
	IsDefaultShipping bool   `json:"is_default_shipping"`
	IsDefaultBilling  bool   `json:"is_default_billing"`
}

type addressUpdateRequest struct {
	Label             *string `json:"label" validate:"omitempty,max=64"`
	RecipientName     *string `json:"recipient_name" validate:"omitempty,max=100"`
	Line1             *string `json:"line1" validate:"omitempty,max=120"`
	Line2             *string `json:"line2" validate:"omitempty,max=120"`
	City              *string `json:"city" validate:"omitempty,max=80"`
	Region            *string `json:"region" validate:"omitempty,max=80"`
	PostalCode        *string `json:"postal_code" validate:"omitempty,max=20"`
	CountryCode       *string `json:"country_code" validate:"omitempty,len=2"`
	Phone             *string `json:"phone" validate:"omitempty,max=32"`
	IsDefaultShipping *bool   `json:"is_default_shipping"`
	IsDefaultBilling  *bool   `json:"is_default_billing"`
}

type preferencesUpdateRequest struct {
	Currency          *string `json:"currency" validate:"omitempty,max=8"`
	Language          *string `json:"language" validate:"omitempty,max=8"`
	EmailNotifications *bool   `json:"email_notifications"`
	SMSNotifications  *bool   `json:"sms_notifications"`
	PushNotifications *bool   `json:"push_notifications"`
	MarketingOptIn    *bool   `json:"marketing_opt_in"`
}

// --- Responses ------------------------------------------------------------------------

type profileResponse struct {
	AccountID      string    `json:"account_id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"display_name"`
	Bio            string    `json:"bio"`
	Phone          string    `json:"phone"`
	AvatarURL      string    `json:"avatar_url"`
	Locale         string    `json:"locale"`
	Timezone       string    `json:"timezone"`
	MarketingOptIn bool      `json:"marketing_opt_in"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type preferencesResponse struct {
	AccountID         string    `json:"account_id"`
	Currency          string    `json:"currency"`
	Language          string    `json:"language"`
	EmailNotifications bool      `json:"email_notifications"`
	SMSNotifications  bool      `json:"sms_notifications"`
	PushNotifications bool      `json:"push_notifications"`
	MarketingOptIn    bool      `json:"marketing_opt_in"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type addressResponse struct {
	ID                string    `json:"id"`
	AccountID         string    `json:"account_id"`
	Label             string    `json:"label"`
	RecipientName     string    `json:"recipient_name"`
	Line1             string    `json:"line1"`
	Line2             string    `json:"line2"`
	City              string    `json:"city"`
	Region            string    `json:"region"`
	PostalCode        string    `json:"postal_code"`
	CountryCode       string    `json:"country_code"`
	Phone             string    `json:"phone"`
	IsDefaultShipping bool      `json:"is_default_shipping"`
	IsDefaultBilling  bool      `json:"is_default_billing"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

