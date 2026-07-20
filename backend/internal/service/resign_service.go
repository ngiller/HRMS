package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type ResignService struct{}

func NewResignService() *ResignService {
	return &ResignService{}
}

func (s *ResignService) Create(ctx context.Context, employeeID string, req *models.CreateResignRequest) (*models.ResignRequest, error) {
	if req.LastWorkingDate == "" {
		return nil, errors.New("tanggal terakhir kerja harus diisi")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan resign harus diisi")
	}
	if req.ResignType == "" {
		req.ResignType = "voluntary"
	}

	r, err := repository.CreateResignRequest(ctx, employeeID, req)
	if err != nil {
		return nil, err
	}

	// Create default exit clearance items
	_ = repository.CreateExitClearanceItems(ctx, r.ID.String())

	// Initiate approval workflow if configured
	wfSvc := NewApprovalWorkflowService()
	_, wfErr := wfSvc.ResolveWorkflowForRequest(ctx, "resign", r.ID.String(), employeeID)
	if wfErr != nil {
		// If no workflow, auto-approve (HR processes it directly)
	}

	return r, nil
}

func (s *ResignService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.ResignListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return repository.ListResignRequests(ctx, page, perPage, status, employeeID)
}

func (s *ResignService) Get(ctx context.Context, id string) (*models.ResignRequest, error) {
	return repository.GetResignRequest(ctx, id)
}

func (s *ResignService) Approve(ctx context.Context, id, userID string) (*models.ResignRequest, error) {
	r, err := repository.GetResignRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.Status != "pending" {
		return nil, errors.New("pengajuan resign sudah diproses")
	}

	// Check all clearance items are done
	allDone, err := repository.CheckAllClearanceItemsChecked(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa clearance: %w", err)
	}
	if !allDone {
		return nil, errors.New("semua item exit clearance harus diselesaikan terlebih dahulu")
	}

	err = repository.UpdateResignRequestStatus(ctx, id, "approved", userID, "")
	if err != nil {
		return nil, err
	}

	// Process the resignation: deactivate employee
	err = repository.ProcessEmployeeResignation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses resign: %w", err)
	}

	return repository.GetResignRequest(ctx, id)
}

func (s *ResignService) Reject(ctx context.Context, id, userID, reason string) (*models.ResignRequest, error) {
	r, err := repository.GetResignRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.Status != "pending" {
		return nil, errors.New("pengajuan resign sudah diproses")
	}

	err = repository.UpdateResignRequestStatus(ctx, id, "rejected", userID, reason)
	if err != nil {
		return nil, err
	}
	return repository.GetResignRequest(ctx, id)
}

// ==================== Exit Clearance ====================

func (s *ResignService) ListClearanceItems(ctx context.Context, resignID string) (*models.ExitClearanceListResponse, error) {
	items, err := repository.ListExitClearanceItems(ctx, resignID)
	if err != nil {
		return nil, err
	}
	return &models.ExitClearanceListResponse{Items: items}, nil
}

func (s *ResignService) UpdateClearanceItem(ctx context.Context, itemID, userID string, isChecked bool) (*models.ExitClearanceItem, error) {
	err := repository.UpdateExitClearanceItem(ctx, itemID, userID, isChecked)
	if err != nil {
		return nil, err
	}

	// Return updated item
	// We need to re-fetch, so let's query directly
	resignID, err := s.getResignIDByItemID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	items, err := repository.ListExitClearanceItems(ctx, resignID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID.String() == itemID || item.ID.String() == itemID {
			return &item, nil
		}
	}
	return nil, errors.New("item clearance tidak ditemukan")
}

func (s *ResignService) getResignIDByItemID(ctx context.Context, itemID string) (string, error) {
	var resignID string
	err := repository.GetResignIDByItemID(ctx, itemID, &resignID)
	return resignID, err
}
