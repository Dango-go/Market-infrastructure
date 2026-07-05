package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
}

type BootstrapInput struct {
	EventID       uuid.UUID
	CorrelationID string
	AccountID     uuid.UUID
	Email         string
	Username      string
	RegisteredAt  time.Time
}

type ProfileView struct {
	AccountID      uuid.UUID `json:"account_id"`
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

type PreferencesView struct {
	AccountID         uuid.UUID `json:"account_id"`
	Currency          string    `json:"currency"`
	Language          string    `json:"language"`
	EmailNotifications bool      `json:"email_notifications"`
	SMSNotifications  bool      `json:"sms_notifications"`
	PushNotifications bool      `json:"push_notifications"`
	MarketingOptIn    bool      `json:"marketing_opt_in"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AddressView struct {
	ID                uuid.UUID `json:"id"`
	AccountID         uuid.UUID `json:"account_id"`
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

