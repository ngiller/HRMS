package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type PayrollService struct{}

func NewPayrollService() *PayrollService {
	return &PayrollService{}
}

func (s *PayrollService) ListPeriods(ctx context.Context, page, perPage int) (*models.PayrollPeriodListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	periods, total, err := repository.ListPayrollPeriods(ctx, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode penggajian: %w", err)
	}

	if periods == nil {
		periods = []models.PayrollPeriodSummary{}
	}

	return &models.PayrollPeriodListResponse{
		Periods: periods,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *PayrollService) GetPeriod(ctx context.Context, periodID string) (*models.PayrollPeriod, error) {
	period, err := repository.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode: %w", err)
	}
	if period == nil {
		return nil, errors.New("periode tidak ditemukan")
	}
	return period, nil
}

func (s *PayrollService) CreatePeriod(ctx context.Context, req *models.CreatePayrollPeriodRequest, userID string) (*models.PayrollPeriod, error) {
	// Validasi
	if req.PeriodName == "" {
		return nil, errors.New("nama periode harus diisi")
	}
	if req.Month < 1 || req.Month > 12 {
		return nil, errors.New("bulan tidak valid (1-12)")
	}
	if req.Year < 2020 || req.Year > 2099 {
		return nil, errors.New("tahun tidak valid")
	}

	// Cek duplikasi
	existingPeriods, _, err := repository.ListPayrollPeriods(ctx, 1, 100)
	if err == nil {
		for _, p := range existingPeriods {
			if p.Month == req.Month && p.Year == req.Year {
				return nil, fmt.Errorf("periode %s %d sudah ada", monthName(req.Month), req.Year)
			}
		}
	}

	period, err := repository.CreatePayrollPeriod(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat periode: %w", err)
	}
	return period, nil
}

func (s *PayrollService) CalculatePayroll(ctx context.Context, periodID string, req *models.CalculatePayrollRequest, userID string) (*models.PayrollPeriod, error) {
	// Validate period exists and is in draft status
	period, err := repository.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode: %w", err)
	}
	if period == nil {
		return nil, errors.New("periode tidak ditemukan")
	}
	if period.Status != models.PayrollStatusDraft && period.Status != models.PayrollStatusCalculated {
		return nil, errors.New("hanya periode draft atau calculated yang bisa dihitung ulang")
	}

	// Determine which employees to calculate for
	var employeeIDs []string
	if req.EmployeeID != "" {
		employeeIDs = []string{req.EmployeeID}
	} else {
		// Get all active employees
		ids, err := repository.GetAllActiveEmployeeIDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat data karyawan: %w", err)
		}
		employeeIDs = ids
	}

	if len(employeeIDs) == 0 {
		return nil, errors.New("tidak ada karyawan aktif untuk dihitung")
	}

	// Calculate for each employee
	var calcErrors []string
	for _, empID := range employeeIDs {
		err := s.calculateSingleEmployee(ctx, periodID, empID, userID, req)
		if err != nil {
			calcErrors = append(calcErrors, fmt.Sprintf("karyawan %s: %v", empID, err))
		}
	}

	// Update period summary
	if err := repository.UpdatePayrollPeriodSummary(ctx, periodID); err != nil {
		return nil, fmt.Errorf("gagal mengupdate ringkasan: %w", err)
	}

	// Update period status to calculated
	if err := repository.UpdatePayrollPeriodStatus(ctx, periodID, models.PayrollStatusCalculated, userID); err != nil {
		return nil, fmt.Errorf("gagal mengupdate status periode: %w", err)
	}

	if len(calcErrors) > 0 {
		// Return partial success
		updatedPeriod, _ := repository.GetPayrollPeriod(ctx, periodID)
		return updatedPeriod, fmt.Errorf("beberapa karyawan gagal dihitung: %s", strings.Join(calcErrors, "; "))
	}

	updatedPeriod, err := repository.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode setelah kalkulasi: %w", err)
	}

	return updatedPeriod, nil
}

func (s *PayrollService) calculateSingleEmployee(ctx context.Context, periodID, employeeID, userID string, req *models.CalculatePayrollRequest) error {
	// Get employee details
	dailyWage, status, err := repository.GetEmployeeDetailsForPayroll(ctx, employeeID)
	if err != nil {
		return err
	}

	// Get latest base salary
	baseSalary, err := repository.GetLatestBaseSalary(ctx, employeeID)
	if err != nil {
		return err
	}

	// Override with request values if provided
	if req.BaseSalary > 0 {
		baseSalary = req.BaseSalary
	}
	if req.DailyWage > 0 {
		dailyWage = req.DailyWage
	}

	// Get active salary components (split by type)
	allowanceComponents, deductionComponents, err := repository.GetActiveSalaryComponents(ctx, employeeID)
	if err != nil {
		return err
	}

	// Deduction-type components -> pass as other_deductions
	var deductionTotal float64
	for _, c := range deductionComponents {
		deductionTotal += c.Amount
	}
	otherDeductions := req.OtherDeductions + deductionTotal

	// If employment_status is 'harian', calculate daily wage based on days worked
	totalDaysWorked := req.TotalDaysWorked
	if totalDaysWorked <= 0 && status == "harian" {
		totalDaysWorked = 22 // default for monthly calculation
	}
	if totalDaysWorked <= 0 {
		totalDaysWorked = 0
	}

	// Calculate daily wage portion
	var dailyWageTotal float64
	if dailyWage > 0 && totalDaysWorked > 0 {
		dailyWageTotal = dailyWage * float64(totalDaysWorked)
	}

	// Call the PL/pgSQL function (with user context for audit triggers)
	_, err = repository.CallCalculateEmployeePayroll(
		ctx, userID, periodID, employeeID,
		baseSalary,
		dailyWageTotal,
		totalDaysWorked,
		allowanceComponents,
		req.OvertimePay,
		req.THRAmount,
		req.BonusAmount,
		req.LoanDeduction,
		otherDeductions,
	)
	return err
}

func (s *PayrollService) ListPayrollItems(ctx context.Context, periodID string, page, perPage int) (*models.PayrollItemsListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	items, total, err := repository.ListPayrollItems(ctx, periodID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat item penggajian: %w", err)
	}

	if items == nil {
		items = []models.PayrollItem{}
	}

	// Sort items: department -> name
	sort.Slice(items, func(i, j int) bool {
		if items[i].DepartmentName != items[j].DepartmentName {
			return items[i].DepartmentName < items[j].DepartmentName
		}
		return items[i].EmployeeName < items[j].EmployeeName
	})

	return &models.PayrollItemsListResponse{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *PayrollService) ApprovePeriod(ctx context.Context, periodID, userID string) (*models.PayrollPeriod, error) {
	period, err := repository.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode: %w", err)
	}
	if period == nil {
		return nil, errors.New("periode tidak ditemukan")
	}
	if period.Status != models.PayrollStatusCalculated {
		return nil, errors.New("hanya periode dengan status 'calculated' yang bisa disetujui")
	}

	if err := repository.UpdatePayrollPeriodStatus(ctx, periodID, models.PayrollStatusApproved, userID); err != nil {
		return nil, fmt.Errorf("gagal menyetujui periode: %w", err)
	}

	return repository.GetPayrollPeriod(ctx, periodID)
}

func (s *PayrollService) PayPeriod(ctx context.Context, periodID, userID string) (*models.PayrollPeriod, error) {
	period, err := repository.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode: %w", err)
	}
	if period == nil {
		return nil, errors.New("periode tidak ditemukan")
	}
	if period.Status != models.PayrollStatusApproved {
		return nil, errors.New("hanya periode dengan status 'approved' yang bisa dibayarkan")
	}

	if err := repository.UpdatePayrollPeriodStatus(ctx, periodID, models.PayrollStatusPaid, userID); err != nil {
		return nil, fmt.Errorf("gagal membayarkan periode: %w", err)
	}

	return repository.GetPayrollPeriod(ctx, periodID)
}

func (s *PayrollService) GetPayslip(ctx context.Context, periodID, employeeID string) (*models.PayslipResponse, error) {
	payslip, err := repository.GetPayslip(ctx, periodID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat slip gaji: %w", err)
	}
	if payslip == nil {
		return nil, errors.New("slip gaji tidak ditemukan")
	}
	return payslip, nil
}

func (s *PayrollService) ListMyPayslips(ctx context.Context, employeeID string) ([]models.PayslipResponse, error) {
	payslips, err := repository.ListMyPayslips(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat slip gaji: %w", err)
	}
	if payslips == nil {
		payslips = []models.PayslipResponse{}
	}
	return payslips, nil
}

// Helpers
func monthName(m int) string {
	names := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if m >= 1 && m <= 12 {
		return names[m]
	}
	return fmt.Sprintf("%d", m)
}


