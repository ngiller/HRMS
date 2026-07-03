package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type DepartmentService struct{}

func NewDepartmentService() *DepartmentService {
	return &DepartmentService{}
}

func (s *DepartmentService) ListDepartments(ctx context.Context, page, perPage int, search string) (*models.DepartmentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	departments, total, err := repository.ListDepartments(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data departemen: %w", err)
	}

	if departments == nil {
		departments = []models.DepartmentSummary{}
	}

	return &models.DepartmentListResponse{
		Departments: departments,
		Total:       total,
		Page:        page,
		PerPage:     perPage,
	}, nil
}

func (s *DepartmentService) GetAllWorkSchedules(ctx context.Context) ([]models.WorkScheduleSummary, error) {
	schedules, err := repository.GetAllWorkSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data jadwal kerja: %w", err)
	}
	if schedules == nil {
		schedules = []models.WorkScheduleSummary{}
	}
	return schedules, nil
}

func (s *DepartmentService) GetAllDepartments(ctx context.Context) ([]models.DepartmentSummary, error) {
	departments, err := repository.GetAllDepartments(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar departemen: %w", err)
	}
	if departments == nil {
		departments = []models.DepartmentSummary{}
	}
	return departments, nil
}

func (s *DepartmentService) GetDepartment(ctx context.Context, id string) (*models.Department, error) {
	dept, err := repository.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data departemen")
	}
	if dept == nil {
		return nil, errors.New("departemen tidak ditemukan")
	}
	return dept, nil
}

func (s *DepartmentService) CreateDepartment(ctx context.Context, req *models.CreateDepartmentRequest, userID string) (*models.Department, error) {
	// Validasi
	if req.Name == "" {
		return nil, errors.New("nama departemen harus diisi")
	}
	if req.Code == "" {
		return nil, errors.New("kode departemen harus diisi")
	}

	// Cek duplikasi code
	exists, err := repository.CheckDepartmentCodeExists(ctx, req.Code, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi departemen: %w", err)
	}
	if exists {
		return nil, errors.New("kode departemen sudah digunakan")
	}

	// Cek duplikasi name
	nameExists, err := repository.CheckDepartmentNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi departemen: %w", err)
	}
	if nameExists {
		return nil, errors.New("nama departemen sudah digunakan")
	}

	dept, err := repository.CreateDepartment(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat departemen: %w", err)
	}
	return dept, nil
}

func (s *DepartmentService) UpdateDepartment(ctx context.Context, id string, req *models.UpdateDepartmentRequest, userID string) (*models.Department, error) {
	// Pastikan departemen ada
	existing, err := repository.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data departemen")
	}
	if existing == nil {
		return nil, errors.New("departemen tidak ditemukan")
	}

	// Validasi duplikasi code jika diubah
	if req.Code != nil && *req.Code != existing.Code {
		exists, err := repository.CheckDepartmentCodeExists(ctx, *req.Code, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi departemen: %w", err)
		}
		if exists {
			return nil, errors.New("kode departemen sudah digunakan")
		}
	}

	// Validasi duplikasi name jika diubah
	if req.Name != nil && *req.Name != existing.Name {
		nameExists, err := repository.CheckDepartmentNameExists(ctx, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi departemen: %w", err)
		}
		if nameExists {
			return nil, errors.New("nama departemen sudah digunakan")
		}
	}

	dept, err := repository.UpdateDepartment(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui departemen: %w", err)
	}
	return dept, nil
}

func (s *DepartmentService) UpdateWorkSchedule(ctx context.Context, id, workScheduleID, userID string) (*models.Department, error) {
	// Pastikan departemen ada
	existing, err := repository.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data departemen")
	}
	if existing == nil {
		return nil, errors.New("departemen tidak ditemukan")
	}

	// Jika work_schedule_id tidak kosong, validasi jadwal kerja ada
	if workScheduleID != "" {
		ws, err := repository.GetWorkScheduleByID(ctx, workScheduleID)
		if err != nil || ws == nil {
			return nil, errors.New("jadwal kerja tidak ditemukan")
		}
	}

	dept, err := repository.UpdateDepartmentWorkSchedule(ctx, id, workScheduleID, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate jadwal kerja: %w", err)
	}
	return dept, nil
}

func (s *DepartmentService) DeleteDepartment(ctx context.Context, id, userID string) error {
	// Pastikan departemen ada
	existing, err := repository.GetDepartmentByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data departemen")
	}
	if existing == nil {
		return errors.New("departemen tidak ditemukan")
	}

	// Cek apakah punya sub-departemen
	hasChildren, err := repository.CheckDepartmentHasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("gagal validasi departemen: %w", err)
	}
	if hasChildren {
		return errors.New("departemen memiliki sub-departemen, tidak dapat dihapus")
	}

	// Cek apakah punya karyawan
	hasEmployees, err := repository.CheckDepartmentHasEmployees(ctx, id)
	if err != nil {
		return fmt.Errorf("gagal validasi departemen: %w", err)
	}
	if hasEmployees {
		return errors.New("departemen memiliki karyawan, tidak dapat dihapus")
	}

	err = repository.DeleteDepartment(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus departemen: %w", err)
	}
	return nil
}
