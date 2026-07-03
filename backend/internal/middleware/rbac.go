package middleware

import (
	"hrms-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
)

// RBAC returns a middleware that checks if the authenticated user's role
// has the required permission (module + action).
//
// Usage:
//
//	employees.Get("/", middleware.RBAC("employee", "read"), handler.ListEmployees)
//	employees.Post("/", middleware.RBAC("employee", "create"), handler.CreateEmployee)
func RBAC(module string, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get role_slug from locals (set by AuthMiddleware)
		roleSlug, ok := c.Locals("role_slug").(string)
		if !ok || roleSlug == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"data":    fiber.Map{},
				"message": "Tidak dapat menentukan peran pengguna",
			})
		}

		// Check permission from database
		hasPermission, err := repository.CheckPermission(c.Context(), roleSlug, module, action)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"data":    fiber.Map{},
				"message": "Gagal memverifikasi akses",
			})
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"data":    fiber.Map{},
				"message": "Anda tidak memiliki akses ke fitur ini",
			})
		}

		return c.Next()
	}
}
