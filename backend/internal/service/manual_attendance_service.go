package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type ManualAttendanceService struct{}

func NewManualAttendanceService() *ManualAttendanceService {
	return &ManualAttendanceService{}
}

func (s *ManualAttendanceService) Create(ctx context.Context, employeeID string, req *models.CreateManualAttendanceRequest) (*models.ManualAttendanceRequest, error) {
	if req.Date == "" {
		return nil, errors.New("tanggal harus diisi")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan harus diisi")
	}
	if req.CheckInTime == "" && req.CheckOutTime == "" {
		return nil, errors.New("setidaknya satu waktu (check-in atau check-out) harus diisi")
	}

	// Check max manual attendance per month from company settings
	// Default is 3 per month
	if err := repository.CheckManualAttendanceCount(ctx, employeeID, req.Date); err != nil {
		return nil, fmt.Errorf("gagal memvalidasi kuota: %w", err)
	}

	r, err := repository.CreateManualAttendanceRequest(ctx, employeeID, req)
	if err != nil {
		return nil, err
	}

	// Initiate approval workflow (if configured). If no workflow, leave as pending for manual approval.
	wfSvc := NewApprovalWorkflowService()
	wfSvc.ResolveWorkflowForRequest(ctx, "manual_attendance", r.ID.String(), employeeID)

	return r, nil
}

func (s *ManualAttendanceService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.ManualAttendanceListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return repository.ListManualAttendanceRequests(ctx, page, perPage, status, employeeID)
}

func (s *ManualAttendanceService) Get(ctx context.Context, id string) (*models.ManualAttendanceRequest, error) {
	return repository.GetManualAttendanceRequest(ctx, id)
}

func (s *ManualAttendanceService) Approve(ctx context.Context, id, userID string) (*models.ManualAttendanceRequest, error) {
	r, err := repository.GetManualAttendanceRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.Status != "pending" {
		return nil, errors.New("pengajuan absensi manual sudah diproses")
	}

	// Prevent self-approval
	if r.EmployeeID.String() == userID {
		return nil, errors.New("tidak dapat menyetujui pengajuan sendiri")
	}

	// Try approval workflow first
	wfSvc := NewApprovalWorkflowService()
	result, err := wfSvc.ProcessApproval(ctx, "manual_attendance", id, userID, "approve", "")
	if err != nil {
		// If no workflow/tracking configured, do direct approval
		if err.Error() == "tracking approval tidak ditemukan" {
			_ = repository.UpdateManualAttendanceRequestStatus(ctx, id, "approved", userID, "")
			updated, _ := repository.GetManualAttendanceRequest(ctx, id)
			_ = repository.CreateAttendanceFromManualRequest(ctx, updated)
			return updated, nil
		}
		return nil, err
	}

	// Only create attendance record if fully approved (no more pending levels)
	updated, _ := repository.GetManualAttendanceRequest(ctx, id)
	if result.FinalStatus == "approved" || updated.Status == "approved" {
		// Update approved_by/approved_at (UpdateEntityStatusTx only sets status, not approved_by)
		_ = repository.UpdateManualAttendanceRequestStatus(ctx, id, "approved", userID, "")
		updated, _ = repository.GetManualAttendanceRequest(ctx, id)
		_ = repository.CreateAttendanceFromManualRequest(ctx, updated)
	}

	return updated, nil
}

func (s *ManualAttendanceService) Reject(ctx context.Context, id, userID, reason string) (*models.ManualAttendanceRequest, error) {
	r, err := repository.GetManualAttendanceRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.Status != "pending" {
		return nil, errors.New("pengajuan absensi manual sudah diproses")
	}

	// Prevent self-rejection
	if r.EmployeeID.String() == userID {
		return nil, errors.New("tidak dapat menolak pengajuan sendiri")
	}

	wfSvc := NewApprovalWorkflowService()
	_, err = wfSvc.ProcessApproval(ctx, "manual_attendance", id, userID, "reject", reason)
	if err != nil {
		// If no workflow/tracking configured, do direct rejection
		if err.Error() == "tracking approval tidak ditemukan" {
			if err := repository.UpdateManualAttendanceRequestStatus(ctx, id, "rejected", userID, reason); err != nil {
				return nil, err
			}
			return repository.GetManualAttendanceRequest(ctx, id)
		}
		return nil, err
	}

	return repository.GetManualAttendanceRequest(ctx, id)
}

func (s *ManualAttendanceService) Cancel(ctx context.Context, id, userID string) error {
	return repository.UpdateManualAttendanceRequestStatus(ctx, id, "cancelled", userID, "")
}
