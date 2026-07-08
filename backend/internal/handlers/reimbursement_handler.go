package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ReimbursementHandler struct {
	reimbursementService *service.ReimbursementService
}

func NewReimbursementHandler(reimbursementService *service.ReimbursementService) *ReimbursementHandler {
	return &ReimbursementHandler{reimbursementService: reimbursementService}
}

// ListReimbursements returns paginated reimbursement list
// GET /api/reimbursements
func (h *ReimbursementHandler) ListReimbursements(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// For employees, only show their own reimbursements
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.reimbursementService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Reimbursements,
		"Berhasil memuat data reimbursement",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetReimbursement returns single reimbursement detail
// GET /api/reimbursements/:id
func (h *ReimbursementHandler) GetReimbursement(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.reimbursementService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	// Employees can only view their own reimbursements
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat reimbursement ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat detail reimbursement"))
}

// CreateReimbursement creates a new reimbursement request
// POST /api/reimbursements
func (h *ReimbursementHandler) CreateReimbursement(c *fiber.Ctx) error {
	req := new(models.CreateReimbursementReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.reimbursementService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "tipe reimbursement harus diisi", "jumlah reimbursement harus lebih dari 0",
			"deskripsi reimbursement harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Pengajuan reimbursement berhasil diajukan"))
}

// ApproveReimbursement approves a reimbursement request
// PUT /api/reimbursements/:id/approve
func (h *ReimbursementHandler) ApproveReimbursement(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "reimbursement", id, userID, "approve", "")
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "reimbursement tidak ditemukan", "tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Reimbursement berhasil disetujui"))
}

// RejectReimbursement rejects a reimbursement request
// PUT /api/reimbursements/:id/reject
func (h *ReimbursementHandler) RejectReimbursement(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateReimbursementStatusReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "reimbursement", id, userID, "reject", req.RejectionReason)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "reimbursement tidak ditemukan", "tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Reimbursement berhasil ditolak"))
}

// PayReimbursement marks reimbursement as paid
// PUT /api/reimbursements/:id/pay
func (h *ReimbursementHandler) PayReimbursement(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.PayReimbursementReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.reimbursementService.Pay(c.Context(), id, userID, req.PaymentMethod)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "reimbursement tidak ditemukan":
			status = fiber.StatusNotFound
		case "reimbursement harus disetujui terlebih dahulu sebelum dibayar":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Reimbursement berhasil dibayarkan"))
}

// UploadReceipt uploads a receipt file for reimbursement
// POST /api/reimbursements/upload
func (h *ReimbursementHandler) UploadReceipt(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("File tidak ditemukan"))
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Hanya file JPEG dan PNG yang diizinkan"))
	}

	// Max 5MB
	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Ukuran file maksimal 5MB"))
	}

	url, err := h.reimbursementService.UploadReceipt(c.Context(), file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{"url": url}, "File berhasil diupload"))
}

// CancelReimbursement cancels a reimbursement request
// PUT /api/reimbursements/:id/cancel
func (h *ReimbursementHandler) CancelReimbursement(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	err := h.reimbursementService.Cancel(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "reimbursement tidak ditemukan":
			status = fiber.StatusNotFound
		case "reimbursement sudah diproses, tidak dapat dibatalkan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Reimbursement berhasil dibatalkan"))
}
