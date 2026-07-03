package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type PositionService struct{}

func NewPositionService() *PositionService {
	return &PositionService{}
}

func (s *PositionService) List(ctx context.Context, page, perPage int, search string) (*models.PositionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	positions, total, err := repository.ListPositions(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data posisi jabatan: %w", err)
	}
	if positions == nil {
		positions = []models.PositionSummary{}
	}

	return &models.PositionListResponse{
		Positions: positions,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *PositionService) GetAll(ctx context.Context) ([]models.PositionSummary, error) {
	positions, err := repository.GetAllPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar posisi jabatan: %w", err)
	}
	if positions == nil {
		positions = []models.PositionSummary{}
	}
	return positions, nil
}

func (s *PositionService) Get(ctx context.Context, id string) (*models.Position, error) {
	pos, err := repository.GetPositionByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data posisi jabatan")
	}
	if pos == nil {
		return nil, errors.New("posisi jabatan tidak ditemukan")
	}
	return pos, nil
}

func (s *PositionService) Create(ctx context.Context, req *models.CreatePositionRequest, userID string) (*models.Position, error) {
	if req.Name == "" {
		return nil, errors.New("nama posisi jabatan harus diisi")
	}
	if req.DepartmentID == "" {
		return nil, errors.New("departemen harus dipilih")
	}

	exists, err := repository.CheckPositionNameExists(ctx, req.Name, req.DepartmentID, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi posisi jabatan: %w", err)
	}
	if exists {
		return nil, errors.New("nama posisi jabatan sudah digunakan di departemen ini")
	}

	pos, err := repository.CreatePosition(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat posisi jabatan: %w", err)
	}
	return pos, nil
}

func (s *PositionService) Update(ctx context.Context, id string, req *models.UpdatePositionRequest, userID string) (*models.Position, error) {
	existing, err := repository.GetPositionByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data posisi jabatan")
	}
	if existing == nil {
		return nil, errors.New("posisi jabatan tidak ditemukan")
	}

	if req.Name != nil && *req.Name != existing.Name {
		deptID := existing.DepartmentID.String()
		if req.DepartmentID != nil {
			deptID = *req.DepartmentID
		}
		exists, err := repository.CheckPositionNameExists(ctx, *req.Name, deptID, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi posisi jabatan: %w", err)
		}
		if exists {
			return nil, errors.New("nama posisi jabatan sudah digunakan di departemen ini")
		}
	}

	pos, err := repository.UpdatePosition(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui posisi jabatan: %w", err)
	}
	return pos, nil
}

func (s *PositionService) Delete(ctx context.Context, id, userID string) error {
	existing, err := repository.GetPositionByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data posisi jabatan")
	}
	if existing == nil {
		return errors.New("posisi jabatan tidak ditemukan")
	}

	hasEmployees, err := repository.CheckPositionHasEmployees(ctx, id)
	if err != nil {
		return fmt.Errorf("gagal validasi posisi jabatan: %w", err)
	}
	if hasEmployees {
		return errors.New("posisi jabatan masih memiliki karyawan, tidak dapat dihapus")
	}

	err = repository.DeletePosition(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus posisi jabatan: %w", err)
	}
	return nil
}
