package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type CompanyService struct{}

func NewCompanyService() *CompanyService {
	return &CompanyService{}
}

func (s *CompanyService) GetSettings(ctx context.Context) (*models.Company, error) {
	company, err := repository.GetCompany(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat pengaturan perusahaan: %w", err)
	}
	if company == nil {
		return nil, errors.New("data perusahaan tidak ditemukan")
	}
	return company, nil
}

func (s *CompanyService) UpdateSettings(ctx context.Context, req *models.UpdateCompanySettingsRequest) (*models.Company, error) {
	if req == nil {
		return nil, errors.New("request tidak boleh kosong")
	}
	company, err := repository.UpdateCompanySettings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate pengaturan: %w", err)
	}
	return company, nil
}

// CalculateTHRForPeriod calculates THR for all employees in a given period.
// THR = (months_worked / 12) * base_salary, capped at 1 month salary.
func (s *CompanyService) CalculateTHRForPeriod(ctx context.Context, periodID string) (*models.THRCalculationResponse, error) {
	period, err := repository.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat periode: %w", err)
	}
	if period == nil {
		return nil, errors.New("periode tidak ditemukan")
	}

	// Get all active employees with their join dates and base salaries
	employees, err := repository.GetAllEmployeesForTHR(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data karyawan: %w", err)
	}

	var results []models.THRCalculationItem
	totalTHR := 0.0

	periodStart := time.Date(period.Year, time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)

	for _, emp := range employees {
		joinDate, err := time.Parse("2006-01-02T15:04:05Z", emp.JoinDate)
		if err != nil {
			joinDate, err = time.Parse("2006-01-02", emp.JoinDate)
			if err != nil {
				continue
			}
		}

		// Calculate months worked up to the period
		monthsWorked := monthsBetween(joinDate, periodStart)
		if monthsWorked < 1 {
			continue
		}
		if monthsWorked > 12 {
			monthsWorked = 12
		}

		// THR = (months_worked / 12) * base_salary
		thrAmount := math.Round((float64(monthsWorked)/12.0)*emp.BaseSalary*100) / 100

		results = append(results, models.THRCalculationItem{
			EmployeeID:   emp.EmployeeID,
			EmployeeName: emp.EmployeeName,
			JoinDate:     emp.JoinDate,
			MonthsWorked: monthsWorked,
			BaseSalary:   emp.BaseSalary,
			THRAmount:    thrAmount,
		})
		totalTHR += thrAmount
	}

	totalTHR = math.Round(totalTHR*100) / 100

	return &models.THRCalculationResponse{
		PeriodID:      periodID,
		PeriodName:    period.PeriodName,
		TotalTHR:      totalTHR,
		EmployeeCount: len(results),
		Items:         results,
	}, nil
}

// monthsBetween calculates the number of whole or partial months between two dates.
// A partial month counts as 1 month if the employee started before the 15th.
func monthsBetween(from, to time.Time) int {
	if from.After(to) {
		return 0
	}

	// Count full years and months
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	totalMonths := years*12 + months

	// If the day of month of from is > to's day of month, we haven't completed the month
	if from.Day() > to.Day() {
		totalMonths--
	}

	if totalMonths < 0 {
		return 0
	}
	// Always round up to 1 if the employee has any time in the month
	if totalMonths == 0 && from.Day() <= to.Day() {
		totalMonths = 1
	}

	return totalMonths
}
