package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type LoanService struct{}

func NewLoanService() *LoanService {
	return &LoanService{}
}

func (s *LoanService) List(ctx context.Context, page, perPage int, status, employeeID string) (*models.LoanListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	loans, total, err := repository.ListLoans(ctx, page, perPage, status, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data pinjaman: %w", err)
	}
	if loans == nil {
		loans = []models.LoanSummary{}
	}

	return &models.LoanListResponse{
		Loans:   loans,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *LoanService) Get(ctx context.Context, id string) (*models.Loan, error) {
	l, err := repository.GetLoanByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data pinjaman")
	}
	if l == nil {
		return nil, errors.New("pinjaman tidak ditemukan")
	}
	return l, nil
}

func (s *LoanService) Create(ctx context.Context, employeeID string, req *models.CreateLoanRequest) (*models.Loan, error) {
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus diisi")
	}
	if req.Amount <= 0 {
		return nil, errors.New("jumlah pinjaman harus lebih dari 0")
	}
	if req.InstallmentCount < 1 || req.InstallmentCount > 24 {
		return nil, errors.New("tenor pinjaman 1-24 bulan")
	}
	if req.LoanType == "" {
		return nil, errors.New("tipe pinjaman harus diisi")
	}
	if req.LoanType != "regular" && req.LoanType != "emergency" && req.LoanType != "education" {
		return nil, errors.New("tipe pinjaman tidak valid (regular/emergency/education)")
	}
	if req.Purpose == "" {
		return nil, errors.New("tujuan pinjaman harus diisi")
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "payroll_deduction"
	}

	l, err := repository.CreateLoan(ctx, employeeID, req, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pinjaman: %w", err)
	}
	return l, nil
}

func (s *LoanService) Approve(ctx context.Context, id, approverID string) (*models.Loan, error) {
	existing, err := repository.GetLoanByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data pinjaman")
	}
	if existing == nil {
		return nil, errors.New("pinjaman tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("pinjaman sudah diproses")
	}

	l, err := repository.UpdateLoanStatus(ctx, id, "approved", "", approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menyetujui pinjaman: %w", err)
	}
	return l, nil
}

func (s *LoanService) Reject(ctx context.Context, id, rejectionReason, approverID string) (*models.Loan, error) {
	existing, err := repository.GetLoanByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data pinjaman")
	}
	if existing == nil {
		return nil, errors.New("pinjaman tidak ditemukan")
	}
	if existing.Status != "pending" {
		return nil, errors.New("pinjaman sudah diproses")
	}

	l, err := repository.UpdateLoanStatus(ctx, id, "rejected", rejectionReason, approverID)
	if err != nil {
		return nil, fmt.Errorf("gagal menolak pinjaman: %w", err)
	}
	return l, nil
}

func (s *LoanService) Disburse(ctx context.Context, id, approverID, date string) (*models.Loan, error) {
	existing, err := repository.GetLoanByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data pinjaman")
	}
	if existing == nil {
		return nil, errors.New("pinjaman tidak ditemukan")
	}
	if existing.Status != "approved" {
		return nil, errors.New("pinjaman harus disetujui terlebih dahulu")
	}

	l, err := repository.DisburseLoan(ctx, id, approverID, date)
	if err != nil {
		return nil, fmt.Errorf("gagal mencairkan pinjaman: %w", err)
	}
	return l, nil
}

func (s *LoanService) Stats(ctx context.Context) (*models.LoanStatsResponse, error) {
	stats, err := repository.GetLoanStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat statistik pinjaman: %w", err)
	}
	return stats, nil
}
