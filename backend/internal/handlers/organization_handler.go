package handlers

import (
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type OrganizationHandler struct {
	orgService *service.OrganizationService
}

func NewOrganizationHandler(orgService *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

// GetTree returns the full organization tree
// GET /api/organization/tree
func (h *OrganizationHandler) GetTree(c *fiber.Ctx) error {
	tree, err := h.orgService.GetTree(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(tree, "Berhasil memuat struktur organisasi"))
}
