package handlers

import (
	"strings"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// MutationHandler handles HTTP requests for employee mutations & promotions
type MutationHandler struct {
	svc *service.MutationService
}

// NewMutationHandler creates a new MutationHandler
func NewMutationHandler(svc *service.MutationService) *MutationHandler {
	return &MutationHandler{svc: svc}
}

// ListMutations returns paginated mutation list
// GET /api/mutations
func (h *MutationHandler) ListMutations(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// Jika role employee, hanya lihat data sendiri
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.svc.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Mutations,
		"Berhasil memuat data mutasi",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetMutation returns a single mutation detail
// GET /api/mutations/:id
func (h *MutationHandler) GetMutation(c *fiber.Ctx) error {
	id := c.Params("id")

	m, err := h.svc.Get(c.Context(), id)
	if err != nil {
		code := fiber.StatusInternalServerError
		if strings.Contains(err.Error(), "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	// Jika role employee, hanya lihat data sendiri
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if m.EmployeeID != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(m, "Berhasil memuat detail mutasi"))
}

// CreateMutation creates a new mutation request
// POST /api/mutations
func (h *MutationHandler) CreateMutation(c *fiber.Ctx) error {
	req := new(models.CreateMutationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	m, err := h.svc.Create(c.Context(), req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case strings.Contains(msg, "harus diisi"), strings.Contains(msg, "tidak valid"):
			code = fiber.StatusBadRequest
		case strings.Contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(m, "Mutasi berhasil diajukan"))
}

// ApproveMutation approves a mutation
// PUT /api/mutations/:id/approve
func (h *MutationHandler) ApproveMutation(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	m, err := h.svc.Approve(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case strings.Contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case strings.Contains(msg, "sudah diproses"):
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(m, "Mutasi berhasil disetujui"))
}

// RejectMutation rejects a mutation
// PUT /api/mutations/:id/reject
func (h *MutationHandler) RejectMutation(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	req := new(struct {
		RejectionReason string `json:"rejection_reason"`
	})
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	m, err := h.svc.Reject(c.Context(), id, userID, req.RejectionReason)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case strings.Contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case strings.Contains(msg, "sudah diproses"):
			code = fiber.StatusConflict
		case strings.Contains(msg, "harus diisi"):
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(m, "Mutasi berhasil ditolak"))
}

// CancelMutation cancels a pending mutation
// PUT /api/mutations/:id/cancel
func (h *MutationHandler) CancelMutation(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	if err := h.svc.Cancel(c.Context(), id, userID); err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case strings.Contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case strings.Contains(msg, "sudah diproses"):
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Mutasi berhasil dibatalkan"))
}
