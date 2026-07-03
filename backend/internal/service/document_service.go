package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type DocumentService struct{}

func NewDocumentService() *DocumentService {
	return &DocumentService{}
}

func (s *DocumentService) List(ctx context.Context, page, perPage int, status, employeeID, docType string) (*models.DocumentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	documents, total, err := repository.ListDocuments(ctx, page, perPage, status, employeeID, docType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data dokumen: %w", err)
	}
	if documents == nil {
		documents = []models.DocumentSummary{}
	}

	return &models.DocumentListResponse{
		Documents: documents,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *DocumentService) Get(ctx context.Context, id string) (*models.EmployeeDocument, error) {
	d, err := repository.GetDocumentByID(ctx, id)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if d == nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	return d, nil
}

func (s *DocumentService) Create(ctx context.Context, req *models.CreateDocumentReq) (*models.EmployeeDocument, error) {
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus dipilih")
	}
	if req.DocType == "" {
		return nil, errors.New("tipe dokumen harus diisi")
	}
	if req.Title == "" {
		return nil, errors.New("judul dokumen harus diisi")
	}

	d, err := repository.CreateDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat dokumen: %w", err)
	}
	return d, nil
}

func (s *DocumentService) Verify(ctx context.Context, id, verifierID string) (*models.EmployeeDocument, error) {
	existing, err := repository.GetDocumentByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("dokumen sudah diproses")
	}

	d, err := repository.UpdateDocumentStatus(ctx, id, "verified", "", verifierID)
	if err != nil {
		return nil, fmt.Errorf("gagal memverifikasi dokumen: %w", err)
	}
	return d, nil
}

func (s *DocumentService) Reject(ctx context.Context, id, rejectionReason, verifierID string) (*models.EmployeeDocument, error) {
	existing, err := repository.GetDocumentByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("dokumen sudah diproses")
	}

	d, err := repository.UpdateDocumentStatus(ctx, id, "rejected", rejectionReason, verifierID)
	if err != nil {
		return nil, fmt.Errorf("gagal menolak dokumen: %w", err)
	}
	return d, nil
}

func (s *DocumentService) Delete(ctx context.Context, id, userID string) error {
	return repository.DeleteDocument(ctx, id, userID)
}
