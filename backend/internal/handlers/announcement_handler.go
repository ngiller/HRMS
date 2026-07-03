package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
}

func NewAnnouncementHandler(announcementService *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{announcementService: announcementService}
}

// ListAnnouncements returns paginated announcement list
// GET /api/announcements
func (h *AnnouncementHandler) ListAnnouncements(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	announcementType := c.Query("type", "")

	resp, err := h.announcementService.List(c.Context(), page, perPage, announcementType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Announcements,
		"Berhasil memuat data pengumuman",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAnnouncement returns single announcement detail
// GET /api/announcements/:id
func (h *AnnouncementHandler) GetAnnouncement(c *fiber.Ctx) error {
	id := c.Params("id")

	a, err := h.announcementService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(a, "Berhasil memuat detail pengumuman"))
}

// CreateAnnouncement creates a new announcement
// POST /api/announcements
func (h *AnnouncementHandler) CreateAnnouncement(c *fiber.Ctx) error {
	req := new(models.CreateAnnouncementReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	a, err := h.announcementService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "judul pengumuman harus diisi", "konten pengumuman harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(a, "Pengumuman berhasil dibuat"))
}

// UpdateAnnouncement updates an announcement
// PUT /api/announcements/:id
func (h *AnnouncementHandler) UpdateAnnouncement(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateAnnouncementReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	a, err := h.announcementService.Update(c.Context(), id, userID, req)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(a, "Pengumuman berhasil diperbarui"))
}

// DeleteAnnouncement soft-deletes an announcement
// DELETE /api/announcements/:id
func (h *AnnouncementHandler) DeleteAnnouncement(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	if err := h.announcementService.Delete(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Pengumuman berhasil dihapus"))
}

// MarkAnnouncementRead marks announcement as read
// POST /api/announcements/:id/read
func (h *AnnouncementHandler) MarkAnnouncementRead(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	if err := h.announcementService.MarkAsRead(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Pengumuman ditandai sudah dibaca"))
}
