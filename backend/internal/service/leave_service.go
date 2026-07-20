package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type LeaveService struct{}

func NewLeaveService() *LeaveService {
	return &LeaveService{}
}

// ─── Leave Types ──────────────────────────────────────────────

func (s *LeaveService) GetLeaveTypes(ctx context.Context) ([]models.LeaveTypeSummary, error) {
	types, err := repository.GetAllLeaveTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat jenis cuti: %w", err)
	}
	return types, nil
}

// ─── Leave Balances ───────────────────────────────────────────

func (s *LeaveService) GetMyBalances(ctx context.Context, employeeID string) (*models.LeaveBalanceResponse, error) {
	if employeeID == "" {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	balances, err := repository.GetLeaveBalancesByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat sisa cuti: %w", err)
	}
	return &models.LeaveBalanceResponse{
		Balances: balances,
		Total:    len(balances),
	}, nil
}

func (s *LeaveService) GetAllBalances(ctx context.Context, year int) (*models.LeaveBalanceResponse, error) {
	if year == 0 {
		year = time.Now().Year()
	}
	balances, err := repository.GetAllLeaveBalances(ctx, year)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat sisa cuti: %w", err)
	}
	return &models.LeaveBalanceResponse{
		Balances: balances,
		Total:    len(balances),
	}, nil
}

// ─── Leave Requests ───────────────────────────────────────────

func (s *LeaveService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.LeaveRequestListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	requests, total, err := repository.ListLeaveRequests(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data cuti: %w", err)
	}

	return &models.LeaveRequestListResponse{
		LeaveRequests: requests,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
	}, nil
}

func (s *LeaveService) Get(ctx context.Context, id string) (*models.LeaveRequest, error) {
	r, err := repository.GetLeaveRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data cuti")
	}
	if r == nil {
		return nil, errors.New("pengajuan cuti tidak ditemukan")
	}
	return r, nil
}

func (s *LeaveService) Create(ctx context.Context, employeeID string, req *models.CreateLeaveRequestReq) (*models.LeaveRequest, error) {
	if req.LeaveTypeID == "" {
		return nil, errors.New("jenis cuti harus dipilih")
	}
	if req.StartDate == "" {
		return nil, errors.New("tanggal mulai harus diisi")
	}
	if req.EndDate == "" {
		return nil, errors.New("tanggal selesai harus diisi")
	}
	if req.TotalDays <= 0 {
		return nil, errors.New("jumlah hari harus lebih dari 0")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan cuti harus diisi")
	}

	// Validasi eligibility berdasarkan jenis cuti
	if err := s.validateLeaveEligibility(ctx, employeeID, req.LeaveTypeID); err != nil {
		return nil, err
	}

	r, err := repository.CreateLeaveRequest(ctx, employeeID, employeeID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengajuan cuti: %w", err)
	}

	// Initialize workflow tracking (non-blocking, ignore errors)
	err = s.initWorkflowTracking(ctx, "leave", r.ID.String(), employeeID)
	if err != nil {
		// Log warning but don't fail - leave was created successfully
		fmt.Printf("[WARN] Leave workflow init: %v\n", err)
	}

	return r, nil
}

// validateLeaveEligibility checks if the employee is eligible for the selected leave type
// based on gender and marital status rules.
func (s *LeaveService) validateLeaveEligibility(ctx context.Context, employeeID, leaveTypeID string) error {
	// Fetch leave type to get the code
	leaveType, err := repository.GetLeaveTypeByID(ctx, leaveTypeID)
	if err != nil {
		return fmt.Errorf("gagal memuat data jenis cuti: %w", err)
	}
	if leaveType == nil {
		return errors.New("jenis cuti tidak ditemukan")
	}

	// Fetch employee data
	employee, err := repository.GetEmployeeByIDRepo(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("gagal memuat data karyawan: %w", err)
	}
	if employee == nil {
		return errors.New("karyawan tidak ditemukan")
	}

	switch leaveType.Code {
	case "melahirkan":
		// Cuti Hamil/Melahirkan: only perempuan + menikah + is_pregnant
		if employee.Gender != "perempuan" {
			return errors.New("cuti hamil/melahirkan hanya untuk karyawan perempuan")
		}
		if employee.MaritalStatus != "menikah" {
			return errors.New("cuti hamil/melahirkan hanya untuk karyawan yang sudah menikah")
		}
		if !employee.IsPregnant {
			return errors.New("cuti hamil/melahirkan hanya untuk karyawan yang sedang hamil. Silakan hubungi HR untuk memperbarui status kehamilan")
		}
	case "keguguran":
		// Cuti Keguguran: only perempuan
		if employee.Gender != "perempuan" {
			return errors.New("cuti keguguran hanya untuk karyawan perempuan")
		}
	case "menikah":
		// Cuti Menikah: only lajang (belum menikah)
		if employee.MaritalStatus != "lajang" {
			return errors.New("cuti menikah hanya untuk karyawan yang belum menikah (lajang)")
		}
	}

	return nil
}

func (s *LeaveService) initWorkflowTracking(ctx context.Context, entityType, entityID, employeeID string) error {
	wfSvc := NewApprovalWorkflowService()
	_, err := wfSvc.ResolveWorkflowForRequest(ctx, entityType, entityID, employeeID)
	return err
}

func (s *LeaveService) Approve(ctx context.Context, id, approverID string) (*models.LeaveRequest, error) {
	existing, err := repository.GetLeaveRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data cuti")
	}
	if existing == nil {
		return nil, errors.New("pengajuan cuti tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("pengajuan cuti sudah diproses")
	}

	r, err := repository.UpdateLeaveStatus(ctx, id, "approved", "", approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menyetujui pengajuan cuti: %w", err)
	}
	return r, nil
}

func (s *LeaveService) Reject(ctx context.Context, id, rejectionReason, approverID string) (*models.LeaveRequest, error) {
	existing, err := repository.GetLeaveRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data cuti")
	}
	if existing == nil {
		return nil, errors.New("pengajuan cuti tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("pengajuan cuti sudah diproses")
	}

	r, err := repository.UpdateLeaveStatus(ctx, id, "rejected", rejectionReason, approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menolak pengajuan cuti: %w", err)
	}
	return r, nil
}

func (s *LeaveService) Cancel(ctx context.Context, id, employeeID, cancelReason string) error {
	existing, err := repository.GetLeaveRequestByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data cuti")
	}
	if existing == nil {
		return errors.New("pengajuan cuti tidak ditemukan")
	}
	if existing.Status != "pending" {
		return errors.New("pengajuan cuti sudah diproses, tidak dapat dibatalkan")
	}

	err = repository.CancelLeaveRequest(ctx, id, employeeID, cancelReason)
	if err != nil {
		return fmt.Errorf("gagal membatalkan pengajuan cuti: %w", err)
	}

	// Cancel workflow tracking (non-blocking, ignore errors)
	wfSvc := NewApprovalWorkflowService()
	_ = wfSvc.CancelWorkflowTracking(ctx, "leave", id)

	return nil
}
