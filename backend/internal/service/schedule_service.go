package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type ScheduleService struct{}

func NewScheduleService() *ScheduleService {
	return &ScheduleService{}
}

// ============================================================
// TEMPLATE METHODS
// ============================================================

func (s *ScheduleService) ListTemplates(ctx context.Context, page, perPage int, search string) (*models.ScheduleTemplateListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	templates, total, err := repository.ListScheduleTemplates(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat template jadwal: %w", err)
	}
	if templates == nil {
		templates = []models.ScheduleTemplateSummary{}
	}

	return &models.ScheduleTemplateListResponse{
		Templates: templates,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *ScheduleService) GetAllTemplates(ctx context.Context) ([]models.ScheduleTemplateSummary, error) {
	templates, err := repository.GetAllScheduleTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar template: %w", err)
	}
	if templates == nil {
		templates = []models.ScheduleTemplateSummary{}
	}
	return templates, nil
}

func (s *ScheduleService) GetTemplate(ctx context.Context, id string) (*models.ScheduleTemplate, error) {
	t, err := repository.GetScheduleTemplateByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat template jadwal")
	}
	if t == nil {
		return nil, errors.New("template jadwal tidak ditemukan")
	}
	return t, nil
}

func (s *ScheduleService) CreateTemplate(ctx context.Context, req *models.CreateScheduleTemplateRequest) (*models.ScheduleTemplate, error) {
	if req.Name == "" {
		return nil, errors.New("nama template harus diisi")
	}
	if req.ScheduleType == "" {
		req.ScheduleType = "weekly"
	}
	if req.ScheduleType != "weekly" && req.ScheduleType != "shift" && req.ScheduleType != "flexible" {
		return nil, errors.New("tipe template tidak valid (weekly, shift, flexible)")
	}
	if len(req.Days) == 0 {
		return nil, errors.New("minimal 1 hari harus diisi")
	}

	exists, err := repository.CheckScheduleTemplateNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi template: %w", err)
	}
	if exists {
		return nil, errors.New("nama template sudah digunakan")
	}

	// Validasi days
	for i, d := range req.Days {
		if d.StartTime == "" {
			req.Days[i].StartTime = "08:00"
		}
		if d.EndTime == "" {
			req.Days[i].EndTime = "17:00"
		}
		if d.BreakStart == "" {
			req.Days[i].BreakStart = "12:00"
		}
		if d.BreakEnd == "" {
			req.Days[i].BreakEnd = "13:00"
		}
		if d.LateToleranceMinutes == 0 {
			req.Days[i].LateToleranceMinutes = 15
		}
		if d.EarlyLeaveTolerance == 0 {
			req.Days[i].EarlyLeaveTolerance = 15
		}
	}

	return repository.CreateScheduleTemplate(ctx, req)
}

func (s *ScheduleService) UpdateTemplate(ctx context.Context, id string, req *models.UpdateScheduleTemplateRequest) (*models.ScheduleTemplate, error) {
	if req.Name != nil && *req.Name == "" {
		return nil, errors.New("nama template tidak boleh kosong")
	}

	existing, err := repository.GetScheduleTemplateByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("template jadwal tidak ditemukan")
	}

	if req.Name != nil && *req.Name != existing.Name {
		exists, err := repository.CheckScheduleTemplateNameExists(ctx, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi template: %w", err)
		}
		if exists {
			return nil, errors.New("nama template sudah digunakan")
		}
	}

	return repository.UpdateScheduleTemplate(ctx, id, req)
}

func (s *ScheduleService) DeleteTemplate(ctx context.Context, id, userID string) error {
	existing, err := repository.GetScheduleTemplateByID(ctx, id)
	if err != nil || existing == nil {
		return errors.New("template jadwal tidak ditemukan")
	}
	return repository.DeleteScheduleTemplate(ctx, id, userID)
}

// ============================================================
// EMPLOYEE SCHEDULE METHODS (Level 2 + Level 3)
// ============================================================

func (s *ScheduleService) ListEmployeeSchedules(ctx context.Context, employeeID string, page, perPage int) (*models.EmployeeScheduleListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	schedules, total, err := repository.ListEmployeeSchedules(ctx, employeeID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat jadwal karyawan: %w", err)
	}
	if schedules == nil {
		schedules = []models.EmployeeScheduleSummary{}
	}

	return &models.EmployeeScheduleListResponse{
		Schedules: schedules,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *ScheduleService) GetEmployeeSchedule(ctx context.Context, id string) (*models.EmployeeSchedule, error) {
	es, err := repository.GetEmployeeScheduleByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat jadwal karyawan")
	}
	if es == nil {
		return nil, errors.New("jadwal karyawan tidak ditemukan")
	}
	return es, nil
}

func (s *ScheduleService) CreateEmployeeSchedule(ctx context.Context, req *models.CreateEmployeeScheduleRequest, userID string) (*models.EmployeeSchedule, error) {
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus dipilih")
	}
	if req.DayOfWeek == nil && req.SpecificDate == nil {
		return nil, errors.New("pilih day_of_week (periodik) atau specific_date (satu hari), salah satu harus diisi")
	}
	if req.EffectiveFrom == "" {
		return nil, errors.New("tanggal berlaku harus diisi")
	}

	return repository.CreateEmployeeSchedule(ctx, req, userID)
}

func (s *ScheduleService) UpdateEmployeeSchedule(ctx context.Context, id string, req *models.UpdateEmployeeScheduleRequest, userID string) (*models.EmployeeSchedule, error) {
	existing, err := repository.GetEmployeeScheduleByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("jadwal karyawan tidak ditemukan")
	}
	return repository.UpdateEmployeeSchedule(ctx, id, req, userID)
}

func (s *ScheduleService) DeleteEmployeeSchedule(ctx context.Context, id, userID string) error {
	existing, err := repository.GetEmployeeScheduleByID(ctx, id)
	if err != nil || existing == nil {
		return errors.New("jadwal karyawan tidak ditemukan")
	}
	return repository.DeleteEmployeeSchedule(ctx, id, userID)
}

// ============================================================
// RESOLVE SCHEDULE — Untuk absensi check-in
// ============================================================

func (s *ScheduleService) ResolveSchedule(ctx context.Context, employeeID, date string) (*models.ResolvedSchedule, error) {
	rs, err := repository.ResolveEmployeeScheduleForDate(ctx, employeeID, date)
	if err != nil {
		return nil, fmt.Errorf("gagal menentukan jadwal: %w", err)
	}
	return rs, nil
}
