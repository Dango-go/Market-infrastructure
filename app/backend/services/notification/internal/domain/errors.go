package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrTemplateNotFound     = apperr.NotFound("notification_template_not_found", "notification template not found")
	ErrNotificationNotFound = apperr.NotFound("notification_not_found", "notification not found")
)
