package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type WorkScheduleService struct{}

func NewWorkScheduleService() *WorkScheduleService {
	return &WorkScheduleService{}
}

func (s *WorkScheduleService) List(ctx context.Context, page, perPage int, search string) (*models.WorkScheduleListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	schedules, total, err := repository.ListWorkSchedules(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data jadwal kerja: %w", err)
	}
	if schedules == nil {
		schedules = []models.WorkScheduleSummary{}
	}

	return &models.WorkScheduleListResponse{
		WorkSchedules: schedules,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
	}, nil
}

func (s *WorkScheduleService) GetAll(ctx context.Context) ([]models.WorkScheduleSummary, error) {
	schedules, err := repository.GetAllWorkSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar jadwal kerja: %w", err)
	}
	if schedules == nil {
		schedules = []models.WorkScheduleSummary{}
	}
	return schedules, nil
}

func (s *WorkScheduleService) Get(ctx context.Context, id string) (*models.WorkSchedule, error) {
	ws, err := repository.GetWorkScheduleByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data jadwal kerja")
	}
	if ws == nil {
		return nil, errors.New("jadwal kerja tidak ditemukan")
	}
	return ws, nil
}

func (s *WorkScheduleService) Create(ctx context.Context, req *models.CreateWorkScheduleRequest, userID string) (*models.WorkSchedule, error) {
	if req.Name == "" {
		return nil, errors.New("nama jadwal kerja harus diisi")
	}
	if req.ScheduleType == "" {
		return nil, errors.New("tipe jadwal harus dipilih")
	}
	if req.ScheduleType != "5_day" && req.ScheduleType != "6_day" && req.ScheduleType != "shift" {
		return nil, errors.New("tipe jadwal tidak valid (5_day, 6_day, atau shift)")
	}

	if req.MondayStart == "" {
		req.MondayStart = "08:00"
	}
	if req.MondayEnd == "" {
		req.MondayEnd = "17:00"
	}
	if req.TuesdayStart == "" {
		req.TuesdayStart = "08:00"
	}
	if req.TuesdayEnd == "" {
		req.TuesdayEnd = "17:00"
	}
	if req.WednesdayStart == "" {
		req.WednesdayStart = "08:00"
	}
	if req.WednesdayEnd == "" {
		req.WednesdayEnd = "17:00"
	}
	if req.ThursdayStart == "" {
		req.ThursdayStart = "08:00"
	}
	if req.ThursdayEnd == "" {
		req.ThursdayEnd = "17:00"
	}
	if req.FridayStart == "" {
		req.FridayStart = "08:00"
	}
	if req.FridayEnd == "" {
		req.FridayEnd = "17:00"
	}
	if req.BreakStart == "" {
		req.BreakStart = "12:00"
	}
	if req.BreakEnd == "" {
		req.BreakEnd = "13:00"
	}
	if req.LateToleranceMinutes == 0 {
		req.LateToleranceMinutes = 15
	}
	if req.EarlyLeaveTolerance == 0 {
		req.EarlyLeaveTolerance = 15
	}
	if req.WeeklyHours == 0 {
		req.WeeklyHours = 40
	}

	exists, err := repository.CheckWorkScheduleNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi jadwal kerja: %w", err)
	}
	if exists {
		return nil, errors.New("nama jadwal kerja sudah digunakan")
	}

	ws, err := repository.CreateWorkSchedule(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat jadwal kerja: %w", err)
	}
	return ws, nil
}

func (s *WorkScheduleService) Update(ctx context.Context, id string, req *models.UpdateWorkScheduleRequest, userID string) (*models.WorkSchedule, error) {
	existing, err := repository.GetWorkScheduleByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data jadwal kerja")
	}
	if existing == nil {
		return nil, errors.New("jadwal kerja tidak ditemukan")
	}

	if req.Name != nil && *req.Name != existing.Name {
		exists, err := repository.CheckWorkScheduleNameExists(ctx, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi jadwal kerja: %w", err)
		}
		if exists {
			return nil, errors.New("nama jadwal kerja sudah digunakan")
		}
	}

	ws, err := repository.UpdateWorkSchedule(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui jadwal kerja: %w", err)
	}
	return ws, nil
}

func (s *WorkScheduleService) Delete(ctx context.Context, id, userID string) error {
	existing, err := repository.GetWorkScheduleByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data jadwal kerja")
	}
	if existing == nil {
		return errors.New("jadwal kerja tidak ditemukan")
	}

	used, err := repository.CheckWorkScheduleUsedByDepartments(ctx, id)
	if err != nil {
		return fmt.Errorf("gagal validasi jadwal kerja: %w", err)
	}
	if used {
		return errors.New("jadwal kerja masih digunakan oleh departemen, tidak dapat dihapus")
	}

	err = repository.DeleteWorkSchedule(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus jadwal kerja: %w", err)
	}
	return nil
}
