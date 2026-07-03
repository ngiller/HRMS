package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type ReprimandService struct{}

func NewReprimandService() *ReprimandService {
	return &ReprimandService{}
}

func (s *ReprimandService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.ReprimandListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	reprimands, total, err := repository.ListReprimands(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data surat peringatan: %w", err)
	}
	if reprimands == nil {
		reprimands = []models.ReprimandSummary{}
	}

	return &models.ReprimandListResponse{
		Reprimands: reprimands,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

func (s *ReprimandService) Get(ctx context.Context, id string) (*models.Reprimand, error) {
	r, err := repository.GetReprimandByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data surat peringatan")
	}
	if r == nil {
		return nil, errors.New("surat peringatan tidak ditemukan")
	}
	return r, nil
}

func (s *ReprimandService) Create(ctx context.Context, req *models.CreateReprimandRequest, issuerID string) (*models.Reprimand, error) {
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus diisi")
	}
	if req.ReprimandType == "" {
		return nil, errors.New("tipe surat peringatan harus diisi")
	}
	if req.Title == "" {
		return nil, errors.New("judul surat peringatan harus diisi")
	}
	if req.ReprimandType != "verbal" && req.ReprimandType != "sp1" && req.ReprimandType != "sp2" && req.ReprimandType != "sp3" {
		return nil, errors.New("tipe surat peringatan tidak valid (verbal/sp1/sp2/sp3)")
	}
	if req.EffectivePeriodMonths <= 0 {
		req.EffectivePeriodMonths = 6
	}

	r, err := repository.CreateReprimand(ctx, req, issuerID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat surat peringatan: %w", err)
	}
	return r, nil
}

func (s *ReprimandService) Acknowledge(ctx context.Context, id, employeeID, note string) (*models.Reprimand, error) {
	r, err := repository.AcknowledgeReprimand(ctx, id, employeeID, note)
	if err != nil {
		return nil, err
	}
	return r, nil
}
