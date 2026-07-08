package models

import "time"

// PushSubscription represents a Web Push subscription from a browser/device
type PushSubscription struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Endpoint   string    `json:"endpoint"`
	P256DHKey  string    `json:"p256dh_key"`
	AuthKey    string    `json:"auth_key"`
	UserAgent  string    `json:"user_agent,omitempty"`
	DeviceName string    `json:"device_name,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SubscribeRequest is the request to subscribe to push notifications
type SubscribeRequest struct {
	Endpoint   string `json:"endpoint"`
	P256DHKey  string `json:"p256dh_key"`
	AuthKey    string `json:"auth_key"`
	DeviceName string `json:"device_name,omitempty"`
}

// PushSubscriptionListResponse wraps paginated push subscriptions
type PushSubscriptionListResponse struct {
	Subscriptions []PushSubscription `json:"subscriptions"`
	Total         int                `json:"total"`
	Page          int                `json:"page"`
	PerPage       int                `json:"per_page"`
}
