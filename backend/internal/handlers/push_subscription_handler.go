package handlers

import (
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// PushSubscriptionHandler handles HTTP requests for push notification subscriptions
type PushSubscriptionHandler struct {
	svc *service.WebPushService
}

// NewPushSubscriptionHandler creates a new PushSubscriptionHandler
func NewPushSubscriptionHandler(svc *service.WebPushService) *PushSubscriptionHandler {
	return &PushSubscriptionHandler{svc: svc}
}

// GetVapidPublicKey returns the VAPID public key for push subscription
// GET /api/push/vapid-public-key
func (h *PushSubscriptionHandler) GetVapidPublicKey(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse(fiber.Map{
		"public_key": h.svc.GetVapidPublicKey(),
	}, "VAPID public key"))
}

// Subscribe saves a new push subscription
// POST /api/push/subscribe
func (h *PushSubscriptionHandler) Subscribe(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	var req models.SubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid: " + err.Error()))
	}
	if req.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Endpoint harus diisi"))
	}
	if req.P256DHKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("P256DH key harus diisi"))
	}
	if req.AuthKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Auth key harus diisi"))
	}

	userAgent := string(c.Request().Header.UserAgent())
	subscription, err := h.svc.Subscribe(c.Context(), userID, &req, userAgent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal menyimpan subscription: " + err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(subscription, "Berlangganan notifikasi berhasil"))
}

// Unsubscribe deactivates a push subscription
// DELETE /api/push/subscribe/:id
func (h *PushSubscriptionHandler) Unsubscribe(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	id := c.Params("id")
	if err := h.svc.Unsubscribe(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse("Subscription tidak ditemukan"))
	}

	return c.JSON(SuccessResponse(nil, "Berhenti berlangganan notifikasi berhasil"))
}

// ListSubscriptions lists all active push subscriptions for the current user
// GET /api/push/subscriptions
func (h *PushSubscriptionHandler) ListSubscriptions(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	subs, err := h.svc.ListSubscriptions(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal mengambil subscription: " + err.Error()))
	}

	return c.JSON(SuccessResponse(subs, "Daftar subscription notifikasi"))
}
