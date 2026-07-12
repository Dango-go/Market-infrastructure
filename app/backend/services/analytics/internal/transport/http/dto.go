package http

type trackEventRequest struct {
	SessionID string  `json:"session_id" validate:"required,max=120"`
	ProductID *string `json:"product_id" validate:"omitempty"`
	EventType string  `json:"event_type" validate:"required,max=40"`
	Path      string  `json:"path" validate:"omitempty,max=500"`
	Referrer  string  `json:"referrer" validate:"omitempty,max=500"`
	Query     string  `json:"query" validate:"omitempty,max=255"`
	UserAgent string  `json:"user_agent" validate:"omitempty,max=500"`
}
