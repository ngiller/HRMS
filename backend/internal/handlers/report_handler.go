package handlers

import (
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GET /api/reports/headcount
func (h *ReportHandler) Headcount(c *fiber.Ctx) error {
	year := c.QueryInt("year", 0)
	stats, err := h.reportService.Headcount(c.Context(), year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat laporan headcount"))
}

// GET /api/reports/payroll-summary
func (h *ReportHandler) PayrollSummary(c *fiber.Ctx) error {
	year := c.QueryInt("year", 0)
	stats, err := h.reportService.PayrollSummary(c.Context(), year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat ringkasan penggajian"))
}

// GET /api/reports/attendance-summary
func (h *ReportHandler) AttendanceSummary(c *fiber.Ctx) error {
	year := c.QueryInt("year", 0)
	month := c.QueryInt("month", 0)
	departmentID := c.Query("department_id", "")
	stats, err := h.reportService.AttendanceSummary(c.Context(), year, month, departmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat ringkasan absensi"))
}

// GET /api/reports/loan-summary
func (h *ReportHandler) LoanSummary(c *fiber.Ctx) error {
	stats, err := h.reportService.LoanSummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat ringkasan pinjaman"))
}

// GET /api/reports/leave-summary
func (h *ReportHandler) LeaveSummary(c *fiber.Ctx) error {
	year := c.QueryInt("year", 0)
	departmentID := c.Query("department_id", "")
	stats, err := h.reportService.LeaveSummary(c.Context(), year, departmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat ringkasan cuti"))
}

// GET /api/reports/overtime-summary
func (h *ReportHandler) OvertimeSummary(c *fiber.Ctx) error {
	year := c.QueryInt("year", 0)
	month := c.QueryInt("month", 0)
	stats, err := h.reportService.OvertimeSummary(c.Context(), year, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat ringkasan lembur"))
}
