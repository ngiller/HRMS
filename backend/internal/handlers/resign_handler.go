package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ResignHandler struct {
	resignService *service.ResignService
}

func NewResignHandler(resignService *service.ResignService) *ResignHandler {
	return &ResignHandler{resignService: resignService}
}

// CreateResign creates a new resignation request
// POST /api/resign
func (h *ResignHandler) CreateResign(c *fiber.Ctx) error {
	employeeID := employeeIDFromContext(c)
	if employeeID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	req := new(models.CreateResignRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	r, err := h.resignService.Create(c.Context(), employeeID, req)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch err.Error() {
		case "tanggal terakhir kerja harus diisi", "alasan resign harus diisi":
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(r, "Pengajuan resign berhasil dibuat"))
}

// ListResigns lists all resignation requests
// GET /api/resign
func (h *ResignHandler) ListResigns(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")

	resp, err := h.resignService.List(c.Context(), page, perPage, status, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Resigns,
		"Berhasil memuat daftar pengajuan resign",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetResign returns a single resignation request
// GET /api/resign/:id
func (h *ResignHandler) GetResign(c *fiber.Ctx) error {
	id := c.Params("id")

	r, err := h.resignService.Get(c.Context(), id)
	if err != nil {
		code := fiber.StatusInternalServerError
		if err.Error() == "pengajuan resign tidak ditemukan" {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Berhasil memuat pengajuan resign"))
}

// ApproveResign approves a resignation request
// PUT /api/resign/:id/approve
func (h *ResignHandler) ApproveResign(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	r, err := h.resignService.Approve(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan resign tidak ditemukan":
			code = fiber.StatusNotFound
		case "pengajuan resign sudah diproses":
			code = fiber.StatusConflict
		case "semua item exit clearance harus diselesaikan terlebih dahulu":
			code = fiber.StatusBadRequest
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Resign berhasil disetujui"))
}

// RejectResign rejects a resignation request
// PUT /api/resign/:id/reject
func (h *ResignHandler) RejectResign(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	var req struct {
		RejectionReason string `json:"rejection_reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		req.RejectionReason = ""
	}

	r, err := h.resignService.Reject(c.Context(), id, userID, req.RejectionReason)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch err.Error() {
		case "pengajuan resign tidak ditemukan":
			code = fiber.StatusNotFound
		case "pengajuan resign sudah diproses":
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(r, "Pengajuan resign berhasil ditolak"))
}

// ListClearanceItems lists exit clearance items for a resignation
// GET /api/resign/:id/clearance
func (h *ResignHandler) ListClearanceItems(c *fiber.Ctx) error {
	resignID := c.Params("id")

	items, err := h.resignService.ListClearanceItems(c.Context(), resignID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(items.Items, "Berhasil memuat daftar clearance"))
}

// UpdateClearanceItem updates an exit clearance item
// PUT /api/resign/clearance/:itemId
func (h *ResignHandler) UpdateClearanceItem(c *fiber.Ctx) error {
	itemID := c.Params("itemId")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	var req struct {
		IsChecked bool `json:"is_checked"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	item, err := h.resignService.UpdateClearanceItem(c.Context(), itemID, userID, req.IsChecked)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(item, "Item clearance berhasil diperbarui"))
}
