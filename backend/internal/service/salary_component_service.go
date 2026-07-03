package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type SalaryComponentService struct{}

func NewSalaryComponentService() *SalaryComponentService {
	return &SalaryComponentService{}
}

func (s *SalaryComponentService) ListComponents(ctx context.Context, employeeID string, page, perPage int) (*models.SalaryComponentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	components, total, err := repository.ListSalaryComponents(ctx, employeeID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat komponen gaji: %w", err)
	}

	if components == nil {
		components = []models.SalaryComponentSummary{}
	}

	return &models.SalaryComponentListResponse{
		Components: components,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

func (s *SalaryComponentService) GetComponent(ctx context.Context, componentID string) (*models.SalaryComponent, error) {
	comp, err := repository.GetSalaryComponent(ctx, componentID)
	if err != nil {
		return nil, errors.New("gagal memuat komponen gaji")
	}
	if comp == nil {
		return nil, errors.New("komponen gaji tidak ditemukan")
	}
	return comp, nil
}

func (s *SalaryComponentService) CreateComponent(ctx context.Context, employeeID string, req *models.CreateSalaryComponentRequest, userID string) (*models.SalaryComponent, error) {
	// Validasi
	if req.ComponentName == "" {
		return nil, errors.New("nama komponen harus diisi")
	}
	if req.ComponentType != "allowance" && req.ComponentType != "deduction" {
		return nil, errors.New("tipe komponen harus allowance atau deduction")
	}
	if req.Amount < 0 {
		return nil, errors.New("jumlah tidak boleh negatif")
	}

	comp, err := repository.CreateSalaryComponent(ctx, employeeID, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal menambah komponen gaji: %w", err)
	}
	return comp, nil
}

func (s *SalaryComponentService) UpdateComponent(ctx context.Context, componentID string, req *models.UpdateSalaryComponentRequest, userID string) (*models.SalaryComponent, error) {
	// Cek apakah komponen ada
	existing, err := repository.GetSalaryComponent(ctx, componentID)
	if err != nil {
		return nil, errors.New("gagal memvalidasi komponen gaji")
	}
	if existing == nil {
		return nil, errors.New("komponen gaji tidak ditemukan")
	}

	// Validasi
	if req.ComponentType != nil && *req.ComponentType != "allowance" && *req.ComponentType != "deduction" {
		return nil, errors.New("tipe komponen harus allowance atau deduction")
	}
	if req.Amount != nil && *req.Amount < 0 {
		return nil, errors.New("jumlah tidak boleh negatif")
	}

	comp, err := repository.UpdateSalaryComponent(ctx, componentID, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate komponen gaji: %w", err)
	}
	return comp, nil
}

func (s *SalaryComponentService) DeleteComponent(ctx context.Context, componentID, userID string) error {
	existing, err := repository.GetSalaryComponent(ctx, componentID)
	if err != nil {
		return errors.New("gagal memvalidasi komponen gaji")
	}
	if existing == nil {
		return errors.New("komponen gaji tidak ditemukan")
	}

	return repository.DeleteSalaryComponent(ctx, componentID, userID)
}
