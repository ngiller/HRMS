package handlers

import (
	"strconv"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ActivityLogHandler handles HTTP requests for activity logs
type ActivityLogHandler struct {
	svc *service.ActivityLogService
}

// NewActivityLogHandler creates a new ActivityLogHandler
func NewActivityLogHandler(svc *service.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{svc: svc}
}

// ListActivityLogs returns paginated activity logs with filters
func (h *ActivityLogHandler) ListActivityLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate != "" {
		if _, err := time.Parse(time.RFC3339, startDate); err != nil {
			if _, err := time.Parse("2006-01-02", startDate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format start_date tidak valid. Gunakan YYYY-MM-DD"))
			}
		}
	}
	if endDate != "" {
		if _, err := time.Parse(time.RFC3339, endDate); err != nil {
			if _, err := time.Parse("2006-01-02", endDate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format end_date tidak valid. Gunakan YYYY-MM-DD"))
			}
		}
	}

	// Regular employees can only see their own activity logs
	userID := c.Query("user_id")
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID = database.UserIDFromContext(c.Locals("user_id"))
	}

	filter := &models.ActivityLogFilter{
		Action:     c.Query("action"),
		EntityType: c.Query("entity_type"),
		UserID:     userID,
		StartDate:  startDate,
		EndDate:    endDate,
		Page:       page,
		PerPage:    perPage,
	}

	result, err := h.svc.ListActivityLogs(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal mengambil log aktivitas: " + err.Error()))
	}

	return c.JSON(SuccessResponse(result, "Daftar log aktivitas"))
}

// GetActivityLog returns a single activity log by ID
func (h *ActivityLogHandler) GetActivityLog(c *fiber.Ctx) error {
	id := c.Params("id")

	roleSlug, _ := c.Locals("role_slug").(string)

	log, err := h.svc.GetActivityLog(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse("Log aktivitas tidak ditemukan"))
	}

	// Regular employees can only see their own logs
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if log.UserID == nil || *log.UserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat log ini"))
		}
	}

	return c.JSON(SuccessResponse(log, "Detail log aktivitas"))
}

// GetEntityTypes returns distinct entity types for filter dropdown
func (h *ActivityLogHandler) GetEntityTypes(c *fiber.Ctx) error {
	types, err := h.svc.GetDistinctEntityTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal mengambil tipe entitas: " + err.Error()))
	}

	return c.JSON(SuccessResponse(fiber.Map{"entity_types": types}, "Daftar tipe entitas"))
}

// GetActions returns distinct action types for filter dropdown
func (h *ActivityLogHandler) GetActions(c *fiber.Ctx) error {
	actions, err := h.svc.GetDistinctActions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal mengambil tipe aksi: " + err.Error()))
	}

	return c.JSON(SuccessResponse(fiber.Map{"actions": actions}, "Daftar tipe aksi"))
}
