package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type PositionGradeHandler struct {
	positionGradeService *service.PositionGradeService
}

func NewPositionGradeHandler(positionGradeService *service.PositionGradeService) *PositionGradeHandler {
	return &PositionGradeHandler{positionGradeService: positionGradeService}
}

// ListPositionGrades returns paginated position grade list
// GET /api/position-grades
func (h *PositionGradeHandler) ListPositionGrades(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.positionGradeService.List(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.PositionGrades,
		"Berhasil memuat data golongan jabatan",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAllPositionGrades returns all active position grades (for dropdown)
// GET /api/position-grades/all
func (h *PositionGradeHandler) GetAllPositionGrades(c *fiber.Ctx) error {
	grades, err := h.positionGradeService.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(grades, "Berhasil memuat data golongan jabatan"))
}

// GetPositionGrade returns single position grade detail
// GET /api/position-grades/:id
func (h *PositionGradeHandler) GetPositionGrade(c *fiber.Ctx) error {
	id := c.Params("id")

	g, err := h.positionGradeService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(g, "Berhasil memuat detail golongan jabatan"))
}

// CreatePositionGrade creates a new position grade
// POST /api/position-grades
func (h *PositionGradeHandler) CreatePositionGrade(c *fiber.Ctx) error {
	req := new(models.CreatePositionGradeRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	g, err := h.positionGradeService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama golongan jabatan harus diisi", "level harus diisi (minimal 1)":
			status = fiber.StatusBadRequest
		case "nama golongan jabatan sudah digunakan", "level sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(g, "Golongan jabatan berhasil ditambahkan"))
}

// UpdatePositionGrade updates an existing position grade
// PUT /api/position-grades/:id
func (h *PositionGradeHandler) UpdatePositionGrade(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdatePositionGradeRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	g, err := h.positionGradeService.Update(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "golongan jabatan tidak ditemukan":
			status = fiber.StatusNotFound
		case "nama golongan jabatan sudah digunakan", "level sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(g, "Golongan jabatan berhasil diperbarui"))
}

// DeletePositionGrade deletes a position grade
// DELETE /api/position-grades/:id
func (h *PositionGradeHandler) DeletePositionGrade(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.positionGradeService.Delete(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "golongan jabatan tidak ditemukan":
			status = fiber.StatusNotFound
		case "golongan jabatan masih digunakan oleh posisi, tidak dapat dihapus":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Golongan jabatan berhasil dihapus"))
}
