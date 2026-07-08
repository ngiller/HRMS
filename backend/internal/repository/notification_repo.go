package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

// NotificationRepo handles database operations for notifications
type NotificationRepo struct{}

// NewNotificationRepo creates a new NotificationRepo
func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{}
}

// ListNotifications returns paginated notifications for a user
func (r *NotificationRepo) ListNotifications(ctx context.Context, userID string, page, perPage int) (*models.NotificationListResponse, error) {
	offset := (page - 1) * perPage

	// Get total count and unread count
	var total, unreadCount int
	err := database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count notifications: %w", err)
	}

	err = database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND NOT is_read`, userID).Scan(&unreadCount)
	if err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}

	// Get paginated notifications
	rows, err := database.Pool.Query(ctx, `
		SELECT id, notification_type, title, body, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]models.NotificationSummary, 0)
	for rows.Next() {
		var n models.NotificationSummary
		if err := rows.Scan(&n.ID, &n.NotificationType, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return &models.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
		UnreadCount:   unreadCount,
	}, nil
}

// GetUnreadCount returns the number of unread notifications for a user
func (r *NotificationRepo) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND NOT is_read`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

// CreateNotification creates a new notification (system-generated, no audit trigger needed)
func (r *NotificationRepo) CreateNotification(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	var n models.Notification
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, notification_type, title, body, data)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		RETURNING id, user_id, notification_type, title, body, data, is_read, is_pushed, is_email_sent, created_at
	`, req.UserID, req.NotificationType, req.Title, req.Body, toJSONB(req.Data)).Scan(
		&n.ID, &n.UserID, &n.NotificationType, &n.Title, &n.Body, &n.Data, &n.IsRead, &n.IsPushed, &n.IsEmailSent, &n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return &n, nil
}

// MarkAsRead marks specific notifications as read for a user
func (r *NotificationRepo) MarkAsRead(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		// Mark all as read
		_, err := database.Pool.Exec(ctx,
			`UPDATE notifications SET is_read = true WHERE user_id = $1 AND NOT is_read`, userID)
		if err != nil {
			return fmt.Errorf("mark all as read: %w", err)
		}
		return nil
	}

	// Mark specific IDs as read
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, userID)
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf(
		`UPDATE notifications SET is_read = true WHERE user_id = $1 AND id IN (%s) AND NOT is_read`,
		strings.Join(placeholders, ","),
	)
	_, err := database.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

// CreateNotificationBatch creates multiple notifications in a batch (for system events)
func (r *NotificationRepo) CreateNotificationBatch(ctx context.Context, requests []*models.CreateNotificationRequest) error {
	if len(requests) == 0 {
		return nil
	}

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, req := range requests {
		_, err := tx.Exec(ctx, `
			INSERT INTO notifications (user_id, notification_type, title, body, data)
			VALUES ($1, $2, $3, $4, $5::jsonb)
		`, req.UserID, req.NotificationType, req.Title, req.Body, toJSONB(req.Data))
		if err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetEmployeeEmailByUserID retrieves an employee's email by their user ID
func GetEmployeeEmailByUserID(ctx context.Context, userID string, email *string) error {
	return database.Pool.QueryRow(ctx, `SELECT email FROM employees WHERE id::text = $1 AND deleted_at IS NULL`, userID).Scan(email)
}

// UpdateNotificationEmailSent marks a notification's is_email_sent flag as true
func UpdateNotificationEmailSent(ctx context.Context, notificationID string) error {
	_, err := database.Pool.Exec(ctx,
		`UPDATE notifications SET is_email_sent = true WHERE id::text = $1`, notificationID)
	if err != nil {
		return fmt.Errorf("update email sent: %w", err)
	}
	return nil
}

// toJSONB converts a map to a JSON string for PostgreSQL jsonb
func toJSONB(data map[string]any) *string {
	if data == nil {
		return nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
