package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type OvertimeHandler struct {
	overtimeService *service.OvertimeService
}

func NewOvertimeHandler(overtimeService *service.OvertimeService) *OvertimeHandler {
	return &OvertimeHandler{overtimeService: overtimeService}
}

// ListOvertimeRequests returns paginated overtime request list
// GET /api/overtime-requests
func (h *OvertimeHandler) ListOvertimeRequests(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// For employees, only show their own overtime requests
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.overtimeService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.OvertimeRequests,
		"Berhasil memuat data lembur",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetOvertimeRequest returns single overtime request detail
// GET /api/overtime-requests/:id
func (h *OvertimeHandler) GetOvertimeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.overtimeService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	// Employees can only view their own overtime requests
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat pengajuan lembur ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat detail lembur"))
}

// CreateOvertimeRequest creates a new overtime request
// POST /api/overtime-requests
func (h *OvertimeHandler) CreateOvertimeRequest(c *fiber.Ctx) error {
	req := new(models.CreateOvertimeRequestReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.overtimeService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "tanggal lembur harus diisi", "waktu mulai harus diisi",
			"waktu selesai harus diisi", "total jam lembur harus lebih dari 0",
			"tipe lembur harus diisi", "tipe lembur tidak valid (weekday/weekend/holiday)",
			"alasan lembur harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Pengajuan lembur berhasil diajukan"))
}

// ApproveOvertimeRequest approves an overtime request
// PUT /api/overtime-requests/:id/approve
func (h *OvertimeHandler) ApproveOvertimeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.overtimeService.Approve(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan lembur tidak ditemukan":
			status = fiber.StatusNotFound
		case "pengajuan lembur sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Pengajuan lembur berhasil disetujui"))
}

// RejectOvertimeRequest rejects an overtime request
// PUT /api/overtime-requests/:id/reject
func (h *OvertimeHandler) RejectOvertimeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateOvertimeStatusReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.overtimeService.Reject(c.Context(), id, req.RejectionReason, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan lembur tidak ditemukan":
			status = fiber.StatusNotFound
		case "pengajuan lembur sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Pengajuan lembur berhasil ditolak"))
}

// CancelOvertimeRequest cancels an overtime request
// PUT /api/overtime-requests/:id/cancel
func (h *OvertimeHandler) CancelOvertimeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	err := h.overtimeService.Cancel(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan lembur tidak ditemukan":
			status = fiber.StatusNotFound
		case "pengajuan lembur sudah diproses, tidak dapat dibatalkan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Pengajuan lembur berhasil dibatalkan"))
}

// GetOvertimeCalculation returns overtime calculation for an approved request
// GET /api/overtime-requests/:id/calculation
func (h *OvertimeHandler) GetOvertimeCalculation(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.overtimeService.GetCalculation(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat perhitungan lembur"))
}
