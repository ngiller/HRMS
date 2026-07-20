package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"hrms-backend/internal/config"
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/google/uuid"
)

type ReimbursementService struct{}

func NewReimbursementService() *ReimbursementService {
	return &ReimbursementService{}
}

func (s *ReimbursementService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.ReimbursementListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	reimbursements, total, err := repository.ListReimbursements(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data reimbursement: %w", err)
	}
	if reimbursements == nil {
		reimbursements = []models.ReimbursementSummary{}
	}

	return &models.ReimbursementListResponse{
		Reimbursements: reimbursements,
		Total:          total,
		Page:           page,
		PerPage:        perPage,
	}, nil
}

func (s *ReimbursementService) Get(ctx context.Context, id string) (*models.Reimbursement, error) {
	r, err := repository.GetReimbursementByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data reimbursement")
	}
	if r == nil {
		return nil, errors.New("reimbursement tidak ditemukan")
	}
	return r, nil
}

func (s *ReimbursementService) Create(ctx context.Context, employeeID string, req *models.CreateReimbursementReq) (*models.Reimbursement, error) {
	if req.Type == "" {
		return nil, errors.New("tipe reimbursement harus diisi")
	}
	if req.Amount <= 0 {
		return nil, errors.New("jumlah reimbursement harus lebih dari 0")
	}
	if req.Description == "" {
		return nil, errors.New("deskripsi reimbursement harus diisi")
	}

	r, err := repository.CreateReimbursement(ctx, employeeID, employeeID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengajuan reimbursement: %w", err)
	}

	// Initialize workflow tracking (non-blocking, ignore errors)
	err = s.initWorkflowTracking(ctx, "reimbursement", r.ID.String(), employeeID)
	if err != nil {
		fmt.Printf("[WARN] Reimbursement workflow init: %v\n", err)
	}

	return r, nil
}

func (s *ReimbursementService) initWorkflowTracking(ctx context.Context, entityType, entityID, employeeID string) error {
	wfSvc := NewApprovalWorkflowService()
	_, err := wfSvc.ResolveWorkflowForRequest(ctx, entityType, entityID, employeeID)
	return err
}

func (s *ReimbursementService) Approve(ctx context.Context, id, approverID string) (*models.Reimbursement, error) {
	existing, err := repository.GetReimbursementByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data reimbursement")
	}
	if existing == nil {
		return nil, errors.New("reimbursement tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("reimbursement sudah diproses")
	}

	r, err := repository.UpdateReimbursementStatus(ctx, id, "approved", "", approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menyetujui reimbursement: %w", err)
	}
	return r, nil
}

func (s *ReimbursementService) Reject(ctx context.Context, id, rejectionReason, approverID string) (*models.Reimbursement, error) {
	existing, err := repository.GetReimbursementByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data reimbursement")
	}
	if existing == nil {
		return nil, errors.New("reimbursement tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("reimbursement sudah diproses")
	}

	r, err := repository.UpdateReimbursementStatus(ctx, id, "rejected", rejectionReason, approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menolak reimbursement: %w", err)
	}
	return r, nil
}

func (s *ReimbursementService) Pay(ctx context.Context, id, paidBy, paymentMethod string) (*models.Reimbursement, error) {
	if paymentMethod == "" {
		paymentMethod = "payroll"
	}

	existing, err := repository.GetReimbursementByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data reimbursement")
	}
	if existing == nil {
		return nil, errors.New("reimbursement tidak ditemukan")
	}
	if existing.Status != "approved" {
		return nil, errors.New("reimbursement harus disetujui terlebih dahulu sebelum dibayar")
	}

	r, err := repository.PayReimbursement(ctx, id, paidBy, paymentMethod)
	if err != nil {
		return nil, fmt.Errorf("gagal membayar reimbursement: %w", err)
	}
	return r, nil
}

func (s *ReimbursementService) UploadReceipt(ctx context.Context, file *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := "receipt_" + uuid.New().String() + ext
	uploadDir := config.Load().UploadDir

	// Create receipts subdirectory
	receiptDir := filepath.Join(uploadDir, "receipts")
	if err := os.MkdirAll(receiptDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	savePath := filepath.Join(receiptDir, filename)
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

	// Return URL path (served via Static /uploads)
	return "/uploads/receipts/" + filename, nil
}

func (s *ReimbursementService) Cancel(ctx context.Context, id, employeeID string) error {
	existing, err := repository.GetReimbursementByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data reimbursement")
	}
	if existing == nil {
		return errors.New("reimbursement tidak ditemukan")
	}
	if existing.Status != "pending" {
		return errors.New("reimbursement sudah diproses, tidak dapat dibatalkan")
	}

	err = repository.CancelReimbursement(ctx, id, employeeID)
	if err != nil {
		return fmt.Errorf("gagal membatalkan reimbursement: %w", err)
	}

	// Cancel workflow tracking (non-blocking, ignore errors)
	wfSvc := NewApprovalWorkflowService()
	_ = wfSvc.CancelWorkflowTracking(ctx, "reimbursement", id)

	return nil
}
