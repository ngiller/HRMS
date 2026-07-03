package handlers

import (
	"fmt"
	"strconv"

	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// getUserID extracts the current user ID from Fiber context locals
// Supports both string and uuid.UUID types (claims.UserID is uuid.UUID)
func getUserID(c *fiber.Ctx) string {
	uid := c.Locals("user_id")
	if uid == nil {
		return ""
	}
	switch v := uid.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ListNotifications returns paginated notifications for the current user
func (h *NotificationHandler) ListNotifications(c *fiber.Ctx) error {
	userID := getUserID(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))

	result, err := h.svc.ListNotifications(c.Context(), userID, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal mengambil notifikasi: " + err.Error()))
	}

	return c.JSON(SuccessResponse(result, "Daftar notifikasi"))
}

// GetUnreadCount returns the unread notification count for the current user
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := getUserID(c)

	count, err := h.svc.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal mengambil jumlah notifikasi: " + err.Error()))
	}

	return c.JSON(SuccessResponse(models.UnreadCountResponse{UnreadCount: count}, "Jumlah notifikasi belum dibaca"))
}

// MarkAsRead marks notifications as read for the current user
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := getUserID(c)

	var req models.MarkReadRequest
	if err := c.BodyParser(&req); err != nil {
		req = models.MarkReadRequest{}
	}

	if err := h.svc.MarkAsRead(c.Context(), userID, req.IDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal menandai notifikasi: " + err.Error()))
	}

	return c.JSON(SuccessResponse(nil, "Notifikasi telah ditandai sebagai sudah dibaca"))
}

// CreateNotification creates a new notification (admin/system endpoint)
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	var req models.CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid: " + err.Error()))
	}

	notification, err := h.svc.CreateNotification(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal membuat notifikasi: " + err.Error()))
	}

	return c.JSON(SuccessResponse(notification, "Notifikasi berhasil dibuat"))
}
