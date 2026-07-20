package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type LeaveHandler struct {
	leaveService *service.LeaveService
}

func NewLeaveHandler(leaveService *service.LeaveService) *LeaveHandler {
	return &LeaveHandler{leaveService: leaveService}
}

// ─── Leave Types ──────────────────────────────────────────────

// GetLeaveTypes returns all available leave types
// GET /api/leaves/types
func (h *LeaveHandler) GetLeaveTypes(c *fiber.Ctx) error {
	types, err := h.leaveService.GetLeaveTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(types, "Berhasil memuat jenis cuti"))
}

// ─── Leave Balances ───────────────────────────────────────────

// GetMyLeaveBalances returns leave balances for the current user
// GET /api/leaves/my-balances
func (h *LeaveHandler) GetMyLeaveBalances(c *fiber.Ctx) error {
	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	resp, err := h.leaveService.GetMyBalances(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp.Balances, "Berhasil memuat sisa cuti"))
}

// GetAllLeaveBalances returns all leave balances (for HR)
// GET /api/leaves/balances
func (h *LeaveHandler) GetAllLeaveBalances(c *fiber.Ctx) error {
	yearStr := c.Query("year", "2026")
	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = 2026
	}

	resp, err := h.leaveService.GetAllBalances(c.Context(), year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp.Balances, "Berhasil memuat sisa cuti"))
}

// ─── Leave Requests ───────────────────────────────────────────

// ListLeaveRequests returns paginated leave request list
// GET /api/leaves
func (h *LeaveHandler) ListLeaveRequests(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.leaveService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.LeaveRequests,
		"Berhasil memuat data cuti",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetLeaveRequest returns a single leave request detail
// GET /api/leaves/:id
func (h *LeaveHandler) GetLeaveRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.leaveService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat cuti ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat detail cuti"))
}

// CreateLeaveRequest creates a new leave request
// POST /api/leaves
func (h *LeaveHandler) CreateLeaveRequest(c *fiber.Ctx) error {
	req := new(models.CreateLeaveRequestReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.leaveService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "jenis cuti harus dipilih", "tanggal mulai harus diisi",
			"tanggal selesai harus diisi", "jumlah hari harus lebih dari 0",
			"alasan cuti harus diisi",
			"cuti hamil/melahirkan hanya untuk karyawan perempuan",
			"cuti hamil/melahirkan hanya untuk karyawan yang sudah menikah",
			"cuti keguguran hanya untuk karyawan perempuan",
			"cuti menikah hanya untuk karyawan yang belum menikah (lajang)":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Pengajuan cuti berhasil diajukan"))
}

// ApproveLeaveRequest approves a leave request
// PUT /api/leaves/:id/approve
func (h *LeaveHandler) ApproveLeaveRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "leave", id, userID, "approve", "")
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan cuti tidak ditemukan", "data cuti tidak ditemukan",
			"tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Pengajuan cuti berhasil disetujui"))
}

// RejectLeaveRequest rejects a leave request
// PUT /api/leaves/:id/reject
func (h *LeaveHandler) RejectLeaveRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateLeaveStatusReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "leave", id, userID, "reject", req.RejectionReason)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan cuti tidak ditemukan", "data cuti tidak ditemukan",
			"tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Pengajuan cuti berhasil ditolak"))
}

// CancelLeaveRequest cancels a leave request
// PUT /api/leaves/:id/cancel
func (h *LeaveHandler) CancelLeaveRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateLeaveStatusReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	err := h.leaveService.Cancel(c.Context(), id, userID, req.CancelReason)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan cuti tidak ditemukan":
			status = fiber.StatusNotFound
		case "pengajuan cuti sudah diproses, tidak dapat dibatalkan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Pengajuan cuti berhasil dibatalkan"))
}
