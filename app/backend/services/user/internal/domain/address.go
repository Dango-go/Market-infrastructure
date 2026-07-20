package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Address is a shipping/billing delivery address.
type Address struct {
	ID                uuid.UUID
	AccountID         uuid.UUID
	Label             string
	RecipientName     string
	Line1             string
	Line2             string
	City              string
	Region            string
	PostalCode        string
	CountryCode       string
	Phone             string
	IsDefaultShipping bool
	IsDefaultBilling  bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

func NewAddress(id, accountID uuid.UUID, label, recipientName, line1, line2, city, region, postalCode, countryCode, phone string, isDefaultShipping, isDefaultBilling bool, now time.Time) *Address {
	return &Address{
		ID:                id,
		AccountID:         accountID,
		Label:             strings.TrimSpace(label),
		RecipientName:     strings.TrimSpace(recipientName),
		Line1:             strings.TrimSpace(line1),
		Line2:             strings.TrimSpace(line2),
		City:              strings.TrimSpace(city),
		Region:            strings.TrimSpace(region),
		PostalCode:        strings.TrimSpace(postalCode),
		CountryCode:       strings.ToUpper(strings.TrimSpace(countryCode)),
		Phone:             strings.TrimSpace(phone),
		IsDefaultShipping: isDefaultShipping,
		IsDefaultBilling:  isDefaultBilling,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (a *Address) IsDeleted() bool { return a.DeletedAt != nil }

func (a *Address) Update(label, recipientName, line1, line2, city, region, postalCode, countryCode, phone *string, isDefaultShipping, isDefaultBilling *bool, now time.Time) {
	if label != nil {
		a.Label = strings.TrimSpace(*label)
	}
	if recipientName != nil {
		a.RecipientName = strings.TrimSpace(*recipientName)
	}
	if line1 != nil {
		a.Line1 = strings.TrimSpace(*line1)
	}
	if line2 != nil {
		a.Line2 = strings.TrimSpace(*line2)
	}
	if city != nil {
		a.City = strings.TrimSpace(*city)
	}
	if region != nil {
		a.Region = strings.TrimSpace(*region)
	}
	if postalCode != nil {
		a.PostalCode = strings.TrimSpace(*postalCode)
	}
	if countryCode != nil {
		a.CountryCode = strings.ToUpper(strings.TrimSpace(*countryCode))
	}
	if phone != nil {
		a.Phone = strings.TrimSpace(*phone)
	}
	if isDefaultShipping != nil {
		a.IsDefaultShipping = *isDefaultShipping
	}
	if isDefaultBilling != nil {
		a.IsDefaultBilling = *isDefaultBilling
	}
	a.UpdatedAt = now
}

func (a *Address) GetID() uuid.UUID             { return a.ID }
func (a *Address) GetAccountID() uuid.UUID      { return a.AccountID }
func (a *Address) GetLabel() string             { return a.Label }
func (a *Address) GetRecipientName() string     { return a.RecipientName }
func (a *Address) GetLine1() string             { return a.Line1 }
func (a *Address) GetLine2() string             { return a.Line2 }
func (a *Address) GetCity() string              { return a.City }
func (a *Address) GetRegion() string            { return a.Region }
func (a *Address) GetPostalCode() string        { return a.PostalCode }
func (a *Address) GetCountryCode() string       { return a.CountryCode }
func (a *Address) GetPhone() string             { return a.Phone }
func (a *Address) GetDefaultShipping() bool     { return a.IsDefaultShipping }
func (a *Address) GetDefaultBilling() bool      { return a.IsDefaultBilling }
func (a *Address) GetCreatedAt() time.Time       { return a.CreatedAt }
func (a *Address) GetUpdatedAt() time.Time       { return a.UpdatedAt }
