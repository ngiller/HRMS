package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type SalaryComponentHandler struct {
	salaryComponentService *service.SalaryComponentService
}

func NewSalaryComponentHandler(salaryComponentService *service.SalaryComponentService) *SalaryComponentHandler {
	return &SalaryComponentHandler{salaryComponentService: salaryComponentService}
}

// ListComponents returns paginated salary components for an employee
// GET /api/employees/:id/salary-components
func (h *SalaryComponentHandler) ListComponents(c *fiber.Ctx) error {
	employeeID := c.Params("id")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)

	resp, err := h.salaryComponentService.ListComponents(c.Context(), employeeID, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Components,
		"Berhasil memuat komponen gaji",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// CreateComponent adds a new salary component for an employee
// POST /api/employees/:id/salary-components
func (h *SalaryComponentHandler) CreateComponent(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	var req models.CreateSalaryComponentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	comp, err := h.salaryComponentService.CreateComponent(c.Context(), employeeID, &req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "harus diisi"):
			code = fiber.StatusBadRequest
		case contains(msg, "harus allowance atau deduction"):
			code = fiber.StatusBadRequest
		case contains(msg, "tidak boleh negatif"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(comp, "Komponen gaji berhasil ditambahkan"))
}

// UpdateComponent updates a salary component
// PUT /api/employees/:id/salary-components/:componentId
func (h *SalaryComponentHandler) UpdateComponent(c *fiber.Ctx) error {
	componentID := c.Params("componentId")

	var req models.UpdateSalaryComponentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	comp, err := h.salaryComponentService.UpdateComponent(c.Context(), componentID, &req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case contains(msg, "harus allowance atau deduction"):
			code = fiber.StatusBadRequest
		case contains(msg, "tidak boleh negatif"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(comp, "Komponen gaji berhasil diperbarui"))
}

// DeleteComponent soft-deletes a salary component
// DELETE /api/employees/:id/salary-components/:componentId
func (h *SalaryComponentHandler) DeleteComponent(c *fiber.Ctx) error {
	componentID := c.Params("componentId")
	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.salaryComponentService.DeleteComponent(c.Context(), componentID, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		if contains(msg, "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Komponen gaji berhasil dihapus"))
}
