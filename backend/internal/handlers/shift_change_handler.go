package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ShiftChangeHandler struct {
	shiftChangeService *service.ShiftChangeService
}

func NewShiftChangeHandler(shiftChangeService *service.ShiftChangeService) *ShiftChangeHandler {
	return &ShiftChangeHandler{shiftChangeService: shiftChangeService}
}

// ListShiftChangeRequests returns paginated shift change request list
// GET /api/shift-change-requests
func (h *ShiftChangeHandler) ListShiftChangeRequests(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// For employees, only show their own shift change requests
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.shiftChangeService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.ShiftChangeRequests,
		"Berhasil memuat data permintaan shift",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetShiftChangeRequest returns single shift change request detail
// GET /api/shift-change-requests/:id
func (h *ShiftChangeHandler) GetShiftChangeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.shiftChangeService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	// Employees can only view their own shift change requests
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat permintaan shift ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat detail permintaan shift"))
}

// CreateShiftChangeRequest creates a new shift change request
// POST /api/shift-change-requests
func (h *ShiftChangeHandler) CreateShiftChangeRequest(c *fiber.Ctx) error {
	req := new(models.CreateShiftChangeRequestReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.shiftChangeService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "tipe permintaan harus diisi", "tipe permintaan tidak valid (individual atau swap)",
			"tanggal target harus diisi", "jadwal yang diminta harus dipilih",
			"alasan permintaan harus diisi", "partner swap harus dipilih untuk tipe swap":
			status = fiber.StatusBadRequest
		case "sudah ada permintaan shift pending untuk tanggal ini":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Permintaan shift berhasil diajukan"))
}

// ApproveShiftChangeRequest approves a shift change request
// PUT /api/shift-change-requests/:id/approve
func (h *ShiftChangeHandler) ApproveShiftChangeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "shift_change", id, userID, "approve", "")
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "permintaan shift tidak ditemukan", "tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Permintaan shift berhasil disetujui"))
}

// RejectShiftChangeRequest rejects a shift change request
// PUT /api/shift-change-requests/:id/reject
func (h *ShiftChangeHandler) RejectShiftChangeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateShiftChangeStatusReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "shift_change", id, userID, "reject", req.RejectionReason)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "permintaan shift tidak ditemukan", "tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Permintaan shift berhasil ditolak"))
}

// ConfirmSwapShiftChangeRequest confirms swap partner
// PUT /api/shift-change-requests/:id/confirm-swap
func (h *ShiftChangeHandler) ConfirmSwapShiftChangeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.shiftChangeService.ConfirmSwap(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "permintaan shift tidak ditemukan":
			status = fiber.StatusNotFound
		case "permintaan shift tidak dalam status menunggu konfirmasi partner":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Konfirmasi partner swap berhasil"))
}

// CancelShiftChangeRequest cancels a shift change request
// PUT /api/shift-change-requests/:id/cancel
func (h *ShiftChangeHandler) CancelShiftChangeRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	err := h.shiftChangeService.Cancel(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "permintaan shift tidak ditemukan":
			status = fiber.StatusNotFound
		case "permintaan shift sudah diproses, tidak dapat dibatalkan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Permintaan shift berhasil dibatalkan"))
}
