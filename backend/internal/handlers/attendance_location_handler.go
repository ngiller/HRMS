package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AttendanceLocationHandler struct {
	attendanceLocationService *service.AttendanceLocationService
}

func NewAttendanceLocationHandler(attendanceLocationService *service.AttendanceLocationService) *AttendanceLocationHandler {
	return &AttendanceLocationHandler{attendanceLocationService: attendanceLocationService}
}

// ListAttendanceLocations returns paginated attendance location list
// GET /api/attendance-locations
func (h *AttendanceLocationHandler) ListAttendanceLocations(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")

	resp, err := h.attendanceLocationService.List(c.Context(), page, perPage, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.AttendanceLocations,
		"Berhasil memuat data lokasi absensi",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetAllAttendanceLocations returns all active attendance locations (for dropdown)
// GET /api/attendance-locations/all
func (h *AttendanceLocationHandler) GetAllAttendanceLocations(c *fiber.Ctx) error {
	locations, err := h.attendanceLocationService.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(locations, "Berhasil memuat data lokasi absensi"))
}

// GetAttendanceLocation returns single attendance location detail
// GET /api/attendance-locations/:id
func (h *AttendanceLocationHandler) GetAttendanceLocation(c *fiber.Ctx) error {
	id := c.Params("id")

	loc, err := h.attendanceLocationService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(loc, "Berhasil memuat detail lokasi absensi"))
}

// CreateAttendanceLocation creates a new attendance location
// POST /api/attendance-locations
func (h *AttendanceLocationHandler) CreateAttendanceLocation(c *fiber.Ctx) error {
	req := new(models.CreateAttendanceLocationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	loc, err := h.attendanceLocationService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "nama lokasi absensi harus diisi", "koordinat latitude dan longitude harus diisi":
			status = fiber.StatusBadRequest
		case "nama lokasi absensi sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(loc, "Lokasi absensi berhasil ditambahkan"))
}

// UpdateAttendanceLocation updates an existing attendance location
// PUT /api/attendance-locations/:id
func (h *AttendanceLocationHandler) UpdateAttendanceLocation(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateAttendanceLocationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	loc, err := h.attendanceLocationService.Update(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "lokasi absensi tidak ditemukan":
			status = fiber.StatusNotFound
		case "nama lokasi absensi sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(loc, "Lokasi absensi berhasil diperbarui"))
}

// DeleteAttendanceLocation soft-deletes an attendance location
// DELETE /api/attendance-locations/:id
func (h *AttendanceLocationHandler) DeleteAttendanceLocation(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.attendanceLocationService.Delete(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "lokasi absensi tidak ditemukan":
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Lokasi absensi berhasil dihapus"))
}
