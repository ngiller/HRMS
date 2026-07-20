package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type ShiftService struct{}

func NewShiftService() *ShiftService {
	return &ShiftService{}
}

func (s *ShiftService) List(ctx context.Context, page, perPage int, search string, departmentID string) (*models.ShiftListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	shifts, total, err := repository.ListShifts(ctx, page, perPage, search, departmentID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data shift: %w", err)
	}
	if shifts == nil {
		shifts = []models.ShiftSummary{}
	}

	return &models.ShiftListResponse{
		Shifts:  shifts,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *ShiftService) GetAll(ctx context.Context, departmentID string) ([]models.ShiftSummary, error) {
	shifts, err := repository.GetAllShifts(ctx, departmentID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar shift: %w", err)
	}
	if shifts == nil {
		shifts = []models.ShiftSummary{}
	}
	return shifts, nil
}

func (s *ShiftService) Get(ctx context.Context, id string) (*models.Shift, error) {
	shift, err := repository.GetShiftByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data shift")
	}
	if shift == nil {
		return nil, errors.New("shift tidak ditemukan")
	}
	return shift, nil
}

func (s *ShiftService) Create(ctx context.Context, req *models.CreateShiftRequest, userID string) (*models.Shift, error) {
	if req.Name == "" {
		return nil, errors.New("nama shift harus diisi")
	}
	if req.Code == "" {
		return nil, errors.New("kode shift harus diisi")
	}
	if req.StartTime == "" {
		return nil, errors.New("jam mulai harus diisi")
	}
	if req.EndTime == "" {
		return nil, errors.New("jam selesai harus diisi")
	}
	if req.Color == "" {
		req.Color = "#3B82F6"
	}

	exists, err := repository.CheckShiftCodeExists(ctx, req.Code, req.DepartmentID, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi shift: %w", err)
	}
	if exists {
		return nil, errors.New("kode shift sudah digunakan")
	}

	exists, err = repository.CheckShiftNameExists(ctx, req.Name, req.DepartmentID, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi shift: %w", err)
	}
	if exists {
		return nil, errors.New("nama shift sudah digunakan")
	}

	shift, err := repository.CreateShift(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat shift: %w", err)
	}
	return shift, nil
}

func (s *ShiftService) Update(ctx context.Context, id string, req *models.UpdateShiftRequest, userID string) (*models.Shift, error) {
	existing, err := repository.GetShiftByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data shift")
	}
	if existing == nil {
		return nil, errors.New("shift tidak ditemukan")
	}

	deptID := ""
	if existing.DepartmentID != nil {
		deptID = existing.DepartmentID.String()
	}
	if req.DepartmentID != nil {
		deptID = *req.DepartmentID
	}

	if (req.Code != nil && *req.Code != existing.Code) || req.DepartmentID != nil {
		codeToCheck := existing.Code
		if req.Code != nil {
			codeToCheck = *req.Code
		}
		exists, err := repository.CheckShiftCodeExists(ctx, codeToCheck, deptID, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi shift: %w", err)
		}
		if exists {
			return nil, errors.New("kode shift sudah digunakan")
		}
	}

	if (req.Name != nil && *req.Name != existing.Name) || req.DepartmentID != nil {
		nameToCheck := existing.Name
		if req.Name != nil {
			nameToCheck = *req.Name
		}
		exists, err := repository.CheckShiftNameExists(ctx, nameToCheck, deptID, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi shift: %w", err)
		}
		if exists {
			return nil, errors.New("nama shift sudah digunakan")
		}
	}

	shift, err := repository.UpdateShift(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui shift: %w", err)
	}
	return shift, nil
}

func (s *ShiftService) Delete(ctx context.Context, id, userID string) error {
	existing, err := repository.GetShiftByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data shift")
	}
	if existing == nil {
		return errors.New("shift tidak ditemukan")
	}

	err = repository.DeleteShift(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus shift: %w", err)
	}
	return nil
}
