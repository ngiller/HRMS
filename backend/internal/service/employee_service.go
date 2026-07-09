package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hrms-backend/internal/config"
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type EmployeeService struct{}

func NewEmployeeService() *EmployeeService {
	return &EmployeeService{}
}

func (s *EmployeeService) ListEmployees(ctx context.Context, page, perPage int, search, departmentID, status string, includeDeleted bool) (*models.EmployeeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	employees, total, err := repository.ListEmployees(ctx, page, perPage, search, departmentID, status, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data karyawan: %w", err)
	}

	if employees == nil {
		employees = []models.EmployeeSummary{}
	}

	return &models.EmployeeListResponse{
		Employees: employees,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *EmployeeService) GetEmployee(ctx context.Context, id string) (*models.Employee, error) {
	employee, err := repository.GetEmployeeByIDRepo(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data karyawan")
	}
	if employee == nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	return employee, nil
}

func (s *EmployeeService) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	stats, err := repository.GetDashboardStats(ctx)
	if err != nil {
		return nil, errors.New("gagal memuat data dashboard")
	}
	return stats, nil
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, req *models.CreateEmployeeRequest, userID string) (*models.Employee, error) {
	// Validasi
	if req.FullName == "" {
		return nil, errors.New("nama lengkap harus diisi")
	}
	if req.Email == "" {
		return nil, errors.New("email harus diisi")
	}
	if req.Password == "" {
		return nil, errors.New("password harus diisi")
	}
	if req.Gender == "" {
		return nil, errors.New("jenis kelamin harus diisi")
	}
	if req.EmploymentStatus == "" {
		return nil, errors.New("status karyawan harus diisi")
	}
	if req.JoinDate == "" {
		return nil, errors.New("tanggal bergabung harus diisi")
	}

	// Cek duplikasi email
	exists, err := repository.CheckEmailExists(ctx, req.Email, "")
	if err != nil {
		return nil, errors.New("gagal memvalidasi email")
	}
	if exists {
		return nil, errors.New("email sudah digunakan")
	}

	// Cek duplikasi employee_id
	if req.EmployeeID != "" {
		exists, err = repository.CheckEmployeeIDExists(ctx, req.EmployeeID, "")
		if err != nil {
			return nil, errors.New("gagal memvalidasi NIK")
		}
		if exists {
			return nil, errors.New("NIK karyawan sudah digunakan")
		}
	}

	employee, err := repository.CreateEmployee(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat karyawan: %w", err)
	}
	return employee, nil
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, id string, req *models.UpdateEmployeeRequest, userID string) (*models.Employee, error) {
	// Cek apakah karyawan ada
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// Cek duplikasi email (jika diubah)
	if req.Email != "" {
		exists, err = repository.CheckEmailExists(ctx, req.Email, id)
		if err != nil {
			return nil, errors.New("gagal memvalidasi email")
		}
		if exists {
			return nil, errors.New("email sudah digunakan")
		}
	}

	employee, err := repository.UpdateEmployee(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate karyawan: %w", err)
	}
	return employee, nil
}

func (s *EmployeeService) GetEmployeeHistory(ctx context.Context, id string, page, perPage int) (*models.EmployeeHistoryListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	histories, total, err := repository.GetEmployeeHistory(ctx, id, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat riwayat karyawan: %w", err)
	}

	if histories == nil {
		histories = []models.EmployeeHistory{}
	}

	return &models.EmployeeHistoryListResponse{
		Histories: histories,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *EmployeeService) LogHistory(ctx context.Context, employeeID, changeType string, oldValue, newValue map[string]any, reason, changedBy string) error {
	return repository.CreateEmployeeHistory(ctx, employeeID, changeType, oldValue, newValue, reason, changedBy)
}

func (s *EmployeeService) ExportEmployees(ctx context.Context) ([]byte, error) {
	employees, err := repository.ListEmployeesForExport(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data karyawan: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Karyawan"
	f.SetSheetName("Sheet1", sheet)

	// Headers
	headers := []string{"NIK", "Nama Lengkap", "Email", "Jenis Kelamin", "Status", "Role", "Posisi", "Departemen", "Tanggal Bergabung", "No. Telepon", "Status Aktif"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Bold headers
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)

	for i, emp := range employees {
		row := i + 2
		vals := []interface{}{
			emp.EmployeeID, emp.FullName, emp.Email,
			emp.Gender, emp.EmploymentStatus,
			emp.RoleName, emp.PositionName, emp.DepartmentName,
			emp.JoinDate, emp.Phone,
			map[bool]string{true: "Aktif", false: "Non-Aktif"}[emp.IsActive],
		}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	// Auto-width columns
	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 20)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("gagal menulis file Excel: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *EmployeeService) ImportEmployees(ctx context.Context, file *multipart.FileHeader, userID string) (*models.ImportResult, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file Excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca sheet: %w", err)
	}

	result := &models.ImportResult{}
	if len(rows) < 2 {
		result.Message = "Tidak ada data untuk diimpor"
		return result, nil
	}

	// Column mapping (0-indexed from Excel):
	// A(0)=NIK, B(1)=Nama*, C(2)=Email*, D(3)=Gender, E(4)=Status,
	// F(5)=TempatLahir, G(6)=TglLahir, H(7)=Agama, I(8)=StatusPernikahan,
	// J(9)=TglBergabung, K(10)=NoTelepon, L(11)=Alamat
	// Use GetCellValue per cell to preserve empty leading cells (e.g. empty NIK)
	for i := range rows[1:] {
		rowNum := i + 2

		cells := make([]string, 12)
		for c := 0; c < 12; c++ {
			colName, _ := excelize.ColumnNumberToName(c + 1)
			val, _ := f.GetCellValue(sheet, fmt.Sprintf("%s%d", colName, rowNum))
			cells[c] = strings.TrimSpace(val)
		}

		fullName := cells[1]
		email := cells[2]
		if fullName == "" || email == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: nama dan email wajib diisi", rowNum))
			continue
		}

		gender := "laki_laki"
		g := strings.ToLower(cells[3])
		if g == "perempuan" || g == "p" {
			gender = "perempuan"
		}

		empStatus := "percobaan"
		s := strings.ToLower(cells[4])
		if s == "tetap" || s == "kontrak" || s == "magang" {
			empStatus = s
		}

		maritalStatus := strings.ToLower(cells[8])

		joinDate := cells[9]
		if joinDate == "" {
			joinDate = "2026-01-01"
		}

		employeeID := strings.TrimSpace(cells[0])
		if employeeID == "" {
			employeeID = fmt.Sprintf("IMP-%d", time.Now().UnixNano())
		}

		req := &models.CreateEmployeeRequest{
			EmployeeID:       employeeID,
			FullName:         fullName,
			Email:            email,
			Password:         "password123",
			Gender:           gender,
			EmploymentStatus: empStatus,
			JoinDate:         joinDate,
			Phone:            cells[10],
			PlaceOfBirth:     cells[5],
			DateOfBirth:      cells[6],
			Religion:         strings.ToLower(cells[7]),
			MaritalStatus:    maritalStatus,
			Address:          cells[11],
		}

		_, err := repository.CreateEmployee(ctx, req, userID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: %v", rowNum, err))
			continue
		}
		result.Success++
	}

	result.Message = fmt.Sprintf("Berhasil mengimpor %d dari %d data", result.Success, len(rows[1:]))
	return result, nil
}

func (s *EmployeeService) UpdateWorkSchedule(ctx context.Context, id, workScheduleID, userID string) (*models.Employee, error) {
	// Validate employee exists
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// Validate work schedule exists (if non-empty)
	if workScheduleID != "" {
		ws, err := repository.GetWorkScheduleByID(ctx, workScheduleID)
		if err != nil || ws == nil {
			return nil, errors.New("jadwal kerja tidak ditemukan")
		}
	}

	employee, err := repository.UpdateEmployeeWorkSchedule(ctx, id, workScheduleID, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate jadwal kerja: %w", err)
	}
	return employee, nil
}

// RegisterFaceDescriptor stores a face descriptor for an employee
func (s *EmployeeService) RegisterFaceDescriptor(ctx context.Context, id, descriptorJSON, userID string) error {
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return errors.New("karyawan tidak ditemukan")
	}

	return repository.UpdateFaceDescriptor(ctx, id, descriptorJSON, userID)
}

func (s *EmployeeService) UploadPhoto(ctx context.Context, id string, file *multipart.FileHeader, userID string) (string, error) {
	// Check employee exists
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return "", errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return "", errors.New("karyawan tidak ditemukan")
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	photoFilename := "employee_" + id + ext
	uploadDir := config.Load().UploadDir

	// Save file to disk
	savePath := filepath.Join(uploadDir, photoFilename)
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("gagal menulis file: %w", err)
	}

	photoURL := "/uploads/" + photoFilename

	// Update database
	if err := repository.UpdateEmployeePhoto(ctx, id, photoURL, userID); err != nil {
		return "", fmt.Errorf("gagal memperbarui foto: %w", err)
	}

	return photoURL, nil
}
func (s *EmployeeService) RestoreEmployee(ctx context.Context, id string, userID string) error {
	return repository.RestoreEmployee(ctx, id, userID)
}

func (s *EmployeeService) GetManagerDashboard(ctx context.Context, userID string) (*models.ManagerDashboardResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("ID pengguna tidak valid")
	}
	stats, err := repository.GetManagerDashboardStats(ctx, uid)
	if err != nil {
		return nil, errors.New("gagal memuat dashboard manajer")
	}
	return stats, nil
}

func (s *EmployeeService) GetHRDashboard(ctx context.Context) (*models.HRDashboardResponse, error) {
	stats, err := repository.GetHRDashboardStats(ctx)
	if err != nil {
		return nil, errors.New("gagal memuat dashboard HR")
	}
	return stats, nil
}

func (s *EmployeeService) DeleteEmployee(ctx context.Context, id string, userID string) error {
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return errors.New("karyawan tidak ditemukan")
	}

	return repository.DeleteEmployee(ctx, id, userID)
}
