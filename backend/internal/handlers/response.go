package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse returns a consistent success JSON response.
// Usage: c.JSON(SuccessResponse(data, "Pesan berhasil"))
func SuccessResponse(data interface{}, message string) fiber.Map {
	return fiber.Map{
		"success": true,
		"data":    data,
		"message": message,
	}
}

// SuccessResponseWithMeta returns a consistent success JSON response with pagination meta.
// Usage: c.JSON(SuccessResponseWithMeta(data, "Pesan berhasil", meta))
func SuccessResponseWithMeta(data interface{}, message string, meta interface{}) fiber.Map {
	return fiber.Map{
		"success": true,
		"data":    data,
		"message": message,
		"meta":    meta,
	}
}

// ErrorResponse returns a consistent error JSON response.
// Usage: c.Status(400).JSON(ErrorResponse("Pesan error"))
func ErrorResponse(message string) fiber.Map {
	return fiber.Map{
		"success": false,
		"data":    fiber.Map{},
		"message": message,
	}
}

// PaginationMeta builds a meta map for paginated list endpoints.
func PaginationMeta(total, page, perPage int) fiber.Map {
	return fiber.Map{
		"total":    total,
		"page":     page,
		"per_page": perPage,
	}
}
