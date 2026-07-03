package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DocumentHandler struct {
	documentService *service.DocumentService
}

func NewDocumentHandler(documentService *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

// ListDocuments returns paginated document list
// GET /api/documents
func (h *DocumentHandler) ListDocuments(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	status := c.Query("status", "")
	employeeID := c.Query("employee_id", "")
	docType := c.Query("doc_type", "")

	// For employees, only show their own documents
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.documentService.List(c.Context(), page, perPage, status, employeeID, docType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Documents,
		"Berhasil memuat data dokumen",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetDocument returns single document detail
// GET /api/documents/:id
func (h *DocumentHandler) GetDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	d, err := h.documentService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	// Employees can only view their own documents
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if d.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses untuk melihat dokumen ini"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(d, "Berhasil memuat detail dokumen"))
}

// CreateDocument creates a new document record
// POST /api/documents
func (h *DocumentHandler) CreateDocument(c *fiber.Ctx) error {
	req := new(models.CreateDocumentReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	d, err := h.documentService.Create(c.Context(), req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "karyawan harus dipilih", "tipe dokumen harus diisi", "judul dokumen harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(d, "Dokumen berhasil ditambahkan"))
}

// VerifyDocument verifies a document
// PUT /api/documents/:id/verify
func (h *DocumentHandler) VerifyDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	d, err := h.documentService.Verify(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "dokumen tidak ditemukan":
			status = fiber.StatusNotFound
		case "dokumen sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(d, "Dokumen berhasil diverifikasi"))
}

// RejectDocument rejects a document
// PUT /api/documents/:id/reject
func (h *DocumentHandler) RejectDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateDocumentStatusReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	d, err := h.documentService.Reject(c.Context(), id, req.RejectionReason, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "dokumen tidak ditemukan":
			status = fiber.StatusNotFound
		case "dokumen sudah diproses":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(d, "Dokumen berhasil ditolak"))
}

// DeleteDocument soft-deletes a document
// DELETE /api/documents/:id
func (h *DocumentHandler) DeleteDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	if err := h.documentService.Delete(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Dokumen berhasil dihapus"))
}
