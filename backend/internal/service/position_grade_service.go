package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type PositionGradeService struct{}

func NewPositionGradeService() *PositionGradeService {
	return &PositionGradeService{}
}

func (s *PositionGradeService) List(ctx context.Context, page, perPage int, search string) (*models.PositionGradeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	grades, total, err := repository.ListPositionGrades(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data golongan jabatan: %w", err)
	}
	if grades == nil {
		grades = []models.PositionGradeSummary{}
	}

	return &models.PositionGradeListResponse{
		PositionGrades: grades,
		Total:          total,
		Page:           page,
		PerPage:        perPage,
	}, nil
}

func (s *PositionGradeService) GetAll(ctx context.Context) ([]models.PositionGradeSummary, error) {
	grades, err := repository.GetAllPositionGrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar golongan jabatan: %w", err)
	}
	if grades == nil {
		grades = []models.PositionGradeSummary{}
	}
	return grades, nil
}

func (s *PositionGradeService) Get(ctx context.Context, id string) (*models.PositionGrade, error) {
	g, err := repository.GetPositionGradeByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data golongan jabatan")
	}
	if g == nil {
		return nil, errors.New("golongan jabatan tidak ditemukan")
	}
	return g, nil
}

func (s *PositionGradeService) Create(ctx context.Context, req *models.CreatePositionGradeRequest, userID string) (*models.PositionGrade, error) {
	if req.Name == "" {
		return nil, errors.New("nama golongan jabatan harus diisi")
	}
	if req.Level < 1 {
		return nil, errors.New("level harus diisi (minimal 1)")
	}

	exists, err := repository.CheckPositionGradeNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi golongan jabatan: %w", err)
	}
	if exists {
		return nil, errors.New("nama golongan jabatan sudah digunakan")
	}

	levelExists, err := repository.CheckPositionGradeLevelExists(ctx, req.Level, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi golongan jabatan: %w", err)
	}
	if levelExists {
		return nil, errors.New("level sudah digunakan")
	}

	g, err := repository.CreatePositionGrade(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat golongan jabatan: %w", err)
	}
	return g, nil
}

func (s *PositionGradeService) Update(ctx context.Context, id string, req *models.UpdatePositionGradeRequest, userID string) (*models.PositionGrade, error) {
	existing, err := repository.GetPositionGradeByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data golongan jabatan")
	}
	if existing == nil {
		return nil, errors.New("golongan jabatan tidak ditemukan")
	}

	if req.Name != nil && *req.Name != existing.Name {
		exists, err := repository.CheckPositionGradeNameExists(ctx, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi golongan jabatan: %w", err)
		}
		if exists {
			return nil, errors.New("nama golongan jabatan sudah digunakan")
		}
	}

	if req.Level != nil && *req.Level != existing.Level {
		levelExists, err := repository.CheckPositionGradeLevelExists(ctx, *req.Level, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi golongan jabatan: %w", err)
		}
		if levelExists {
			return nil, errors.New("level sudah digunakan")
		}
	}

	g, err := repository.UpdatePositionGrade(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui golongan jabatan: %w", err)
	}
	return g, nil
}

func (s *PositionGradeService) Delete(ctx context.Context, id, userID string) error {
	existing, err := repository.GetPositionGradeByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data golongan jabatan")
	}
	if existing == nil {
		return errors.New("golongan jabatan tidak ditemukan")
	}

	hasPositions, err := repository.CheckPositionGradeHasPositions(ctx, id)
	if err != nil {
		return fmt.Errorf("gagal validasi golongan jabatan: %w", err)
	}
	if hasPositions {
		return errors.New("golongan jabatan masih digunakan oleh posisi, tidak dapat dihapus")
	}

	err = repository.DeletePositionGrade(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus golongan jabatan: %w", err)
	}
	return nil
}
