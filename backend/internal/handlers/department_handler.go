package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DepartmentHandler struct {
	departmentService *service.DepartmentService
}

func NewDepartmentHandler(departmentService *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{departmentService: departmentService}
}

// ListDepartments returns paginated department list
// GET /api/departments
func (h *DepartmentHandler) ListDepartments(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.departmentService.ListDepartments(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Departments,
		"Berhasil memuat data departemen",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetWorkSchedules returns all active work schedules (for dropdown)
// GET /api/departments/work-schedules
func (h *DepartmentHandler) GetWorkSchedules(c *fiber.Ctx) error {
	schedules, err := h.departmentService.GetAllWorkSchedules(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal memuat data jadwal kerja"))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(schedules, "Berhasil memuat data jadwal kerja"))
}

// GetAllDepartments returns all active departments (for dropdowns)
// GET /api/departments/all
func (h *DepartmentHandler) GetAllDepartments(c *fiber.Ctx) error {
	departments, err := h.departmentService.GetAllDepartments(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(departments, "Berhasil memuat data departemen"))
}

// GetDepartment returns single department detail
// GET /api/departments/:id
func (h *DepartmentHandler) GetDepartment(c *fiber.Ctx) error {
	id := c.Params("id")

	dept, err := h.departmentService.GetDepartment(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(dept, "Berhasil memuat detail departemen"))
}

// CreateDepartment creates a new department
// POST /api/departments
func (h *DepartmentHandler) CreateDepartment(c *fiber.Ctx) error {
	req := new(models.CreateDepartmentRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	dept, err := h.departmentService.CreateDepartment(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama departemen harus diisi", "kode departemen harus diisi":
			status = fiber.StatusBadRequest
		case "kode departemen sudah digunakan", "nama departemen sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(dept, "Departemen berhasil ditambahkan"))
}

// UpdateDepartment updates an existing department
// PUT /api/departments/:id
func (h *DepartmentHandler) UpdateDepartment(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateDepartmentRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	dept, err := h.departmentService.UpdateDepartment(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "departemen tidak ditemukan":
			status = fiber.StatusNotFound
		case "kode departemen sudah digunakan", "nama departemen sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(dept, "Departemen berhasil diperbarui"))
}

// UpdateWorkSchedule assigns a work schedule to a department
// PUT /api/departments/:id/work-schedule
func (h *DepartmentHandler) UpdateWorkSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.AssignScheduleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	dept, err := h.departmentService.UpdateWorkSchedule(c.Context(), id, req.WorkScheduleID, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "departemen tidak ditemukan":
			status = fiber.StatusNotFound
		case "jadwal kerja tidak ditemukan":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(dept, "Jadwal kerja berhasil diupdate"))
}

// DeleteDepartment soft-deletes a department
// DELETE /api/departments/:id
func (h *DepartmentHandler) DeleteDepartment(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.departmentService.DeleteDepartment(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "departemen tidak ditemukan":
			status = fiber.StatusNotFound
		case "departemen memiliki sub-departemen, tidak dapat dihapus",
			"departemen memiliki karyawan, tidak dapat dihapus":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Departemen berhasil dihapus"))
}
