package handlers

import (
	"fmt"
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"
	"os"

	"github.com/gofiber/fiber/v2"
)

type AttendanceRecordHandler struct {
	attendanceRecordService *service.AttendanceRecordService
}

func NewAttendanceRecordHandler(attendanceRecordService *service.AttendanceRecordService) *AttendanceRecordHandler {
	return &AttendanceRecordHandler{attendanceRecordService: attendanceRecordService}
}

func employeeIDFromContext(c *fiber.Ctx) string {
	return database.UserIDFromContext(c.Locals("user_id"))
}

// GetTodayStatus returns today's attendance status for the current user
// GET /api/attendance/today
func (h *AttendanceRecordHandler) GetTodayStatus(c *fiber.Ctx) error {
	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	status, err := h.attendanceRecordService.GetTodayStatus(c.Context(), employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(status, "Berhasil memuat status absensi"))
}

// CheckIn performs attendance check-in
// POST /api/attendance/check-in
func (h *AttendanceRecordHandler) CheckIn(c *fiber.Ctx) error {
	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	req := new(models.CheckInRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	record, err := h.attendanceRecordService.CheckIn(c.Context(), employeeID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "sudah melakukan check-in hari ini":
			status = fiber.StatusConflict
		case "tidak memiliki jadwal kerja, hubungi HR":
			status = fiber.StatusBadRequest
		default:
			fmt.Println("CheckIn Error:", err.Error())
			os.WriteFile("/tmp/checkin_err.log", []byte(err.Error()), 0644)
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(record, "Check-in berhasil"))
}

// CheckOut performs attendance check-out
// PUT /api/attendance/check-out
func (h *AttendanceRecordHandler) CheckOut(c *fiber.Ctx) error {
	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	req := new(models.CheckOutRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	record, err := h.attendanceRecordService.CheckOut(c.Context(), employeeID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "belum melakukan check-in hari ini":
			status = fiber.StatusBadRequest
		case "sudah melakukan check-out hari ini":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(record, "Check-out berhasil"))
}

// ListMyAttendance returns the current user's attendance history
// GET /api/attendance/my-history
func (h *AttendanceRecordHandler) ListMyAttendance(c *fiber.Ctx) error {
	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	dateFrom := c.Query("date_from", "")
	dateTo := c.Query("date_to", "")

	resp, err := h.attendanceRecordService.ListMyHistory(c.Context(), employeeID, page, perPage, dateFrom, dateTo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Records,
		"Berhasil memuat riwayat absensi",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAttendanceRecord returns a single attendance record detail
// GET /api/attendance/:id
func (h *AttendanceRecordHandler) GetAttendanceRecord(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("ID absensi harus diisi"))
	}

	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	record, err := h.attendanceRecordService.GetAttendanceRecord(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Gagal memuat detail absensi"))
	}
	if record == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse("Data absensi tidak ditemukan"))
	}

	// Regular employees can only see their own records
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" && record.EmployeeID != employeeID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses ke data absensi ini"))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(record, "Berhasil memuat detail absensi"))
}

// ExportAttendanceReport exports attendance report as Excel
// GET /api/attendance/report/export
func (h *AttendanceRecordHandler) ExportAttendanceReport(c *fiber.Ctx) error {
	// Regular employees cannot export attendance reports
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses ke laporan absensi"))
	}

	deptID := c.Query("department_id", "")
	employeeID := c.Query("employee_id", "")
	status := c.Query("status", "")
	dateFrom := c.Query("date_from", "")
	dateTo := c.Query("date_to", "")

	fileBytes, err := h.attendanceRecordService.ExportReport(c.Context(), deptID, employeeID, status, dateFrom, dateTo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	c.Response().Header.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header.Set("Content-Disposition", "attachment; filename=laporan-absensi.xlsx")
	return c.Send(fileBytes)
}

// ListAttendanceReport returns attendance report (HR/Manager)
// GET /api/attendance
func (h *AttendanceRecordHandler) ListAttendanceReport(c *fiber.Ctx) error {
	// Regular employees cannot access the report endpoint — use /my-history instead
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses ke laporan absensi"))
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	deptID := c.Query("department_id", "")
	employeeID := c.Query("employee_id", "")
	status := c.Query("status", "")
	dateFrom := c.Query("date_from", "")
	dateTo := c.Query("date_to", "")

	resp, err := h.attendanceRecordService.ListReport(c.Context(), page, perPage, deptID, employeeID, status, dateFrom, dateTo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Records,
		"Berhasil memuat laporan absensi",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}
