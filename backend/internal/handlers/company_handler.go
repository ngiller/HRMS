package handlers

import (
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type CompanyHandler struct {
	companyService *service.CompanyService
}

func NewCompanyHandler(companyService *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

// GetSettings returns company settings including BPJS config
// GET /api/company/settings
func (h *CompanyHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.companyService.GetSettings(c.Context())
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(settings, "Berhasil memuat pengaturan perusahaan"))
}

// UpdateSettings updates company settings including BPJS config
// PUT /api/company/settings
func (h *CompanyHandler) UpdateSettings(c *fiber.Ctx) error {
	var req models.UpdateCompanySettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	settings, err := h.companyService.UpdateSettings(c.Context(), &req)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak boleh kosong") {
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(settings, "Pengaturan berhasil diupdate"))
}

// GetEmployeeBPJSConfig returns BPJS config for a specific employee
// GET /api/employees/:id/bpjs-config
func (h *CompanyHandler) GetEmployeeBPJSConfig(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	cfg, err := repository.GetEmployeeBPJSConfig(c.Context(), employeeID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(cfg, "Berhasil memuat konfigurasi BPJS karyawan"))
}

// UpdateEmployeeBPJSConfig updates BPJS config for a specific employee
// PUT /api/employees/:id/bpjs-config
func (h *CompanyHandler) UpdateEmployeeBPJSConfig(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	var req struct {
		BPJSConfig *models.BPJSConfig `json:"bpjs_config"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	if err := repository.UpdateEmployeeBPJSConfig(c.Context(), employeeID, req.BPJSConfig); err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(nil, "Konfigurasi BPJS berhasil diupdate"))
}

// GetEmployeeTaxConfig returns Tax config for a specific employee
// GET /api/employees/:id/tax-config
func (h *CompanyHandler) GetEmployeeTaxConfig(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	cfg, err := repository.GetEmployeeTaxConfig(c.Context(), employeeID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(cfg, "Berhasil memuat konfigurasi pajak karyawan"))
}

// UpdateEmployeeTaxConfig updates Tax config for a specific employee
// PUT /api/employees/:id/tax-config
func (h *CompanyHandler) UpdateEmployeeTaxConfig(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	var req struct {
		TaxConfig *models.TaxConfig `json:"tax_config"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	if err := repository.UpdateEmployeeTaxConfig(c.Context(), employeeID, req.TaxConfig); err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(nil, "Konfigurasi pajak berhasil diupdate"))
}

// GetEmployeeOvertimeConfig returns Overtime config for a specific employee
// GET /api/employees/:id/overtime-config
func (h *CompanyHandler) GetEmployeeOvertimeConfig(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	cfg, err := repository.GetEmployeeOvertimeConfig(c.Context(), employeeID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(cfg, "Berhasil memuat konfigurasi lembur karyawan"))
}

// UpdateEmployeeOvertimeConfig updates Overtime config for a specific employee
// PUT /api/employees/:id/overtime-config
func (h *CompanyHandler) UpdateEmployeeOvertimeConfig(c *fiber.Ctx) error {
	employeeID := c.Params("id")

	var req struct {
		OvertimeConfig *models.OvertimeConfig `json:"overtime_config"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	if err := repository.UpdateEmployeeOvertimeConfig(c.Context(), employeeID, req.OvertimeConfig); err != nil {
		code := fiber.StatusInternalServerError
		if contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(nil, "Konfigurasi lembur berhasil diupdate"))
}

