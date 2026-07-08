package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type ShiftChangeService struct{}

func NewShiftChangeService() *ShiftChangeService {
	return &ShiftChangeService{}
}

func (s *ShiftChangeService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.ShiftChangeRequestListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	requests, total, err := repository.ListShiftChangeRequests(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data permintaan shift: %w", err)
	}
	if requests == nil {
		requests = []models.ShiftChangeRequestSummary{}
	}

	return &models.ShiftChangeRequestListResponse{
		ShiftChangeRequests: requests,
		Total:               total,
		Page:                page,
		PerPage:             perPage,
	}, nil
}

func (s *ShiftChangeService) Get(ctx context.Context, id string) (*models.ShiftChangeRequest, error) {
	r, err := repository.GetShiftChangeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data permintaan shift")
	}
	if r == nil {
		return nil, errors.New("permintaan shift tidak ditemukan")
	}
	return r, nil
}

func (s *ShiftChangeService) Create(ctx context.Context, employeeID string, req *models.CreateShiftChangeRequestReq) (*models.ShiftChangeRequest, error) {
	if req.RequestType == "" {
		return nil, errors.New("tipe permintaan harus diisi")
	}
	if req.RequestType != "individual" && req.RequestType != "swap" {
		return nil, errors.New("tipe permintaan tidak valid (individual atau swap)")
	}
	if req.TargetDate == "" {
		return nil, errors.New("tanggal target harus diisi")
	}
	if req.RequestedScheduleID == "" {
		return nil, errors.New("jadwal yang diminta harus dipilih")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan permintaan harus diisi")
	}
	if req.RequestType == "swap" && req.SwapPartnerID == "" {
		return nil, errors.New("partner swap harus dipilih untuk tipe swap")
	}

	exists, err := repository.CheckShiftChangeDuplicatePending(ctx, employeeID, req.TargetDate)
	if err != nil {
		return nil, fmt.Errorf("gagal validasi permintaan shift: %w", err)
	}
	if exists {
		return nil, errors.New("sudah ada permintaan shift pending untuk tanggal ini")
	}

	r, err := repository.CreateShiftChangeRequest(ctx, employeeID, employeeID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat permintaan shift: %w", err)
	}

	// Initialize workflow tracking (non-blocking, ignore errors)
	// Shift changes are unconditional (single-step approval), so conditionValue = 0
	err = s.initWorkflowTracking(ctx, "shift_change", r.ID.String(), employeeID, 0)
	if err != nil {
		fmt.Printf("[WARN] Shift change workflow init: %v\n", err)
	}

	return r, nil
}

func (s *ShiftChangeService) initWorkflowTracking(ctx context.Context, entityType, entityID, employeeID string, conditionValue float64) error {
	wfSvc := NewApprovalWorkflowService()
	_, err := wfSvc.ResolveWorkflowForRequest(ctx, entityType, entityID, employeeID, conditionValue)
	return err
}

func (s *ShiftChangeService) Approve(ctx context.Context, id, approverID string) (*models.ShiftChangeRequest, error) {
	existing, err := repository.GetShiftChangeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data permintaan shift")
	}
	if existing == nil {
		return nil, errors.New("permintaan shift tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("permintaan shift sudah diproses")
	}

	r, err := repository.UpdateShiftChangeStatus(ctx, id, "approved", "", approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menyetujui permintaan shift: %w", err)
	}
	return r, nil
}

func (s *ShiftChangeService) Reject(ctx context.Context, id, rejectionReason, approverID string) (*models.ShiftChangeRequest, error) {
	existing, err := repository.GetShiftChangeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data permintaan shift")
	}
	if existing == nil {
		return nil, errors.New("permintaan shift tidak ditemukan")
	}
	if existing.Status != "pending" && existing.Status != "partner_pending" {
		return nil, errors.New("permintaan shift sudah diproses")
	}

	r, err := repository.UpdateShiftChangeStatus(ctx, id, "rejected", rejectionReason, approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menolak permintaan shift: %w", err)
	}
	return r, nil
}

func (s *ShiftChangeService) ConfirmSwap(ctx context.Context, id, userID string) (*models.ShiftChangeRequest, error) {
	existing, err := repository.GetShiftChangeRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data permintaan shift")
	}
	if existing == nil {
		return nil, errors.New("permintaan shift tidak ditemukan")
	}
	if existing.Status != "partner_pending" {
		return nil, errors.New("permintaan shift tidak dalam status menunggu konfirmasi partner")
	}

	r, err := repository.ConfirmSwapPartner(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal konfirmasi partner swap: %w", err)
	}
	return r, nil
}

func (s *ShiftChangeService) Cancel(ctx context.Context, id, employeeID string) error {
	existing, err := repository.GetShiftChangeRequestByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data permintaan shift")
	}
	if existing == nil {
		return errors.New("permintaan shift tidak ditemukan")
	}
	if existing.Status != "pending" && existing.Status != "partner_pending" {
		return errors.New("permintaan shift sudah diproses, tidak dapat dibatalkan")
	}

	err = repository.CancelShiftChangeRequest(ctx, id, employeeID)
	if err != nil {
		return fmt.Errorf("gagal membatalkan permintaan shift: %w", err)
	}

	// Cancel workflow tracking (non-blocking, ignore errors)
	wfSvc := NewApprovalWorkflowService()
	_ = wfSvc.CancelWorkflowTracking(ctx, "shift_change", id)

	return nil
}
