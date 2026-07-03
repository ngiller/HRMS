package handlers

import (
	"errors"
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type KPIHandler struct {
	kpiService *service.KPIService
}

func NewKPIHandler(kpiService *service.KPIService) *KPIHandler {
	return &KPIHandler{kpiService: kpiService}
}

// ListKPITemplates returns paginated KPI template list
// GET /api/kpi/templates
func (h *KPIHandler) ListKPITemplates(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	year := c.QueryInt("year", 0)

	resp, err := h.kpiService.ListTemplates(c.Context(), page, perPage, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Templates,
		"Berhasil memuat template KPI",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetKPITemplate returns single KPI template detail
// GET /api/kpi/templates/:id
func (h *KPIHandler) GetKPITemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	t, err := h.kpiService.GetTemplate(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(t, "Berhasil memuat template KPI"))
}

// ListKPIReviews returns paginated KPI review list
// GET /api/kpi/reviews
func (h *KPIHandler) ListKPIReviews(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	// For employees, only show their own KPI reviews
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.kpiService.ListReviews(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Reviews,
		"Berhasil memuat review KPI",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetKPIReview returns single KPI review detail
// GET /api/kpi/reviews/:id
func (h *KPIHandler) GetKPIReview(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.kpiService.GetReview(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	// Employees can only view their own KPI reviews
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if r.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat review KPI ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat review KPI"))
}

// CreateKPIReview creates a new KPI review
// POST /api/kpi/reviews
func (h *KPIHandler) CreateKPIReview(c *fiber.Ctx) error {
	req := new(models.CreateKPIReviewRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	r, err := h.kpiService.CreateReview(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		
		rootErr := err
		for errors.Unwrap(rootErr) != nil {
			rootErr = errors.Unwrap(rootErr)
		}
		
		switch rootErr.Error() {
		case "karyawan harus diisi", "template KPI harus diisi",
			"periode harus diisi", "tahun harus >= 2024":
			status = fiber.StatusBadRequest
		case "review KPI untuk karyawan pada periode dan tahun tersebut sudah ada":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(rootErr.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Review KPI berhasil dibuat"))
}

// CreateKPITemplate creates a new KPI template with indicators
// POST /api/kpi/templates
func (h *KPIHandler) CreateKPITemplate(c *fiber.Ctx) error {
	req := new(models.CreateKPITemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	t, err := h.kpiService.CreateTemplate(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError

		rootErr := err
		for errors.Unwrap(rootErr) != nil {
			rootErr = errors.Unwrap(rootErr)
		}

		switch rootErr.Error() {
		case "judul template harus diisi", "tipe periode harus diisi",
			"tahun harus >= 2024", "minimal 1 indikator harus ditambahkan",
			"nama indikator harus diisi", "target indikator harus lebih dari 0",
			"bobot indikator harus lebih dari 0", "total bobot indikator harus 100%":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(rootErr.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(t, "Template KPI berhasil dibuat"))
}

// UpdateKPITemplate updates a KPI template
// PUT /api/kpi/templates/:id
func (h *KPIHandler) UpdateKPITemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateKPITemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	t, err := h.kpiService.UpdateTemplate(c.Context(), id, req)
	if err != nil {
		status := fiber.StatusInternalServerError

		rootErr := err
		for errors.Unwrap(rootErr) != nil {
			rootErr = errors.Unwrap(rootErr)
		}

		switch rootErr.Error() {
		case "judul template harus diisi", "tipe periode harus diisi",
			"tahun harus >= 2024", "minimal 1 indikator harus ditambahkan",
			"nama indikator harus diisi", "target indikator harus lebih dari 0",
			"bobot indikator harus lebih dari 0", "total bobot indikator harus 100%",
			"template KPI tidak ditemukan":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(rootErr.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(t, "Template KPI berhasil diupdate"))
}

// DeleteKPITemplate soft-deletes a KPI template
// DELETE /api/kpi/templates/:id
func (h *KPIHandler) DeleteKPITemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.kpiService.DeleteTemplate(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "template KPI tidak ditemukan" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Template KPI berhasil dihapus"))
}
