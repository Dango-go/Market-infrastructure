package http

type createTemplateRequest struct {
	Code    string `json:"code" validate:"required,max=120"`
	Channel string `json:"channel" validate:"required,oneof=email sms push in_app"`
	Subject string `json:"subject" validate:"required,max=255"`
	Body    string `json:"body" validate:"required,max=4000"`
}

type createNotificationRequest struct {
	AccountID    string  `json:"account_id" validate:"required,uuid"`
	TemplateID   *string `json:"template_id"`
	Channel      string  `json:"channel" validate:"required,oneof=email sms push in_app"`
	Subject      string  `json:"subject" validate:"required,max=255"`
	Body         string  `json:"body" validate:"required,max=4000"`
	MetadataJSON string  `json:"metadata_json" validate:"omitempty,max=4000"`
}
