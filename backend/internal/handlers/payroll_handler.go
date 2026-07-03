package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type PayrollHandler struct {
	payrollService *service.PayrollService
}

func NewPayrollHandler(payrollService *service.PayrollService) *PayrollHandler {
	return &PayrollHandler{payrollService: payrollService}
}

// ListPeriods returns paginated payroll periods
// GET /api/payroll/periods
func (h *PayrollHandler) ListPeriods(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)

	resp, err := h.payrollService.ListPeriods(c.Context(), page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Periods,
		"Berhasil memuat periode penggajian",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// CreatePeriod creates a new payroll period
// POST /api/payroll/periods
func (h *PayrollHandler) CreatePeriod(c *fiber.Ctx) error {
	var req models.CreatePayrollPeriodRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	period, err := h.payrollService.CreatePeriod(c.Context(), &req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "harus diisi"), contains(msg, "tidak valid"):
			code = fiber.StatusBadRequest
		case contains(msg, "sudah ada"):
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(period, "Periode penggajian berhasil dibuat"))
}

// GetPeriod returns a single payroll period
// GET /api/payroll/periods/:id
func (h *PayrollHandler) GetPeriod(c *fiber.Ctx) error {
	periodID := c.Params("id")

	period, err := h.payrollService.GetPeriod(c.Context(), periodID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(period, "Berhasil memuat periode"))
}

// CalculatePayroll calculates payroll for all (or a specific) employee in a period
// POST /api/payroll/periods/:id/calculate
func (h *PayrollHandler) CalculatePayroll(c *fiber.Ctx) error {
	periodID := c.Params("id")

	var req models.CalculatePayrollRequest
	if err := c.BodyParser(&req); err != nil {
		// If body is empty, proceed with empty req (calculate all)
		req = models.CalculatePayrollRequest{}
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	period, err := h.payrollService.CalculatePayroll(c.Context(), periodID, &req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case contains(msg, "hanya periode"):
			code = fiber.StatusBadRequest
		case contains(msg, "tidak ada karyawan"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(period, "Penggajian berhasil dihitung"))
}

// ListPayrollItems returns paginated payroll items for a period
// GET /api/payroll/periods/:id/items
func (h *PayrollHandler) ListPayrollItems(c *fiber.Ctx) error {
	periodID := c.Params("id")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 50)

	resp, err := h.payrollService.ListPayrollItems(c.Context(), periodID, page, perPage)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Items,
		"Berhasil memuat item penggajian",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// ApprovePeriod approves a calculated payroll period
// PUT /api/payroll/periods/:id/approve
func (h *PayrollHandler) ApprovePeriod(c *fiber.Ctx) error {
	periodID := c.Params("id")
	userID := database.UserIDFromContext(c.Locals("user_id"))

	period, err := h.payrollService.ApprovePeriod(c.Context(), periodID, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case contains(msg, "hanya periode"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(period, "Periode berhasil disetujui"))
}

// PayPeriod marks an approved period as paid
// PUT /api/payroll/periods/:id/pay
func (h *PayrollHandler) PayPeriod(c *fiber.Ctx) error {
	periodID := c.Params("id")
	userID := database.UserIDFromContext(c.Locals("user_id"))

	period, err := h.payrollService.PayPeriod(c.Context(), periodID, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case contains(msg, "hanya periode"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(period, "Periode berhasil dibayarkan"))
}

// GetPayslip returns a payslip for a specific employee in a period
// GET /api/payroll/payslips/:periodId/:employeeId
func (h *PayrollHandler) GetPayslip(c *fiber.Ctx) error {
	periodID := c.Params("periodId")
	employeeID := c.Params("employeeId")

	payslip, err := h.payrollService.GetPayslip(c.Context(), periodID, employeeID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(payslip, "Berhasil memuat slip gaji"))
}

// ListMyPayslips returns all payslips for the logged-in employee
// GET /api/payroll/my-payslips
func (h *PayrollHandler) ListMyPayslips(c *fiber.Ctx) error {
	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	payslips, err := h.payrollService.ListMyPayslips(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(payslips, "Berhasil memuat daftar slip gaji"))
}

// GetMyPayslip returns the detail of a single payslip for the logged-in employee
// GET /api/payroll/my-payslips/:periodId
func (h *PayrollHandler) GetMyPayslip(c *fiber.Ctx) error {
	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}
	periodID := c.Params("periodId")

	payslip, err := h.payrollService.GetPayslip(c.Context(), periodID, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(payslip, "Berhasil memuat detail slip gaji"))
}
