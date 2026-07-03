package models

import (
	"time"
)

// Notification represents a notification for an employee
type Notification struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	NotificationType string          `json:"notification_type"`
	Title            string          `json:"title"`
	Body             string          `json:"body"`
	Data             *map[string]any `json:"data,omitempty"`
	IsRead           bool            `json:"is_read"`
	IsPushed         bool            `json:"is_pushed"`
	IsEmailSent      bool            `json:"is_email_sent"`
	CreatedAt        time.Time       `json:"created_at"`
}

// NotificationSummary is a lightweight notification for listing
type NotificationSummary struct {
	ID               string    `json:"id"`
	NotificationType string    `json:"notification_type"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	IsRead           bool      `json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}

// NotificationListResponse wraps paginated notifications
type NotificationListResponse struct {
	Notifications []NotificationSummary `json:"notifications"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
	UnreadCount   int                   `json:"unread_count"`
}

// UnreadCountResponse is the response for unread count
type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

// NotificationPreference represents per-user notification delivery preferences
type NotificationPreference struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	NotificationType string `json:"notification_type"`
	InApp            bool   `json:"in_app"`
	Push             bool   `json:"push"`
	Email            bool   `json:"email"`
	WhatsApp         bool   `json:"whatsapp"`
}

// CreateNotificationRequest is the request to create a notification
type CreateNotificationRequest struct {
	UserID           string         `json:"user_id"`
	NotificationType string         `json:"notification_type"`
	Title            string         `json:"title"`
	Body             string         `json:"body"`
	Data             map[string]any `json:"data,omitempty"`
}

// MarkReadRequest is the request to mark notifications as read
type MarkReadRequest struct {
	IDs []string `json:"ids,omitempty"` // Empty means mark all as read
}
