package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type RosterHandler struct {
	rosterService *service.RosterService
}

func NewRosterHandler(rosterService *service.RosterService) *RosterHandler {
	return &RosterHandler{rosterService: rosterService}
}

// ============================================================
// DEPARTMENT ROSTERS
// ============================================================

// ListRosters returns paginated roster list
// GET /api/rosters
func (h *RosterHandler) ListRosters(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	departmentID := c.Query("department_id", "")

	resp, err := h.rosterService.List(c.Context(), page, perPage, departmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Rosters,
		"Berhasil memuat data roster",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetRoster returns single roster detail
// GET /api/rosters/:id
func (h *RosterHandler) GetRoster(c *fiber.Ctx) error {
	id := c.Params("id")

	roster, err := h.rosterService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(roster, "Berhasil memuat detail roster"))
}

// CreateRoster creates a new roster
// POST /api/rosters
func (h *RosterHandler) CreateRoster(c *fiber.Ctx) error {
	req := new(models.CreateDepartmentRosterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	roster, err := h.rosterService.Create(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "departemen harus dipilih", "nama roster harus diisi", "bulan tidak valid (1-12)", "tahun tidak valid":
			status = fiber.StatusBadRequest
		case "nama roster sudah digunakan di departemen ini":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(roster, "Roster berhasil ditambahkan"))
}

// UpdateRoster updates an existing roster
// PUT /api/rosters/:id
func (h *RosterHandler) UpdateRoster(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.UpdateDepartmentRosterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	roster, err := h.rosterService.Update(c.Context(), id, req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "roster tidak ditemukan":
			status = fiber.StatusNotFound
		case "nama roster sudah digunakan":
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(roster, "Roster berhasil diperbarui"))
}

// DeleteRoster soft-deletes a roster
// DELETE /api/rosters/:id
func (h *RosterHandler) DeleteRoster(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.rosterService.Delete(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "roster tidak ditemukan":
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Roster berhasil dihapus"))
}

// ============================================================
// ROSTER ENTRIES
// ============================================================

// ListRosterEntries returns all entries for a roster
// GET /api/rosters/:id/entries
func (h *RosterHandler) ListRosterEntries(c *fiber.Ctx) error {
	id := c.Params("id")

	entries, err := h.rosterService.ListEntries(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(entries, "Berhasil memuat entri roster"))
}

// GetRosterCalendar returns calendar view for a roster
// GET /api/rosters/:id/calendar
func (h *RosterHandler) GetRosterCalendar(c *fiber.Ctx) error {
	id := c.Params("id")

	calendar, err := h.rosterService.GetCalendar(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(calendar, "Berhasil memuat kalender roster"))
}

// CreateRosterEntry creates a single roster entry
// POST /api/roster-entries
func (h *RosterHandler) CreateRosterEntry(c *fiber.Ctx) error {
	req := new(models.CreateRosterEntryRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	entry, err := h.rosterService.CreateEntry(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "roster harus dipilih", "karyawan harus dipilih", "tanggal harus diisi", "shift harus dipilih":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(entry, "Entri roster berhasil ditambahkan"))
}

// BulkCreateRosterEntries creates multiple roster entries at once
// POST /api/rosters/:id/entries/bulk
func (h *RosterHandler) BulkCreateRosterEntries(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.BulkRosterEntryRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}
	req.RosterID = id

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.rosterService.BulkCreateEntries(c.Context(), req, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "roster harus dipilih", "minimal 1 entri harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(fiber.Map{}, "Entri roster berhasil disimpan"))
}

// DeleteRosterEntry deletes a roster entry
// DELETE /api/roster-entries/:id
func (h *RosterHandler) DeleteRosterEntry(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.rosterService.DeleteEntry(c.Context(), id, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Entri roster berhasil dihapus"))
}

// DeleteEmployeeRosterEntries deletes all entries for an employee in a roster
// DELETE /api/rosters/:id/employees/:employee_id
func (h *RosterHandler) DeleteEmployeeRosterEntries(c *fiber.Ctx) error {
	id := c.Params("id")
	employeeID := c.Params("employee_id")

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.rosterService.DeleteEmployeeEntries(c.Context(), id, employeeID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Karyawan berhasil dihapus dari roster"))
}

