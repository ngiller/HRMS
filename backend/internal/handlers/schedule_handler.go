package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

// ============================================================
// SCHEDULE TEMPLATES
// ============================================================

// ListTemplates returns paginated schedule templates
// GET /api/schedule-templates
func (h *ScheduleHandler) ListTemplates(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.scheduleService.ListTemplates(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Templates,
		"Berhasil memuat template jadwal",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAllTemplates returns all active templates (for dropdown)
// GET /api/schedule-templates/all
func (h *ScheduleHandler) GetAllTemplates(c *fiber.Ctx) error {
	templates, err := h.scheduleService.GetAllTemplates(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(templates, "Berhasil memuat daftar template"))
}

// GetTemplate returns a single template with days
// GET /api/schedule-templates/:id
func (h *ScheduleHandler) GetTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	t, err := h.scheduleService.GetTemplate(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(t, "Berhasil memuat detail template"))
}

// CreateTemplate creates a new schedule template
// POST /api/schedule-templates
func (h *ScheduleHandler) CreateTemplate(c *fiber.Ctx) error {
	req := new(models.CreateScheduleTemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	t, err := h.scheduleService.CreateTemplate(c.Context(), req)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch {
		case contains(err.Error(), "harus diisi"):
			code = fiber.StatusBadRequest
		case contains(err.Error(), "tidak valid"):
			code = fiber.StatusBadRequest
		case contains(err.Error(), "sudah digunakan"):
			code = fiber.StatusConflict
		case contains(err.Error(), "minimal"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(t, "Template jadwal berhasil ditambahkan"))
}

// UpdateTemplate updates a schedule template
// PUT /api/schedule-templates/:id
func (h *ScheduleHandler) UpdateTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateScheduleTemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	t, err := h.scheduleService.UpdateTemplate(c.Context(), id, req)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch {
		case contains(err.Error(), "tidak ditemukan"):
			code = fiber.StatusNotFound
		case contains(err.Error(), "sudah digunakan"):
			code = fiber.StatusConflict
		case contains(err.Error(), "tidak boleh kosong"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(t, "Template jadwal berhasil diperbarui"))
}

// DeleteTemplate soft-deletes a schedule template
// DELETE /api/schedule-templates/:id
func (h *ScheduleHandler) DeleteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.scheduleService.DeleteTemplate(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Template jadwal berhasil dihapus"))
}

// ============================================================
// EMPLOYEE SCHEDULES (Level 2 + Level 3)
// ============================================================

// ListEmployeeSchedules returns paginated employee schedules
// GET /api/employee-schedules?employee_id=xxx
func (h *ScheduleHandler) ListEmployeeSchedules(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	employeeID := c.Query("employee_id", "")

	resp, err := h.scheduleService.ListEmployeeSchedules(c.Context(), employeeID, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Schedules,
		"Berhasil memuat jadwal karyawan",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetEmployeeSchedule returns single employee schedule detail
// GET /api/employee-schedules/:id
func (h *ScheduleHandler) GetEmployeeSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	es, err := h.scheduleService.GetEmployeeSchedule(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(es, "Berhasil memuat detail jadwal"))
}

// CreateEmployeeSchedule creates a new employee schedule
// POST /api/employee-schedules
func (h *ScheduleHandler) CreateEmployeeSchedule(c *fiber.Ctx) error {
	req := new(models.CreateEmployeeScheduleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	es, err := h.scheduleService.CreateEmployeeSchedule(c.Context(), req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch {
		case contains(err.Error(), "harus diisi"):
			code = fiber.StatusBadRequest
		case contains(err.Error(), "sudah memiliki jadwal"):
			code = fiber.StatusConflict
		case contains(err.Error(), "violates foreign key"):
			code = fiber.StatusBadRequest
			return c.Status(code).JSON(ErrorResponse("Template jadwal tidak ditemukan"))
		case contains(err.Error(), "chk_schedule_date"):
			code = fiber.StatusBadRequest
			return c.Status(code).JSON(ErrorResponse("Tidak bisa memilih hari (periodik) dan tanggal spesifik bersamaan"))
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(es, "Jadwal karyawan berhasil ditambahkan"))
}

// UpdateEmployeeSchedule updates an employee schedule
// PUT /api/employee-schedules/:id
func (h *ScheduleHandler) UpdateEmployeeSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateEmployeeScheduleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	es, err := h.scheduleService.UpdateEmployeeSchedule(c.Context(), id, req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(es, "Jadwal karyawan berhasil diperbarui"))
}

// DeleteEmployeeSchedule deletes an employee schedule
// DELETE /api/employee-schedules/:id
func (h *ScheduleHandler) DeleteEmployeeSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.scheduleService.DeleteEmployeeSchedule(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Jadwal karyawan berhasil dihapus"))
}

// ============================================================
// RESOLVE SCHEDULE — Untuk absensi
// ============================================================

// ResolveSchedule returns the effective schedule for an employee on a given date
// GET /api/employee-schedules/resolve?employee_id=xxx&date=2026-06-17
func (h *ScheduleHandler) ResolveSchedule(c *fiber.Ctx) error {
	employeeID := c.Query("employee_id", "")
	date := c.Query("date", "")

	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("employee_id harus diisi"))
	}
	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("date harus diisi (YYYY-MM-DD)"))
	}

	rs, err := h.scheduleService.ResolveSchedule(c.Context(), employeeID, date)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ada jadwal") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(rs, "Jadwal berhasil ditentukan"))
}
