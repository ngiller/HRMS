package handlers

import (
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type EmployeeHandler struct {
	employeeService *service.EmployeeService
}

func NewEmployeeHandler(employeeService *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employeeService: employeeService}
}

// ListEmployees returns paginated employee list
// GET /api/employees
func (h *EmployeeHandler) ListEmployees(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.employeeService.ListEmployees(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    resp.Employees,
		"meta": fiber.Map{
			"total":    resp.Total,
			"page":     resp.Page,
			"per_page": resp.PerPage,
		},
	})
}

// GetEmployee returns single employee detail
// GET /api/employees/:id
func (h *EmployeeHandler) GetEmployee(c *fiber.Ctx) error {
	id := c.Params("id")

	employee, err := h.employeeService.GetEmployee(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    employee,
	})
}

// Dashboard returns executive dashboard stats
// GET /api/dashboard
func (h *EmployeeHandler) Dashboard(c *fiber.Ctx) error {
	stats, err := h.employeeService.GetDashboard(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}
