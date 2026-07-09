package handlers

import (
	"strings"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type EmployeeHandler struct {
	employeeService *service.EmployeeService
}

func NewEmployeeHandler(employeeService *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employeeService: employeeService}
}

// ListEmployees returns paginated employee list
// GET /api/employees
func (h *EmployeeHandler) ListEmployees(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)
	search := c.Query("search", "")
	departmentID := c.Query("department_id", "")
	status := c.Query("status", "")
	includeDeleted := c.QueryBool("include_deleted", false)

	resp, err := h.employeeService.ListEmployees(c.Context(), page, perPage, search, departmentID, status, includeDeleted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	// Sembunyikan gaji untuk karyawan biasa (role employee)
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		for i := range resp.Employees {
			resp.Employees[i].BaseSalary = 0
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Employees,
		"Berhasil memuat data karyawan",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// GetEmployee returns single employee detail
// GET /api/employees/:id
func (h *EmployeeHandler) GetEmployee(c *fiber.Ctx) error {
	id := c.Params("id")

	employee, err := h.employeeService.GetEmployee(c.Context(), id)
	if err != nil {
		if err.Error() == "karyawan tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	// Sembunyikan gaji untuk karyawan biasa (kecuali data dirinya sendiri)
	roleSlug, _ := c.Locals("role_slug").(string)
	if roleSlug == "employee" {
		requestingUserID, _ := c.Locals("user_id").(string)
		if requestingUserID != "" && employee.ID.String() != requestingUserID {
			employee.BaseSalary = nil
			employee.DailyWage = nil
		}
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(employee, "Berhasil memuat detail karyawan"))
}

// Dashboard returns executive dashboard stats
// GET /api/dashboard
func (h *EmployeeHandler) Dashboard(c *fiber.Ctx) error {
	stats, err := h.employeeService.GetDashboard(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat dashboard"))
}

// ManagerDashboard returns manager-specific dashboard stats
// GET /api/dashboard/manager
func (h *EmployeeHandler) ManagerDashboard(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Tidak dapat mengidentifikasi pengguna"))
	}

	stats, err := h.employeeService.GetManagerDashboard(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat dashboard manajer"))
}

// HRDashboard returns HR-specific dashboard stats
// GET /api/dashboard/hr
func (h *EmployeeHandler) HRDashboard(c *fiber.Ctx) error {
	stats, err := h.employeeService.GetHRDashboard(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(stats, "Berhasil memuat dashboard HR"))
}

// CreateEmployee creates a new employee
// POST /api/employees
func (h *EmployeeHandler) CreateEmployee(c *fiber.Ctx) error {
	var req models.CreateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	employee, err := h.employeeService.CreateEmployee(c.Context(), &req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "harus diisi"):
			code = fiber.StatusBadRequest
		case contains(msg, "sudah digunakan"):
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	// Log creation
	h.employeeService.LogHistory(c.Context(), employee.ID.String(), "created",
		nil,
		map[string]any{
			"full_name":         employee.FullName,
			"employment_status": employee.EmploymentStatus,
		},
		"Karyawan baru", userID)

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(employee, "Karyawan berhasil ditambahkan"))
}

// UpdateEmployee updates an existing employee
// PUT /api/employees/:id
func (h *EmployeeHandler) UpdateEmployee(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	// Fetch old data for diffing
	oldEmp, _ := h.employeeService.GetEmployee(c.Context(), id)

	employee, err := h.employeeService.UpdateEmployee(c.Context(), id, &req, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		switch {
		case contains(msg, "tidak ditemukan"):
			code = fiber.StatusNotFound
		case contains(msg, "sudah digunakan"):
			code = fiber.StatusConflict
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	// Log changes (safe — error tidak digagalkan)
	if oldEmp != nil {
		func() {
			oldV := map[string]any{}
			newV := map[string]any{}

			if oldEmp.FullName != employee.FullName {
				oldV["full_name"] = oldEmp.FullName
				newV["full_name"] = employee.FullName
			}
			if oldEmp.PositionName != employee.PositionName {
				oldV["position_name"] = oldEmp.PositionName
				newV["position_name"] = employee.PositionName
			}
			if oldEmp.DepartmentName != employee.DepartmentName {
				oldV["department_name"] = oldEmp.DepartmentName
				newV["department_name"] = employee.DepartmentName
			}
			if oldEmp.EmploymentStatus != employee.EmploymentStatus {
				oldV["employment_status"] = oldEmp.EmploymentStatus
				newV["employment_status"] = employee.EmploymentStatus
			}
			if oldEmp.RoleName != employee.RoleName {
				oldV["role_name"] = oldEmp.RoleName
				newV["role_name"] = employee.RoleName
			}
			if oldEmp.IsActive != employee.IsActive {
				oldV["is_active"] = oldEmp.IsActive
				newV["is_active"] = employee.IsActive
			}

			if len(oldV) == 0 {
				return
			}

			changeType := "updated"
			if _, ok := newV["position_name"]; ok {
				changeType = "promotion"
			}
			if _, ok := newV["department_name"]; ok {
				changeType = "mutation"
			}
			if _, ok := newV["is_active"]; ok && !employee.IsActive {
				changeType = "deactivated"
			}

			h.employeeService.LogHistory(c.Context(), employee.ID.String(), changeType, oldV, newV, "", userID)
		}()
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(employee, "Karyawan berhasil diperbarui"))
}

// UpdateWorkSchedule sets individual work schedule override for an employee
// PUT /api/employees/:id/work-schedule
func (h *EmployeeHandler) UpdateWorkSchedule(c *fiber.Ctx) error {
	id := c.Params("id")

	req := new(models.EmployeeWorkScheduleOverride)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	employee, err := h.employeeService.UpdateWorkSchedule(c.Context(), id, req.WorkScheduleID, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		switch {
		case contains(err.Error(), "tidak ditemukan"):
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(employee, "Jadwal kerja berhasil diupdate"))
}

// RestoreEmployee restores a soft-deleted employee
// PUT /api/employees/:id/restore
func (h *EmployeeHandler) RestoreEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.employeeService.RestoreEmployee(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		if contains(msg, "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Karyawan berhasil diaktifkan kembali"))
}

// DeleteEmployee soft-deletes an employee
// DELETE /api/employees/:id
func (h *EmployeeHandler) DeleteEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.employeeService.DeleteEmployee(c.Context(), id, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		if contains(msg, "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Karyawan berhasil dinonaktifkan"))
}

// GetEmployeeHistory returns paginated employee history
// GET /api/employees/:id/history
func (h *EmployeeHandler) GetEmployeeHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 25)

	resp, err := h.employeeService.GetEmployeeHistory(c.Context(), id, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponseWithMeta(
		resp.Histories,
		"Berhasil memuat riwayat karyawan",
		PaginationMeta(resp.Total, resp.Page, resp.PerPage),
	))
}

// ExportEmployees exports all employees to Excel
// GET /api/employees/export
func (h *EmployeeHandler) ExportEmployees(c *fiber.Ctx) error {
	fileBytes, err := h.employeeService.ExportEmployees(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	c.Response().Header.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header.Set("Content-Disposition", "attachment; filename=employees.xlsx")
	return c.Send(fileBytes)
}

// ImportEmployees imports employees from uploaded Excel file
// POST /api/employees/import
func (h *EmployeeHandler) ImportEmployees(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("File tidak ditemukan"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	result, err := h.employeeService.ImportEmployees(c.Context(), file, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(result, "Import selesai"))
}

// RegisterFaceDescriptor stores face descriptor for an employee
// POST /api/employees/:id/face-descriptor
func (h *EmployeeHandler) RegisterFaceDescriptor(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.FaceDescriptorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format data tidak valid"))
	}

	if req.Descriptor == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Deskriptor wajah harus diisi"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	err := h.employeeService.RegisterFaceDescriptor(c.Context(), id, req.Descriptor, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		if contains(msg, "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Deskriptor wajah berhasil disimpan"))
}

// UploadPhoto uploads employee photo
// POST /api/employees/:id/photo
func (h *EmployeeHandler) UploadPhoto(c *fiber.Ctx) error {
	id := c.Params("id")

	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("File foto tidak ditemukan"))
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Hanya file JPEG dan PNG yang diizinkan"))
	}

	// Max 2MB
	if file.Size > 2*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Ukuran file maksimal 2MB"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))

	photoURL, err := h.employeeService.UploadPhoto(c.Context(), id, file, userID)
	if err != nil {
		code := fiber.StatusInternalServerError
		msg := err.Error()
		if contains(msg, "tidak ditemukan") {
			code = fiber.StatusNotFound
		}
		return c.Status(code).JSON(ErrorResponse(msg))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{"photo_url": photoURL}, "Foto berhasil diupload"))
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
