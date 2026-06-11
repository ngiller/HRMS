package service

import (
	"context"
	"errors"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type EmployeeService struct{}

func NewEmployeeService() *EmployeeService {
	return &EmployeeService{}
}

func (s *EmployeeService) ListEmployees(ctx context.Context, page, perPage int, search string) (*models.EmployeeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	employees, total, err := repository.ListEmployees(ctx, page, perPage, search)
	if err != nil {
		return nil, errors.New("gagal memuat data karyawan")
	}

	if employees == nil {
		employees = []models.EmployeeSummary{}
	}

	return &models.EmployeeListResponse{
		Employees: employees,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *EmployeeService) GetEmployee(ctx context.Context, id string) (*models.Employee, error) {
	employee, err := repository.GetEmployeeByIDRepo(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data karyawan")
	}
	if employee == nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	return employee, nil
}

func (s *EmployeeService) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	stats, err := repository.GetDashboardStats(ctx)
	if err != nil {
		return nil, errors.New("gagal memuat data dashboard")
	}
	return stats, nil
}
