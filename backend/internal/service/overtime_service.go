package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type OvertimeService struct{}

func NewOvertimeService() *OvertimeService {
	return &OvertimeService{}
}

func (s *OvertimeService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.OvertimeRequestListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	requests, total, err := repository.ListOvertimeRequests(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data lembur: %w", err)
	}
	if requests == nil {
		requests = []models.OvertimeRequestSummary{}
	}

	return &models.OvertimeRequestListResponse{
		OvertimeRequests: requests,
		Total:            total,
		Page:             page,
		PerPage:          perPage,
	}, nil
}

func (s *OvertimeService) Get(ctx context.Context, id string) (*models.OvertimeRequest, error) {
	r, err := repository.GetOvertimeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data lembur")
	}
	if r == nil {
		return nil, errors.New("pengajuan lembur tidak ditemukan")
	}

	// Try to get calculation
	calc, calcErr := repository.GetOvertimeCalculation(ctx, id)
	if calcErr == nil && calc != nil {
		r.HourlyRate = calc.HourlyRate
		r.OvertimePay = calc.OvertimePay
	}

	return r, nil
}

func (s *OvertimeService) Create(ctx context.Context, employeeID string, req *models.CreateOvertimeRequestReq) (*models.OvertimeRequest, error) {
	if req.Date == "" {
		return nil, errors.New("tanggal lembur harus diisi")
	}
	if req.StartTime == "" {
		return nil, errors.New("waktu mulai harus diisi")
	}
	if req.EndTime == "" {
		return nil, errors.New("waktu selesai harus diisi")
	}
	if req.TotalHours <= 0 {
		return nil, errors.New("total jam lembur harus lebih dari 0")
	}
	if req.OvertimeType == "" {
		return nil, errors.New("tipe lembur harus diisi")
	}
	if req.OvertimeType != "weekday" && req.OvertimeType != "weekend" && req.OvertimeType != "holiday" {
		return nil, errors.New("tipe lembur tidak valid (weekday/weekend/holiday)")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan lembur harus diisi")
	}

	r, err := repository.CreateOvertimeRequest(ctx, employeeID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengajuan lembur: %w", err)
	}

	// Initialize workflow tracking (non-blocking, ignore errors)
	err = s.initWorkflowTracking(ctx, "overtime", r.ID.String(), employeeID, r.TotalHours)
	if err != nil {
		fmt.Printf("[WARN] Overtime workflow init: %v\n", err)
	}

	return r, nil
}

func (s *OvertimeService) initWorkflowTracking(ctx context.Context, entityType, entityID, employeeID string, conditionValue float64) error {
	wfSvc := NewApprovalWorkflowService()
	_, err := wfSvc.ResolveWorkflowForRequest(ctx, entityType, entityID, employeeID, conditionValue)
	return err
}

func (s *OvertimeService) Approve(ctx context.Context, id, approverID string) (*models.OvertimeRequest, error) {
	existing, err := repository.GetOvertimeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data lembur")
	}
	if existing == nil {
		return nil, errors.New("pengajuan lembur tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("pengajuan lembur sudah diproses")
	}

	r, err := repository.UpdateOvertimeStatus(ctx, id, "approved", "", approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menyetujui pengajuan lembur: %w", err)
	}
	return r, nil
}

func (s *OvertimeService) Reject(ctx context.Context, id, rejectionReason, approverID string) (*models.OvertimeRequest, error) {
	existing, err := repository.GetOvertimeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data lembur")
	}
	if existing == nil {
		return nil, errors.New("pengajuan lembur tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("pengajuan lembur sudah diproses")
	}

	r, err := repository.UpdateOvertimeStatus(ctx, id, "rejected", rejectionReason, approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menolak pengajuan lembur: %w", err)
	}
	return r, nil
}

func (s *OvertimeService) Cancel(ctx context.Context, id, employeeID string) error {
	existing, err := repository.GetOvertimeRequestByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data lembur")
	}
	if existing == nil {
		return errors.New("pengajuan lembur tidak ditemukan")
	}
	if existing.Status != "pending" {
		return errors.New("pengajuan lembur sudah diproses, tidak dapat dibatalkan")
	}

	err = repository.CancelOvertimeRequest(ctx, id, employeeID)
	if err != nil {
		return fmt.Errorf("gagal membatalkan pengajuan lembur: %w", err)
	}

	// Cancel workflow tracking (non-blocking, ignore errors)
	wfSvc := NewApprovalWorkflowService()
	_ = wfSvc.CancelWorkflowTracking(ctx, "overtime", id)

	return nil
}

func (s *OvertimeService) GetCalculation(ctx context.Context, id string) (*models.OvertimeCalculationResponse, error) {
	r, err := repository.GetOvertimeCalculation(ctx, id)
	if err != nil {
		return nil, errors.New("perhitungan lembur belum tersedia (pengajuan harus disetujui terlebih dahulu)")
	}
	return r, nil
}
