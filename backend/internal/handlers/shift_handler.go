package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ShiftHandler struct {
	shiftService *service.ShiftService
}

func NewShiftHandler(shiftService *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{shiftService: shiftService}
}

// ListShifts returns paginated shift list
// GET /api/shifts
func (h *ShiftHandler) ListShifts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")
	departmentID := c.Query("department_id", "")

	resp, err := h.shiftService.List(c.Context(), page, perPage, search, departmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Shifts,
		"Berhasil memuat data shift",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAllShifts returns all active shifts (for dropdown)
// GET /api/shifts/all
func (h *ShiftHandler) GetAllShifts(c *fiber.Ctx) error {
	departmentID := c.Query("department_id", "")
	shifts, err := h.shiftService.GetAll(c.Context(), departmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(shifts, "Berhasil memuat data shift"))
}

// GetShift returns single shift detail
// GET /api/shifts/:id
func (h *ShiftHandler) GetShift(c *fiber.Ctx) error {
	id := c.Params("id")

	shift, err := h.shiftService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(shift, "Berhasil memuat detail shift"))
}

// CreateShift creates a new shift
// POST /api/shifts
func (h *ShiftHandler) CreateShift(c *fiber.Ctx) error {
	req := new(models.CreateShiftRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	shift, err := h.shiftService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama shift harus diisi", "kode shift harus diisi", "jam mulai harus diisi", "jam selesai harus diisi":
			status = fiber.StatusBadRequest
		case "kode shift sudah digunakan", "nama shift sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(shift, "Shift berhasil ditambahkan"))
}

// UpdateShift updates an existing shift
// PUT /api/shifts/:id
func (h *ShiftHandler) UpdateShift(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateShiftRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	shift, err := h.shiftService.Update(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "shift tidak ditemukan":
			status = fiber.StatusNotFound
		case "kode shift sudah digunakan", "nama shift sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(shift, "Shift berhasil diperbarui"))
}

// DeleteShift soft-deletes a shift
// DELETE /api/shifts/:id
func (h *ShiftHandler) DeleteShift(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.shiftService.Delete(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "shift tidak ditemukan":
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Shift berhasil dihapus"))
}
