package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"hrms-backend/internal/config"
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// WebPushService handles Web Push notification delivery via Browser Push API
type WebPushService struct {
	repo    *repository.PushSubscriptionRepo
	vapidPK string // VAPID public key (base64 URL-safe)
	vapidSK string // VAPID private key (base64 URL-safe)
	contact string // contact email for VAPID
}

// NewWebPushService creates a new WebPushService, auto-generating VAPID keys if needed
func NewWebPushService(vapidPK, vapidSK, contact string) *WebPushService {
	// Auto-generate VAPID keys if not set
	if vapidPK == "" || vapidSK == "" {
		pk, sk, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			log.Printf("⚠️ Failed to generate VAPID keys: %v", err)
		} else {
			vapidPK = pk
			vapidSK = sk
			log.Println("🔑 INFO: VAPID keys auto-generated. Set VAPID_PUBLIC_KEY & VAPID_PRIVATE_KEY di .env untuk persistensi.")
		}
	}

	return &WebPushService{
		repo:    repository.NewPushSubscriptionRepo(),
		vapidPK: vapidPK,
		vapidSK: vapidSK,
		contact: contact,
	}
}

// GetVapidPublicKey returns the VAPID public key for clients to subscribe
func (s *WebPushService) GetVapidPublicKey() string {
	return s.vapidPK
}

// Subscribe saves a new push subscription
func (s *WebPushService) Subscribe(ctx context.Context, userID string, req *models.SubscribeRequest, userAgent string) (*models.PushSubscription, error) {
	return s.repo.Create(ctx, userID, req, userAgent)
}

// Unsubscribe deactivates a push subscription
func (s *WebPushService) Unsubscribe(ctx context.Context, id, userID string) error {
	return s.repo.Deactivate(ctx, id, userID)
}

// ListSubscriptions returns all active subscriptions for a user
func (s *WebPushService) ListSubscriptions(ctx context.Context, userID string) ([]models.PushSubscription, error) {
	return s.repo.ListByUser(ctx, userID)
}

// NotifyUser sends a push notification to all active devices of a specific user
func (s *WebPushService) NotifyUser(ctx context.Context, userID string, notification *models.Notification) error {
	subs, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("get subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return nil // No subscriptions, skip silently
	}

	return s.sendToSubscriptions(ctx, subs, notification)
}

// sendToSubscriptions sends a push notification payload to multiple subscriptions
func (s *WebPushService) sendToSubscriptions(ctx context.Context, subs []models.PushSubscription, notification *models.Notification) error {
	payload := map[string]any{
		"title":   notification.Title,
		"body":    notification.Body,
		"type":    notification.NotificationType,
		"data":    notification.Data,
		"tag":     notification.NotificationType,
		"icon":    "/icons/icon-192.png",
		"badge":   "/icons/icon-192.png",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	for _, sub := range subs {
		subErr := s.sendSingle(ctx, sub, payloadBytes)
		if subErr != nil {
			lastErr = subErr
			// If subscription is invalid (410 Gone), deactivate it
			if strings.Contains(subErr.Error(), "invalid subscription") {
				if deactErr := s.repo.DeactivateByEndpoint(ctx, sub.Endpoint); deactErr != nil {
					log.Printf("⚠️ Failed to deactivate invalid subscription: %v", deactErr)
				}
			}
		}
	}
	return lastErr
}

// sendSingle sends a push notification to a single subscription
func (s *WebPushService) sendSingle(ctx context.Context, sub models.PushSubscription, payload []byte) error {
	// Parse subscription from stored data
	pushSub := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256DHKey,
			Auth:   sub.AuthKey,
		},
	}

	// Send notification with VAPID
	resp, err := webpush.SendNotification(payload, pushSub, &webpush.Options{
		Subscriber:      s.contact,
		VAPIDPublicKey:  s.vapidPK,
		VAPIDPrivateKey: s.vapidSK,
		TTL:             int(7 * 24 * 60 * 60), // 7 days TTL
	})
	if err != nil {
		return fmt.Errorf("send push: %w", err)
	}
	defer resp.Body.Close()

	// Webpush-go only returns error on >=500; check for 404/410 (expired subscription)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return fmt.Errorf("invalid subscription (HTTP %d)", resp.StatusCode)
	}

	return nil
}

// InitGlobalWebPushService initializes the global WebPushService from config
func InitGlobalWebPushService(cfg *config.Config) *WebPushService {
	contact := cfg.SMTPFrom
	if contact == "" {
		contact = "noreply@hrms.com"
	}
	return NewWebPushService(
		cfg.VAPIDPublicKey,
		cfg.VAPIDPrivateKey,
		contact,
	)
}

// GlobalWebPushService is the singleton instance accessible from other services
var GlobalWebPushService *WebPushService

// SetGlobalWebPushService sets the global WebPushService instance
func SetGlobalWebPushService(svc *WebPushService) {
	GlobalWebPushService = svc
}

// SendPushNotification sends a push notification using the global service (non-blocking)
func SendPushNotification(ctx context.Context, userID string, notification *models.Notification) {
	if GlobalWebPushService == nil {
		return
	}
	go func() {
		// Use a background context for the async call
		bgCtx := context.Background()
		if err := GlobalWebPushService.NotifyUser(bgCtx, userID, notification); err != nil {
			log.Printf("⚠️ Failed to send push notification: %v", err)
		}
	}()
}
