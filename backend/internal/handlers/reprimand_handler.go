package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ReprimandHandler struct {
	reprimandService *service.ReprimandService
}

func NewReprimandHandler(reprimandService *service.ReprimandService) *ReprimandHandler {
	return &ReprimandHandler{reprimandService: reprimandService}
}

// ListReprimands returns paginated reprimand list
// GET /api/reprimands
func (h *ReprimandHandler) ListReprimands(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.reprimandService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Reprimands,
		"Berhasil memuat data surat peringatan",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetReprimand returns single reprimand detail
// GET /api/reprimands/:id
func (h *ReprimandHandler) GetReprimand(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.reprimandService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat detail surat peringatan"))
}

// CreateReprimand creates a new reprimand
// POST /api/reprimands
func (h *ReprimandHandler) CreateReprimand(c *fiber.Ctx) error {
	req := new(models.CreateReprimandRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.reprimandService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "karyawan harus diisi", "tipe surat peringatan harus diisi",
			"judul surat peringatan harus diisi",
			"tipe surat peringatan tidak valid (verbal/sp1/sp2/sp3)":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Surat peringatan berhasil diterbitkan"))
}

// AcknowledgeReprimand acknowledges a reprimand (employee confirms receipt)
// PUT /api/reprimands/:id/acknowledge
func (h *ReprimandHandler) AcknowledgeReprimand(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateReprimandStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.reprimandService.Acknowledge(c.Context(), id, userID, req.AcknowledgmentNote)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "surat peringatan tidak ditemukan atau sudah diakui" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Surat peringatan berhasil diakui"))
}
