package middleware

import (
	"strings"

	"hrms-backend/internal/config"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"data":    fiber.Map{},
				"message": "Authorization header diperlukan",
			})
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"data":    fiber.Map{},
				"message": "Format token tidak valid",
			})
		}

		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"data":    fiber.Map{},
				"message": "Token tidak valid atau sudah kadaluarsa",
			})
		}

		// Set user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("employee_id", claims.EmployeeID)
		c.Locals("email", claims.Email)
		c.Locals("role_slug", claims.RoleSlug)

		return c.Next()
	}
}

func CORSConfig(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		// Echo back origin for development (supports any port)
		if origin != "" {
			c.Set("Access-Control-Allow-Origin", origin)
		}
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
