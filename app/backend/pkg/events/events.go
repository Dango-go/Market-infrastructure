package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Topic string

const (
	// aliases
	TopicUserRegistered               Topic = "user.registered"
	TopicUserProfileCreated           Topic = "user.profile.created"
	TopicUserProfileUpdated           Topic = "user.profile.updated"
	TopicUserAddressesUpdated         Topic = "user.addresses.updated"
	TopicUserPreferencesUpdated       Topic = "user.preferences.updated"
	TopicUserWishlistUpdated          Topic = "user.wishlist.updated"
	TopicProductCreated               Topic = "product.created"
	TopicProductUpdated               Topic = "product.updated"
	TopicProductCompatibilityUpdated  Topic = "product.compatibility.updated"
	TopicStockReserved                Topic = "stock.reserved"
	TopicStockReleased                Topic = "stock.released"
	TopicStockAdjusted                Topic = "stock.adjusted"
	TopicPriceChanged                 Topic = "price.changed"
	TopicPromotionStarted             Topic = "promotion.started"
	TopicCartCheckedOut               Topic = "cart.checked_out"
	TopicOrderCreated                 Topic = "order.created"
	TopicOrderStatusUpdated           Topic = "order.status.updated"
	TopicShipmentCreated              Topic = "shipment.created"
	TopicShipmentStatusUpdated        Topic = "shipment.status.updated"
	TopicPaymentCreated               Topic = "payment.created"
	TopicPaymentSucceeded             Topic = "payment.succeeded"
	TopicPaymentFailed                Topic = "payment.failed"
	TopicPaymentRefunded              Topic = "payment.refunded"
	TopicUserFollowed                 Topic = "user.followed"
	TopicProjectCreated               Topic = "project.created"
	TopicProjectUpdated               Topic = "project.updated"
	TopicVersionReleased              Topic = "version.released"
	TopicArtifactUploaded             Topic = "artifact.uploaded"
	TopicDownloadCompleted            Topic = "download.completed"
	TopicReviewCreated                Topic = "review.created"
	TopicCommentCreated               Topic = "comment.created"
	TopicStarAdded                    Topic = "star.added"
	TopicNotificationSent             Topic = "notification.sent"
	TopicAnalyticsUpdated             Topic = "analytics.updated"
)

type Envelope struct {
	ID            uuid.UUID       `json:"id"`
	Type          Topic           `json:"type"`  // order.created
	Version       int             `json:"version"`
	Source        string          `json:"source"`
	Subject       string          `json:"subject"`
	CorrelationID string          `json:"correlation_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Data          json.RawMessage `json:"data"`   // json data
}

func NewEnvelope(id uuid.UUID, topic Topic, source, subject, correlationID string, occurredAt time.Time, payload any) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil { return Envelope{}, err }
	return Envelope{ID: id, Type: topic, Version: 1, Source: source, Subject: subject, CorrelationID: correlationID, OccurredAt: occurredAt, Data: data}, nil
}

func (e Envelope) Decode(target any) error { return json.Unmarshal(e.Data, target) }
