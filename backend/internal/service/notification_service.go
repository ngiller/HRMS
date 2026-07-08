package service

import (
	"context"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

// NotificationService handles business logic for notifications
type NotificationService struct {
	repo         *repository.NotificationRepo
	emailService *EmailService
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(emailService *EmailService) *NotificationService {
	return &NotificationService{
		repo:         repository.NewNotificationRepo(),
		emailService: emailService,
	}
}

// ListNotifications returns paginated notifications for a user
func (s *NotificationService) ListNotifications(ctx context.Context, userID string, page, perPage int) (*models.NotificationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return s.repo.ListNotifications(ctx, userID, page, perPage)
}

// GetUnreadCount returns the number of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if req.NotificationType == "" {
		return nil, fmt.Errorf("notification_type is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	n, err := s.repo.CreateNotification(ctx, req)
	if err != nil {
		return nil, err
	}

	// Try to send email notification if email service is available
	if s.emailService != nil && s.emailService.IsEnabled() {
		// Get user's email asynchronously
		go func() {
			emailCtx := context.Background()
			var toEmail string
			err := repository.GetEmployeeEmailByUserID(emailCtx, req.UserID, &toEmail)
			if err == nil && toEmail != "" {
				if sendErr := s.emailService.SendNotification(emailCtx, req, toEmail); sendErr == nil {
					_ = repository.UpdateNotificationEmailSent(emailCtx, n.ID)
				}
			}
		}()
	}

	return n, nil
}

// MarkAsRead marks notifications as read
func (s *NotificationService) MarkAsRead(ctx context.Context, userID string, ids []string) error {
	return s.repo.MarkAsRead(ctx, userID, ids)
}

