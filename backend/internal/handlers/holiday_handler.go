package handlers

import (
	"strconv"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type HolidayHandler struct {
	holidayService *service.HolidayService
}

func NewHolidayHandler(holidayService *service.HolidayService) *HolidayHandler {
	return &HolidayHandler{holidayService: holidayService}
}

// ListHolidays returns paginated holiday list
// GET /api/holidays
func (h *HolidayHandler) ListHolidays(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	year := c.QueryInt("year", 0)
	holidayType := c.Query("type", "")

	resp, err := h.holidayService.List(c.Context(), page, perPage, year, holidayType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Holidays,
		"Berhasil memuat data hari libur",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetHoliday returns single holiday detail
// GET /api/holidays/:id
func (h *HolidayHandler) GetHoliday(c *fiber.Ctx) error {
	id := c.Params("id")

	holiday, err := h.holidayService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(holiday, "Berhasil memuat detail hari libur"))
}

// CreateHoliday creates a new holiday
// POST /api/holidays
func (h *HolidayHandler) CreateHoliday(c *fiber.Ctx) error {
	req := new(models.CreateHolidayReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	holiday, err := h.holidayService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama hari libur harus diisi", "tanggal hari libur harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(holiday, "Hari libur berhasil ditambahkan"))
}

// UpdateHoliday updates a holiday
// PUT /api/holidays/:id
func (h *HolidayHandler) UpdateHoliday(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateHolidayReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	holiday, err := h.holidayService.Update(c.Context(), id, userID, req)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(holiday, "Hari libur berhasil diperbarui"))
}

// DeleteHoliday deletes a holiday
// DELETE /api/holidays/:id
func (h *HolidayHandler) DeleteHoliday(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	if err := h.holidayService.Delete(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Hari libur berhasil dihapus"))
}

// GetHolidaysByYear returns all holidays for a year
// GET /api/holidays/year/:year
func (h *HolidayHandler) GetHolidaysByYear(c *fiber.Ctx) error {
	yearStr := c.Params("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		year = 2026
	}

	resp, err := h.holidayService.GetByYear(c.Context(), year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp, "Berhasil memuat hari libur tahun "+yearStr))
}
