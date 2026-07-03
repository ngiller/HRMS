package service

import (
	"context"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

// NotificationService handles business logic for notifications
type NotificationService struct {
	repo *repository.NotificationRepo
}

// NewNotificationService creates a new NotificationService
func NewNotificationService() *NotificationService {
	return &NotificationService{
		repo: repository.NewNotificationRepo(),
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
	return s.repo.CreateNotification(ctx, req)
}

// MarkAsRead marks notifications as read
func (s *NotificationService) MarkAsRead(ctx context.Context, userID string, ids []string) error {
	return s.repo.MarkAsRead(ctx, userID, ids)
}

// NotifyLeaveApproved creates a notification when a leave request is approved
func (s *NotificationService) NotifyLeaveApproved(ctx context.Context, employeeID, leaveType string) error {
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "leave_approved",
		Title:            "Cuti Disetujui",
		Body:             fmt.Sprintf("Pengajuan cuti %s Anda telah disetujui.", leaveType),
		Data: map[string]any{
			"type": "leave",
		},
	})
	return err
}

// NotifyLeaveRejected creates a notification when a leave request is rejected
func (s *NotificationService) NotifyLeaveRejected(ctx context.Context, employeeID, leaveType, reason string) error {
	body := fmt.Sprintf("Pengajuan cuti %s Anda ditolak.", leaveType)
	if reason != "" {
		body += " Alasan: " + reason
	}
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "leave_rejected",
		Title:            "Cuti Ditolak",
		Body:             body,
		Data: map[string]any{
			"type": "leave",
		},
	})
	return err
}

// NotifyOvertimeApproved creates a notification when overtime is approved
func (s *NotificationService) NotifyOvertimeApproved(ctx context.Context, employeeID string) error {
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "overtime_approved",
		Title:            "Lembur Disetujui",
		Body:             "Pengajuan lembur Anda telah disetujui.",
		Data: map[string]any{
			"type": "overtime",
		},
	})
	return err
}

// NotifyReimbursementApproved creates a notification when reimbursement is approved
func (s *NotificationService) NotifyReimbursementApproved(ctx context.Context, employeeID string) error {
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "reimbursement_approved",
		Title:            "Reimbursement Disetujui",
		Body:             "Pengajuan reimbursement Anda telah disetujui dan akan segera dibayarkan.",
		Data: map[string]any{
			"type": "reimbursement",
		},
	})
	return err
}

// NotifyLoanApproved creates a notification when a loan is approved
func (s *NotificationService) NotifyLoanApproved(ctx context.Context, employeeID string) error {
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "loan_approved",
		Title:            "Pinjaman Disetujui",
		Body:             "Pengajuan pinjaman Anda telah disetujui.",
		Data: map[string]any{
			"type": "loan",
		},
	})
	return err
}

// NotifyShiftChangeApproved creates a notification when a shift change is approved
func (s *NotificationService) NotifyShiftChangeApproved(ctx context.Context, employeeID string) error {
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "shift_change_approved",
		Title:            "Perubahan Shift Disetujui",
		Body:             "Permintaan perubahan shift Anda telah disetujui.",
		Data: map[string]any{
			"type": "shift_change",
		},
	})
	return err
}

// NotifyReprimandIssued creates a notification when a reprimand is issued
func (s *NotificationService) NotifyReprimandIssued(ctx context.Context, employeeID, reprimandType string) error {
	_, err := s.repo.CreateNotification(ctx, &models.CreateNotificationRequest{
		UserID:           employeeID,
		NotificationType: "reprimand_issued",
		Title:            "Surat Peringatan",
		Body:             fmt.Sprintf("Anda menerima Surat Peringatan (%s). Silakan konfirmasi.", reprimandType),
		Data: map[string]any{
			"type": "reprimand",
		},
	})
	return err
}
