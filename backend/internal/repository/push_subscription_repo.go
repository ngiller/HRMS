package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

// PushSubscriptionRepo handles database operations for push subscriptions
type PushSubscriptionRepo struct{}

// NewPushSubscriptionRepo creates a new PushSubscriptionRepo
func NewPushSubscriptionRepo() *PushSubscriptionRepo {
	return &PushSubscriptionRepo{}
}

// Create inserts a new push subscription (upsert on conflict)
func (r *PushSubscriptionRepo) Create(ctx context.Context, userID string, req *models.SubscribeRequest, userAgent string) (*models.PushSubscription, error) {
	var sub models.PushSubscription
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh_key, auth_key, user_agent, device_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, endpoint) 
		DO UPDATE SET p256dh_key = $3, auth_key = $4, user_agent = $5, device_name = $6, is_active = TRUE, updated_at = NOW()
		RETURNING id::text, user_id::text, endpoint, p256dh_key, auth_key, COALESCE(user_agent, ''), COALESCE(device_name, ''), is_active, created_at, updated_at
	`, userID, req.Endpoint, req.P256DHKey, req.AuthKey, userAgent, req.DeviceName).Scan(
		&sub.ID, &sub.UserID, &sub.Endpoint, &sub.P256DHKey, &sub.AuthKey,
		&sub.UserAgent, &sub.DeviceName, &sub.IsActive, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create push subscription: %w", err)
	}
	return &sub, nil
}

// ListByUser returns all active subscriptions for a user
func (r *PushSubscriptionRepo) ListByUser(ctx context.Context, userID string) ([]models.PushSubscription, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT id::text, user_id::text, endpoint, p256dh_key, auth_key, COALESCE(user_agent, ''), COALESCE(device_name, ''), is_active, created_at, updated_at
		FROM push_subscriptions
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list push subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.P256DHKey, &sub.AuthKey,
			&sub.UserAgent, &sub.DeviceName, &sub.IsActive, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan push subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// ListAllActive returns all active subscriptions across all users (for broadcast)
func (r *PushSubscriptionRepo) ListAllActive(ctx context.Context) ([]models.PushSubscription, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT id::text, user_id::text, endpoint, p256dh_key, auth_key, COALESCE(user_agent, ''), COALESCE(device_name, ''), is_active, created_at, updated_at
		FROM push_subscriptions
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list all active push subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.P256DHKey, &sub.AuthKey,
			&sub.UserAgent, &sub.DeviceName, &sub.IsActive, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan push subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// Deactivate marks a subscription as inactive (unsubscribe)
func (r *PushSubscriptionRepo) Deactivate(ctx context.Context, id string, userID string) error {
	result, err := database.Pool.Exec(ctx, `
		UPDATE push_subscriptions SET is_active = FALSE, updated_at = NOW()
		WHERE id::text = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("deactivate push subscription: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("push subscription not found")
	}
	return nil
}

// DeactivateByEndpoint marks a subscription as inactive by endpoint (for invalidated subscriptions)
func (r *PushSubscriptionRepo) DeactivateByEndpoint(ctx context.Context, endpoint string) error {
	_, err := database.Pool.Exec(ctx, `
		UPDATE push_subscriptions SET is_active = FALSE, updated_at = NOW()
		WHERE endpoint = $1
	`, endpoint)
	if err != nil {
		return fmt.Errorf("deactivate push subscription by endpoint: %w", err)
	}
	return nil
}
