package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	roleService *service.RoleService
}

func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// ListRoles returns paginated role list
// GET /api/roles
func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.roleService.ListRoles(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Roles,
		"Berhasil memuat data role",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetRole returns single role detail
// GET /api/roles/:id
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	id := c.Params("id")

	role, err := h.roleService.GetRole(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(role, "Berhasil memuat detail role"))
}

// CreateRole creates a new role
// POST /api/roles
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	req := new(models.CreateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	role, err := h.roleService.CreateRole(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama role harus diisi", "slug role harus diisi":
			status = fiber.StatusBadRequest
		case "nama role sudah digunakan", "slug role sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(role, "Role berhasil ditambahkan"))
}

// UpdateRole updates an existing role
// PUT /api/roles/:id
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	role, err := h.roleService.UpdateRole(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "role tidak ditemukan":
			status = fiber.StatusNotFound
		case "nama role sudah digunakan", "slug role sudah digunakan":
			status = fiber.StatusConflict
		case "nama role tidak boleh kosong", "slug role tidak boleh kosong":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	if service.GetSSEHub() != nil {
		service.GetSSEHub().BroadcastAll(service.SSEEvent{
			Type: "role_update",
			Data: map[string]interface{}{
				"role_id": id,
				"slug":    role.Slug,
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(role, "Role berhasil diperbarui"))
}

// DeleteRole deletes a role (only non-system roles)
// DELETE /api/roles/:id
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.roleService.DeleteRole(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "role tidak ditemukan":
			status = fiber.StatusNotFound
		case "role sistem tidak dapat dihapus", "role masih digunakan oleh karyawan, tidak dapat dihapus":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Role berhasil dihapus"))
}

// GetPermissionTemplate returns available modules & actions for permission editing
// GET /api/roles/permissions/template
func (h *RoleHandler) GetPermissionTemplate(c *fiber.Ctx) error {
	template := h.roleService.GetPermissionTemplate(c.Context())
	return c.Status(fiber.StatusOK).JSON(SuccessResponse(template, "Berhasil memuat template permission"))
}
