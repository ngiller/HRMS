package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type KPIService struct{}

func NewKPIService() *KPIService {
	return &KPIService{}
}

func (s *KPIService) ListTemplates(ctx context.Context, page, perPage int, year int) (*models.KPITemplateListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	templates, total, err := repository.ListKPITemplates(ctx, page, perPage, year)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat template KPI: %w", err)
	}
	if templates == nil {
		templates = []models.KPITemplate{}
	}

	return &models.KPITemplateListResponse{
		Templates: templates,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *KPIService) GetTemplate(ctx context.Context, id string) (*models.KPITemplateDetail, error) {
	t, err := repository.GetKPITemplateByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat template KPI")
	}
	if t == nil {
		return nil, errors.New("template KPI tidak ditemukan")
	}
	return t, nil
}

func (s *KPIService) ListReviews(ctx context.Context, page, perPage int, status, employeeID string) (*models.KPIReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	reviews, total, err := repository.ListKPIReviews(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat review KPI: %w", err)
	}
	if reviews == nil {
		reviews = []models.KPIReviewSummary{}
	}

	return &models.KPIReviewListResponse{
		Reviews: reviews,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *KPIService) GetReview(ctx context.Context, id string) (*models.KPIReview, error) {
	r, err := repository.GetKPIReviewByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat review KPI")
	}
	if r == nil {
		return nil, errors.New("review KPI tidak ditemukan")
	}
	return r, nil
}

func (s *KPIService) CreateReview(ctx context.Context, req *models.CreateKPIReviewRequest, userID string) (*models.KPIReview, error) {
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus diisi")
	}
	if req.KPITemplateID == "" {
		return nil, errors.New("template KPI harus diisi")
	}
	if req.Period == "" {
		return nil, errors.New("periode harus diisi")
	}
	if req.Year < 2024 {
		return nil, errors.New("tahun harus >= 2024")
	}

	r, err := repository.CreateKPIReview(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat review KPI: %w", err)
	}
	return r, nil
}

func (s *KPIService) CreateTemplate(ctx context.Context, req *models.CreateKPITemplateRequest, createdBy string) (*models.KPITemplateDetail, error) {
	if req.Title == "" {
		return nil, errors.New("judul template harus diisi")
	}
	if req.PeriodType == "" {
		return nil, errors.New("tipe periode harus diisi")
	}
	if req.Year < 2024 {
		return nil, errors.New("tahun harus >= 2024")
	}
	if len(req.Indicators) == 0 {
		return nil, errors.New("minimal 1 indikator harus ditambahkan")
	}

	// Validate indicator weights sum to 100
	totalWeight := 0.0
	for _, ind := range req.Indicators {
		if ind.Name == "" {
			return nil, errors.New("nama indikator harus diisi")
		}
		if ind.Target <= 0 {
			return nil, errors.New("target indikator harus lebih dari 0")
		}
		if ind.Weight <= 0 {
			return nil, errors.New("bobot indikator harus lebih dari 0")
		}
		totalWeight += ind.Weight
	}
	if totalWeight < 99.5 || totalWeight > 100.5 {
		return nil, errors.New("total bobot indikator harus 100%")
	}

	t, err := repository.CreateKPITemplate(ctx, req, createdBy)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat template KPI: %w", err)
	}
	return t, nil
}

func (s *KPIService) UpdateTemplate(ctx context.Context, id string, req *models.UpdateKPITemplateRequest) (*models.KPITemplateDetail, error) {
	if req.Title == "" {
		return nil, errors.New("judul template harus diisi")
	}
	if req.PeriodType == "" {
		return nil, errors.New("tipe periode harus diisi")
	}
	if req.Year < 2024 {
		return nil, errors.New("tahun harus >= 2024")
	}
	if len(req.Indicators) == 0 {
		return nil, errors.New("minimal 1 indikator harus ditambahkan")
	}

	totalWeight := 0.0
	for _, ind := range req.Indicators {
		if ind.Name == "" {
			return nil, errors.New("nama indikator harus diisi")
		}
		if ind.Target <= 0 {
			return nil, errors.New("target indikator harus lebih dari 0")
		}
		if ind.Weight <= 0 {
			return nil, errors.New("bobot indikator harus lebih dari 0")
		}
		totalWeight += ind.Weight
	}
	if totalWeight < 99.5 || totalWeight > 100.5 {
		return nil, errors.New("total bobot indikator harus 100%")
	}

	t, err := repository.UpdateKPITemplate(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate template KPI: %w", err)
	}
	return t, nil
}

func (s *KPIService) DeleteTemplate(ctx context.Context, id, userID string) error {
	return repository.DeleteKPITemplate(ctx, id, userID)
}
