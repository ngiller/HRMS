package handlers

import (
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type LoanHandler struct {
	loanService *service.LoanService
}

func NewLoanHandler(loanService *service.LoanService) *LoanHandler {
	return &LoanHandler{loanService: loanService}
}

// ListLoans returns paginated loan list
// GET /api/loans
func (h *LoanHandler) ListLoans(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// For employees, only show their own loans
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.loanService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Loans,
		"Berhasil memuat data pinjaman",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetLoan returns single loan detail
// GET /api/loans/:id
func (h *LoanHandler) GetLoan(c *fiber.Ctx) error {
	id := c.Params("id")

	l, err := h.loanService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	// Employees can only view their own loans
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if l.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat pinjaman ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(l, "Berhasil memuat detail pinjaman"))
}

// CreateLoan creates a new loan
// POST /api/loans
func (h *LoanHandler) CreateLoan(c *fiber.Ctx) error {
	req := new(models.CreateLoanRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	l, err := h.loanService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "karyawan harus diisi", "jumlah pinjaman harus lebih dari 0",
			"tenor pinjaman 1-24 bulan", "tipe pinjaman harus diisi",
			"tipe pinjaman tidak valid (regular/emergency/education)",
			"tujuan pinjaman harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(l, "Pinjaman berhasil diajukan"))
}

// ApproveLoan approves a loan
// PUT /api/loans/:id/approve
func (h *LoanHandler) ApproveLoan(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "loan", id, userID, "approve", "")
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pinjaman tidak ditemukan", "data pinjaman tidak ditemukan",
			"tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Pinjaman berhasil disetujui"))
}

// RejectLoan rejects a loan
// PUT /api/loans/:id/reject
func (h *LoanHandler) RejectLoan(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateLoanStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	wfSvc := service.NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(c.Context(), "loan", id, userID, "reject", req.RejectionReason)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pinjaman tidak ditemukan", "data pinjaman tidak ditemukan",
			"tracking approval tidak ditemukan":
			status = fiber.StatusNotFound
		case "request ini sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Pinjaman berhasil ditolak"))
}

// CancelLoan cancels a loan request
// PUT /api/loans/:id/cancel
func (h *LoanHandler) CancelLoan(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	err := h.loanService.Cancel(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pinjaman tidak ditemukan":
			status = fiber.StatusNotFound
		case "pinjaman sudah diproses, tidak dapat dibatalkan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Pinjaman berhasil dibatalkan"))
}

// DisburseLoan disburses a loan (cairkan)
// PUT /api/loans/:id/disburse
func (h *LoanHandler) DisburseLoan(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateLoanStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	date := req.DisburseDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	l, err := h.loanService.Disburse(c.Context(), id, userID, date)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "pinjaman tidak ditemukan":
			status = fiber.StatusNotFound
		case "pinjaman harus disetujui terlebih dahulu":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(l, "Pinjaman berhasil dicairkan"))
}

// GetLoanStats returns loan statistics
// GET /api/loans/stats
func (h *LoanHandler) GetLoanStats(c *fiber.Ctx) error {
	stats, err := h.loanService.Stats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat statistik pinjaman"))
}
