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

// GetTree returns the organization tree filtered by user role
// GET /api/organization/tree
func (h *OrganizationHandler) GetTree(c *fiber.Ctx) error {
	roleSlug, _ := c.Locals("role_slug").(string)
	userID, _ := c.Locals("user_id").(string)

	tree, err := h.orgService.GetTree(c.Context(), roleSlug, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(tree, "Berhasil memuat struktur organisasi"))
}
