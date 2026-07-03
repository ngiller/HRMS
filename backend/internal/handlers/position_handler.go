package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type PositionHandler struct {
	positionService *service.PositionService
}

func NewPositionHandler(positionService *service.PositionService) *PositionHandler {
	return &PositionHandler{positionService: positionService}
}

// ListPositions returns paginated position list
// GET /api/positions
func (h *PositionHandler) ListPositions(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.positionService.List(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Positions,
		"Berhasil memuat data posisi jabatan",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAllPositions returns all active positions (for dropdown)
// GET /api/positions/all
func (h *PositionHandler) GetAllPositions(c *fiber.Ctx) error {
	positions, err := h.positionService.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(positions, "Berhasil memuat data posisi jabatan"))
}

// GetPosition returns single position detail
// GET /api/positions/:id
func (h *PositionHandler) GetPosition(c *fiber.Ctx) error {
	id := c.Params("id")

	pos, err := h.positionService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(pos, "Berhasil memuat detail posisi jabatan"))
}

// CreatePosition creates a new position
// POST /api/positions
func (h *PositionHandler) CreatePosition(c *fiber.Ctx) error {
	req := new(models.CreatePositionRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	pos, err := h.positionService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama posisi jabatan harus diisi", "departemen harus dipilih":
			status = fiber.StatusBadRequest
		case "nama posisi jabatan sudah digunakan di departemen ini":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(pos, "Posisi jabatan berhasil ditambahkan"))
}

// UpdatePosition updates an existing position
// PUT /api/positions/:id
func (h *PositionHandler) UpdatePosition(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdatePositionRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	pos, err := h.positionService.Update(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "posisi jabatan tidak ditemukan":
			status = fiber.StatusNotFound
		case "nama posisi jabatan sudah digunakan di departemen ini":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(pos, "Posisi jabatan berhasil diperbarui"))
}

// DeletePosition soft-deletes a position
// DELETE /api/positions/:id
func (h *PositionHandler) DeletePosition(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.positionService.Delete(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "posisi jabatan tidak ditemukan":
			status = fiber.StatusNotFound
		case "posisi jabatan masih memiliki karyawan, tidak dapat dihapus":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Posisi jabatan berhasil dihapus"))
}
