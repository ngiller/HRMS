package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ManualAttendanceHandler struct {
	manualAttendanceService *service.ManualAttendanceService
}

func NewManualAttendanceHandler(manualAttendanceService *service.ManualAttendanceService) *ManualAttendanceHandler {
	return &ManualAttendanceHandler{manualAttendanceService: manualAttendanceService}
}

// CreateManualAttendance creates a new manual attendance request
// POST /api/manual-attendance
func (h *ManualAttendanceHandler) CreateManualAttendance(c *fiber.Ctx) error {
	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	req := new(models.CreateManualAttendanceRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	r, err := h.manualAttendanceService.Create(c.Context(), employeeID, req)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch err.Error() {
		case "tanggal harus diisi", "alasan harus diisi", "setidaknya satu waktu (check-in atau check-out) harus diisi":
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Pengajuan absensi manual berhasil dibuat"))
}

// ListManualAttendance lists all manual attendance requests
// GET /api/manual-attendance
func (h *ManualAttendanceHandler) ListManualAttendance(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// Regular employees can only see their own data
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.manualAttendanceService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Requests,
		"Berhasil memuat daftar pengajuan absensi manual",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetManualAttendance returns a single manual attendance request
// GET /api/manual-attendance/:id
func (h *ManualAttendanceHandler) GetManualAttendance(c *fiber.Ctx) error {
	id := c.Params("id")

	roleSlug, _ := c.Locals("role_slug").(string)

	r, err := h.manualAttendanceService.Get(c.Context(), id)
	if err != nil {
		code := fiber.StatusInternalServerError
		if err.Error() == "pengajuan absensi manual tidak ditemukan" {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	// Regular employees can only see their own requests
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat pengajuan ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat pengajuan absensi manual"))
}

// ApproveManualAttendance approves a manual attendance request
// PUT /api/manual-attendance/:id/approve
func (h *ManualAttendanceHandler) ApproveManualAttendance(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.manualAttendanceService.Approve(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan absensi manual tidak ditemukan":
			code = fiber.StatusNotFound
		case "pengajuan absensi manual sudah diproses":
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Pengajuan absensi manual berhasil disetujui"))
}

// RejectManualAttendance rejects a manual attendance request
// PUT /api/manual-attendance/:id/reject
func (h *ManualAttendanceHandler) RejectManualAttendance(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	var req struct {
		RejectionReason string `json:"rejection_reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		req.RejectionReason = ""
	}

	r, err := h.manualAttendanceService.Reject(c.Context(), id, userID, req.RejectionReason)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan absensi manual tidak ditemukan":
			code = fiber.StatusNotFound
		case "pengajuan absensi manual sudah diproses":
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Pengajuan absensi manual berhasil ditolak"))
}

// CancelManualAttendance cancels a manual attendance request
// PUT /api/manual-attendance/:id/cancel
func (h *ManualAttendanceHandler) CancelManualAttendance(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	err := h.manualAttendanceService.Cancel(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		if err.Error() == "pengajuan absensi manual tidak ditemukan" {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Pengajuan absensi manual berhasil dibatalkan"))
}
