package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type WorkScheduleHandler struct {
	workScheduleService *service.WorkScheduleService
}

func NewWorkScheduleHandler(workScheduleService *service.WorkScheduleService) *WorkScheduleHandler {
	return &WorkScheduleHandler{workScheduleService: workScheduleService}
}

// ListWorkSchedules returns paginated work schedule list
// GET /api/work-schedules
func (h *WorkScheduleHandler) ListWorkSchedules(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.workScheduleService.List(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.WorkSchedules,
		"Berhasil memuat data jadwal kerja",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAllWorkSchedules returns all active work schedules (for dropdown)
// GET /api/work-schedules/all
func (h *WorkScheduleHandler) GetAllWorkSchedules(c *fiber.Ctx) error {
	schedules, err := h.workScheduleService.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(schedules, "Berhasil memuat data jadwal kerja"))
}

// GetWorkSchedule returns single work schedule detail
// GET /api/work-schedules/:id
func (h *WorkScheduleHandler) GetWorkSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	ws, err := h.workScheduleService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(ws, "Berhasil memuat detail jadwal kerja"))
}

// CreateWorkSchedule creates a new work schedule
// POST /api/work-schedules
func (h *WorkScheduleHandler) CreateWorkSchedule(c *fiber.Ctx) error {
	req := new(models.CreateWorkScheduleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	ws, err := h.workScheduleService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama jadwal kerja harus diisi", "tipe jadwal harus dipilih", "tipe jadwal tidak valid (5_day, 6_day, atau shift)":
			status = fiber.StatusBadRequest
		case "nama jadwal kerja sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(ws, "Jadwal kerja berhasil ditambahkan"))
}

// UpdateWorkSchedule updates an existing work schedule
// PUT /api/work-schedules/:id
func (h *WorkScheduleHandler) UpdateWorkSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateWorkScheduleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	ws, err := h.workScheduleService.Update(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "jadwal kerja tidak ditemukan":
			status = fiber.StatusNotFound
		case "nama jadwal kerja sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(ws, "Jadwal kerja berhasil diperbarui"))
}

// DeleteWorkSchedule soft-deletes a work schedule
// DELETE /api/work-schedules/:id
func (h *WorkScheduleHandler) DeleteWorkSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.workScheduleService.Delete(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "jadwal kerja tidak ditemukan":
			status = fiber.StatusNotFound
		case "jadwal kerja masih digunakan oleh departemen, tidak dapat dihapus":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Jadwal kerja berhasil dihapus"))
}
