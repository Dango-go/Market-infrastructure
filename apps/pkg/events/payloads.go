package events

import "time"

type UserRegistered struct {
	AccountID    string    `json:"account_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	RegisteredAt time.Time `json:"registered_at"`
}

type UserProfileCreated struct {
	AccountID   string    `json:"account_id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserProfileUpdated struct {
	AccountID      string    `json:"account_id"`
	DisplayName    string    `json:"display_name"`
	AvatarURL      string    `json:"avatar_url"`
	Locale         string    `json:"locale"`
	Timezone       string    `json:"timezone"`
	MarketingOptIn bool      `json:"marketing_opt_in"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserAddressUpdated struct {
	AccountID         string    `json:"account_id"`
	AddressID         string    `json:"address_id"`
	Action            string    `json:"action"`
	Label             string    `json:"label,omitempty"`
	IsDefaultShipping bool      `json:"is_default_shipping"`
	IsDefaultBilling  bool      `json:"is_default_billing"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UserPreferencesUpdated struct {
	AccountID          string    `json:"account_id"`
	Currency           string    `json:"currency"`
	Language           string    `json:"language"`
	EmailNotifications bool      `json:"email_notifications"`
	SMSNotifications   bool      `json:"sms_notifications"`
	PushNotifications  bool      `json:"push_notifications"`
	MarketingOptIn     bool      `json:"marketing_opt_in"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type UserWishlistUpdated struct {
	AccountID string    `json:"account_id"`
	ProductID string    `json:"product_id"`
	Action    string    `json:"action"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductCreated struct {
	ProductID  string `json:"product_id"`
	CategoryID string `json:"category_id"`
	BrandID    string `json:"brand_id"`
	Slug       string `json:"slug"`
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Featured   bool   `json:"featured"`
}

type ProductUpdated struct {
	ProductID  string `json:"product_id"`
	CategoryID string `json:"category_id"`
	BrandID    string `json:"brand_id"`
	Slug       string `json:"slug"`
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Featured   bool   `json:"featured"`
}

type CompatibilityRulePayload struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type ProductCompatibilityUpdated struct {
	ProductID string                     `json:"product_id"`
	Slug      string                     `json:"slug"`
	Rules     []CompatibilityRulePayload `json:"rules"`
}

type StockReserved struct {
	ReservationID string `json:"reservation_id"`
	ProductID     string `json:"product_id"`
	WarehouseID   string `json:"warehouse_id"`
	Reference     string `json:"reference"`
	Quantity      int64  `json:"quantity"`
}

type StockReleased struct {
	ReservationID string `json:"reservation_id"`
	ProductID     string `json:"product_id"`
	WarehouseID   string `json:"warehouse_id"`
	Reference     string `json:"reference"`
	Quantity      int64  `json:"quantity"`
}

type StockAdjusted struct {
	ProductID   string `json:"product_id"`
	WarehouseID string `json:"warehouse_id"`
	OnHand      int64  `json:"on_hand"`
	Reserved    int64  `json:"reserved"`
	Available   int64  `json:"available"`
	Delta       int64  `json:"delta"`
}

type PriceChanged struct {
	ProductID      string `json:"product_id"`
	Currency       string `json:"currency"`
	AmountCents    int64  `json:"amount_cents"`
	CompareAtCents int64  `json:"compare_at_cents"`
	Active         bool   `json:"active"`
}

type PromotionStarted struct {
	PromotionID  string    `json:"promotion_id"`
	Code         string    `json:"code"`
	DiscountType string    `json:"discount_type"`
	ValueCents   int64     `json:"value_cents"`
	PercentOff   int       `json:"percent_off"`
	StartsAt     time.Time `json:"starts_at"`
}

type CartCheckedOutItem struct {
	ProductID      string `json:"product_id"`
	Quantity       int32  `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type CartCheckedOut struct {
	CartID        string               `json:"cart_id"`
	AccountID     string               `json:"account_id"`
	Currency      string               `json:"currency"`
	SubtotalCents int64                `json:"subtotal_cents"`
	Items         []CartCheckedOutItem `json:"items"`
}

type OrderCreatedItem struct {
	ProductID      string `json:"product_id"`
	Quantity       int32  `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type OrderCreated struct {
	OrderID         string             `json:"order_id"`
	AccountID       string             `json:"account_id"`
	CartID          string             `json:"cart_id,omitempty"`
	Status          string             `json:"status"`
	Currency        string             `json:"currency"`
	SubtotalCents   int64              `json:"subtotal_cents"`
	ShippingCents   int64              `json:"shipping_cents"`
	TotalCents      int64              `json:"total_cents"`
	DeliveryMethod  string             `json:"delivery_method"`
	DeliveryAddress string             `json:"delivery_address"`
	Items           []OrderCreatedItem `json:"items"`
}

type OrderStatusUpdated struct {
	OrderID   string `json:"order_id"`
	AccountID string `json:"account_id"`
	Status    string `json:"status"`
}

type ShipmentCreated struct {
	ShipmentID         string `json:"shipment_id"`
	OrderID            string `json:"order_id"`
	AccountID          string `json:"account_id"`
	Status             string `json:"status"`
	Carrier            string `json:"carrier"`
	ServiceLevel       string `json:"service_level"`
	TrackingNumber     string `json:"tracking_number,omitempty"`
	DestinationAddress string `json:"destination_address"`
}

type ShipmentStatusUpdated struct {
	ShipmentID     string `json:"shipment_id"`
	OrderID        string `json:"order_id"`
	AccountID      string `json:"account_id"`
	Status         string `json:"status"`
	Carrier        string `json:"carrier"`
	TrackingNumber string `json:"tracking_number,omitempty"`
}

type PaymentCreated struct {
	PaymentID      string `json:"payment_id"`
	OrderID        string `json:"order_id"`
	AccountID      string `json:"account_id"`
	Status         string `json:"status"`
	Provider       string `json:"provider"`
	Method         string `json:"method"`
	Currency       string `json:"currency"`
	AmountCents    int64  `json:"amount_cents"`
	TransactionRef string `json:"transaction_ref,omitempty"`
}

type PaymentSucceeded struct {
	PaymentID      string `json:"payment_id"`
	OrderID        string `json:"order_id"`
	AccountID      string `json:"account_id"`
	Status         string `json:"status"`
	AmountCents    int64  `json:"amount_cents"`
	Currency       string `json:"currency"`
	TransactionRef string `json:"transaction_ref,omitempty"`
}

type PaymentFailed struct {
	PaymentID      string `json:"payment_id"`
	OrderID        string `json:"order_id"`
	AccountID      string `json:"account_id"`
	Status         string `json:"status"`
	FailureReason  string `json:"failure_reason"`
	TransactionRef string `json:"transaction_ref,omitempty"`
}

type PaymentRefunded struct {
	PaymentID      string `json:"payment_id"`
	OrderID        string `json:"order_id"`
	AccountID      string `json:"account_id"`
	Status         string `json:"status"`
	AmountCents    int64  `json:"amount_cents"`
	Currency       string `json:"currency"`
	TransactionRef string `json:"transaction_ref,omitempty"`
}

type UserFollowed struct { FollowerID string `json:"follower_id"`; FolloweeID string `json:"followee_id"` }

type ProjectCreated struct { ProjectID string `json:"project_id"`; OwnerID string `json:"owner_id"`; Slug string `json:"slug"`; Name string `json:"name"`; ProjectType string `json:"project_type"`; Platforms []string `json:"platforms"` }

type ProjectUpdated struct { ProjectID string `json:"project_id"`; Slug string `json:"slug"`; Name string `json:"name"`; Platforms []string `json:"platforms"` }

type VersionReleased struct { ProjectID string `json:"project_id"`; ReleaseID string `json:"release_id"`; Version string `json:"version"`; Channel string `json:"channel"` }

type ArtifactUploaded struct { ArtifactID string `json:"artifact_id"`; ReleaseID string `json:"release_id"`; ProjectID string `json:"project_id"`; Filename string `json:"filename"`; SizeBytes int64 `json:"size_bytes"`; Checksum string `json:"checksum"` }

type DownloadCompleted struct { ArtifactID string `json:"artifact_id"`; ProjectID string `json:"project_id"`; AccountID string `json:"account_id"` }

type ReviewCreated struct { ReviewID string `json:"review_id"`; ProjectID string `json:"project_id"`; AuthorID string `json:"author_id"`; Rating int `json:"rating"` }

type CommentCreated struct { CommentID string `json:"comment_id"`; ProjectID string `json:"project_id"`; AuthorID string `json:"author_id"` }

type StarAdded struct { ProjectID string `json:"project_id"`; AccountID string `json:"account_id"` }

type NotificationSent struct { NotificationID string `json:"notification_id"`; RecipientID string `json:"recipient_id"`; Channel string `json:"channel"` }

type AnalyticsUpdated struct { Subject string `json:"subject"`; Metric string `json:"metric"`; Value int64 `json:"value"` }
