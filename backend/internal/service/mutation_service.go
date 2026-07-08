package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

// MutationService handles business logic for employee mutations & promotions
type MutationService struct {
	repo *repository.MutationRepo
}

// NewMutationService creates a new MutationService
func NewMutationService() *MutationService {
	return &MutationService{
		repo: repository.NewMutationRepo(),
	}
}

// List returns paginated mutations
func (s *MutationService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.MutationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	mutations, total, err := s.repo.List(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data mutasi: %w", err)
	}

	return &models.MutationListResponse{
		Mutations: mutations,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

// Get returns a single mutation by ID
func (s *MutationService) Get(ctx context.Context, id string) (*models.EmployeeMutation, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("mutasi tidak ditemukan")
	}
	return m, nil
}

// Create creates a new mutation request
func (s *MutationService) Create(ctx context.Context, req *models.CreateMutationRequest, creatorID string) (*models.EmployeeMutation, error) {
	// Validasi
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus diisi")
	}
	if req.MutationType == "" {
		return nil, errors.New("tipe mutasi harus diisi")
	}
	validTypes := map[string]bool{
		"promotion": true, "demotion": true, "transfer": true,
		"position_change": true, "status_change": true, "salary_change": true,
	}
	if !validTypes[req.MutationType] {
		return nil, errors.New("tipe mutasi tidak valid (promotion/demotion/transfer/position_change/status_change/salary_change)")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan mutasi harus diisi")
	}
	if req.EffectiveDate == "" {
		return nil, errors.New("tanggal berlaku harus diisi")
	}

	// Get current employee data to fill old values
	deptID, posID, gradeID, empStatus, baseSalary, err := s.repo.GetEmployeeData(ctx, req.EmployeeID)
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// Validation: at least one new value should be different
	if req.NewDepartmentID == "" && req.NewPositionID == "" && req.NewPositionGradeID == "" &&
		req.NewEmploymentStatus == "" && req.NewBaseSalary == nil {
		return nil, errors.New("minimal satu perubahan harus diisi (departemen/jabatan/grade/status/gaji)")
	}

	// Create the mutation
	m, err := s.repo.Create(ctx, req, creatorID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat mutasi: %w", err)
	}

	// Save old values (before mutation) to the mutation record
	if updateErr := s.repo.UpdateOldValues(ctx, m.ID, deptID, posID, gradeID, empStatus, baseSalary); updateErr != nil {
		// Non-fatal — log but don't block creation
		_ = updateErr
	}

	// Log history
	_ = repository.CreateEmployeeHistory(ctx, req.EmployeeID, "mutation_"+req.MutationType,
		map[string]any{
			"department_id":     deptID,
			"position_id":       posID,
			"position_grade_id": gradeID,
			"employment_status": empStatus,
			"base_salary":       baseSalary,
		},
		map[string]any{
			"department_id":     req.NewDepartmentID,
			"position_id":       req.NewPositionID,
			"position_grade_id": req.NewPositionGradeID,
			"employment_status": req.NewEmploymentStatus,
			"base_salary":       req.NewBaseSalary,
		},
		req.Reason, creatorID,
	)

	return m, nil
}

// Approve approves a mutation and applies the changes to employee
func (s *MutationService) Approve(ctx context.Context, id, approverID string) (*models.EmployeeMutation, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("mutasi tidak ditemukan")
	}
	if m.Status != "pending" {
		return nil, errors.New("mutasi sudah diproses")
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, id, "approved", approverID, ""); err != nil {
		return nil, fmt.Errorf("gagal approve mutasi: %w", err)
	}

	// Apply changes to employee record
	if err := s.repo.ApplyMutation(ctx, id); err != nil {
		return nil, fmt.Errorf("gagal menerapkan mutasi: %w", err)
	}

	// Log history
	_ = repository.CreateEmployeeHistory(ctx, m.EmployeeID, "mutation_"+m.MutationType+"_approved",
		map[string]any{"mutation_id": m.ID},
		map[string]any{"status": "approved"},
		"Disetujui oleh "+approverID, approverID,
	)

	return s.repo.GetByID(ctx, id)
}

// Reject rejects a mutation
func (s *MutationService) Reject(ctx context.Context, id, approverID, reason string) (*models.EmployeeMutation, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("mutasi tidak ditemukan")
	}
	if m.Status != "pending" {
		return nil, errors.New("mutasi sudah diproses")
	}

	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("alasan penolakan harus diisi")
	}

	if err := s.repo.UpdateStatus(ctx, id, "rejected", approverID, reason); err != nil {
		return nil, fmt.Errorf("gagal reject mutasi: %w", err)
	}

	return s.repo.GetByID(ctx, id)
}

// ExportExcel exports mutations as Excel file
func (s *MutationService) ExportExcel(ctx context.Context, status, employeeID string) ([]byte, string, error) {
	mutations, err := s.repo.ListAll(ctx, status, employeeID)
	if err != nil {
		return nil, "", fmt.Errorf("gagal mengambil data mutasi: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Riwayat Mutasi"
	f.SetSheetName("Sheet1", sheet)

	typeLabels := map[string]string{
		"promotion": "Promosi", "demotion": "Demosi", "transfer": "Mutasi Departemen",
		"position_change": "Perubahan Jabatan", "status_change": "Perubahan Status", "salary_change": "Perubahan Gaji",
	}
	statusLabels := map[string]string{
		"pending": "Menunggu", "approved": "Disetujui", "rejected": "Ditolak", "cancelled": "Dibatalkan",
	}

	headers := []string{"No", "Nama Karyawan", "Jenis Mutasi", "Departemen Lama", "Departemen Baru",
		"Jabatan Lama", "Jabatan Baru", "Gaji Lama", "Gaji Baru", "Status Lama", "Status Baru",
		"Tanggal Berlaku", "Status", "Alasan", "Tanggal Pengajuan"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)

	for i, m := range mutations {
		row := i + 2
		oldSalary := ""
		if m.OldBaseSalary != nil {
			oldSalary = fmt.Sprintf("Rp %.0f", *m.OldBaseSalary)
		}
		newSalary := ""
		if m.NewBaseSalary != nil {
			newSalary = fmt.Sprintf("Rp %.0f", *m.NewBaseSalary)
		}

		vals := []interface{}{
			i + 1,
			m.EmployeeName,
			typeLabels[m.MutationType],
			m.OldDepartmentName, m.NewDepartmentName,
			m.OldPositionName, m.NewPositionName,
			oldSalary, newSalary,
			m.OldEmploymentStatus, m.NewEmploymentStatus,
			m.EffectiveDate,
			statusLabels[m.Status],
			m.Reason,
			m.CreatedAt.Format("2006-01-02"),
		}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 22)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("gagal menulis file Excel: %w", err)
	}

	filename := fmt.Sprintf("riwayat-mutasi-%s.xlsx", time.Now().Format("2006-01-02"))
	return buf.Bytes(), filename, nil
}

// Cancel cancels a pending mutation
func (s *MutationService) Cancel(ctx context.Context, id, userID string) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("mutasi tidak ditemukan")
	}
	if m.Status != "pending" {
		return errors.New("mutasi sudah diproses, tidak dapat dibatalkan")
	}

	return s.repo.UpdateStatus(ctx, id, "cancelled", userID, "")
}
