package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventPageView      EventType = "page_view"
	EventProductView   EventType = "product_view"
	EventSearch        EventType = "search"
	EventAddToCart     EventType = "add_to_cart"
	EventBeginCheckout EventType = "begin_checkout"
	EventPurchase      EventType = "purchase"
)

type Event struct {
	ID         uuid.UUID
	AccountID  *uuid.UUID
	SessionID  string
	ProductID  *uuid.UUID
	EventType  EventType
	Path       string
	Referrer   string
	Query      string
	UserAgent  string
	CreatedAt  time.Time
}

type Overview struct {
	Days              int32
	TotalEvents       int64
	UniqueSessions    int64
	ProductViews      int64
	Searches          int64
	AddToCarts        int64
	Checkouts         int64
	Purchases         int64
}

type TopProduct struct {
	ProductID       uuid.UUID
	Views           int64
	AddToCarts      int64
	Purchases       int64
	ConversionScore float64
}

func NewEvent(id uuid.UUID, accountID *uuid.UUID, sessionID string, productID *uuid.UUID, eventType EventType, path, referrer, query, userAgent string, now time.Time) *Event {
	return &Event{ID: id, AccountID: accountID, SessionID: sessionID, ProductID: productID, EventType: eventType, Path: path, Referrer: referrer, Query: query, UserAgent: userAgent, CreatedAt: now}
}
