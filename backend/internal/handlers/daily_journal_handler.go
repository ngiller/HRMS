package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DailyJournalHandler struct {
	dailyJournalService *service.DailyJournalService
}

func NewDailyJournalHandler(dailyJournalService *service.DailyJournalService) *DailyJournalHandler {
	return &DailyJournalHandler{dailyJournalService: dailyJournalService}
}

// ListJournals returns paginated daily journal list
// GET /api/daily-journals
func (h *DailyJournalHandler) ListJournals(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	departmentID := c.Query("department_id", "")
	employeeID := c.Query("employee_id", "")
	dateFrom := c.Query("date_from", "")
	dateTo := c.Query("date_to", "")

	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		employeeID = database.UserIDFromContext(c.Locals("user_id"))
	}

	resp, err := h.dailyJournalService.List(c.Context(), page, perPage, departmentID, employeeID, dateFrom, dateTo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Journals,
		"Berhasil memuat jurnal harian",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetJournal returns single journal detail
// GET /api/daily-journals/:id
func (h *DailyJournalHandler) GetJournal(c *fiber.Ctx) error {
	id := c.Params("id")

	j, err := h.dailyJournalService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		userID := database.UserIDFromContext(c.Locals("user_id"))
		if j.EmployeeID.String() != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Anda tidak memiliki akses"))
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(j, "Berhasil memuat jurnal harian"))
}

// CreateJournal creates a new daily journal entry
// POST /api/daily-journals
func (h *DailyJournalHandler) CreateJournal(c *fiber.Ctx) error {
	req := new(models.CreateDailyJournalRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	j, err := h.dailyJournalService.Create(c.Context(), userID, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "deskripsi pekerjaan harus diisi":
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(j, "Jurnal harian berhasil dibuat"))
}

// AcknowledgeJournal acknowledges a journal entry (manager confirms)
// PUT /api/daily-journals/:id/acknowledge
func (h *DailyJournalHandler) AcknowledgeJournal(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	j, err := h.dailyJournalService.Acknowledge(c.Context(), id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "jurnal tidak ditemukan atau sudah diakui" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(j, "Jurnal harian berhasil diakui"))
}
