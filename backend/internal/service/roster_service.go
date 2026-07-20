package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type RosterService struct{}

func NewRosterService() *RosterService {
	return &RosterService{}
}

func (s *RosterService) List(ctx context.Context, page, perPage int, departmentID string) (*models.DepartmentRosterListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	rosters, total, err := repository.ListDepartmentRosters(ctx, page, perPage, departmentID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data roster: %w", err)
	}
	if rosters == nil {
		rosters = []models.DepartmentRosterSummary{}
	}

	return &models.DepartmentRosterListResponse{
		Rosters: rosters,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *RosterService) Get(ctx context.Context, id string) (*models.DepartmentRoster, error) {
	roster, err := repository.GetDepartmentRosterByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data roster")
	}
	if roster == nil {
		return nil, errors.New("roster tidak ditemukan")
	}
	return roster, nil
}

func (s *RosterService) Create(ctx context.Context, req *models.CreateDepartmentRosterRequest, userID string) (*models.DepartmentRoster, error) {
	if req.DepartmentID == "" {
		return nil, errors.New("departemen harus dipilih")
	}
	if req.Name == "" {
		return nil, errors.New("nama roster harus diisi")
	}
	if req.Month < 1 || req.Month > 12 {
		return nil, errors.New("bulan tidak valid (1-12)")
	}
	if req.Year < 2020 {
		return nil, errors.New("tahun tidak valid")
	}

	exists, err := repository.CheckRosterNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi roster: %w", err)
	}
	if exists {
		return nil, errors.New("nama roster sudah digunakan di departemen ini")
	}

	roster, err := repository.CreateDepartmentRoster(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat roster: %w", err)
	}
	return roster, nil
}

func (s *RosterService) Update(ctx context.Context, id string, req *models.UpdateDepartmentRosterRequest, userID string) (*models.DepartmentRoster, error) {
	existing, err := repository.GetDepartmentRosterByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data roster")
	}
	if existing == nil {
		return nil, errors.New("roster tidak ditemukan")
	}

	if req.Name != nil && *req.Name != existing.Name {
		exists, err := repository.CheckRosterNameExists(ctx, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi roster: %w", err)
		}
		if exists {
			return nil, errors.New("nama roster sudah digunakan")
		}
	}

	roster, err := repository.UpdateDepartmentRoster(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui roster: %w", err)
	}
	return roster, nil
}

func (s *RosterService) Delete(ctx context.Context, id, userID string) error {
	existing, err := repository.GetDepartmentRosterByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data roster")
	}
	if existing == nil {
		return errors.New("roster tidak ditemukan")
	}

	err = repository.DeleteDepartmentRoster(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus roster: %w", err)
	}
	return nil
}

// ============================================================
// ROSTER ENTRIES
// ============================================================

func (s *RosterService) ListEntries(ctx context.Context, rosterID string) ([]models.RosterEntry, error) {
	entries, err := repository.ListRosterEntries(ctx, rosterID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat entri roster: %w", err)
	}
	if entries == nil {
		entries = []models.RosterEntry{}
	}
	return entries, nil
}

func (s *RosterService) GetCalendar(ctx context.Context, rosterID string) ([]models.RosterCalendarEntry, error) {
	entries, err := repository.GetRosterCalendar(ctx, rosterID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat kalender roster: %w", err)
	}
	if entries == nil {
		entries = []models.RosterCalendarEntry{}
	}
	return entries, nil
}

func (s *RosterService) CreateEntry(ctx context.Context, req *models.CreateRosterEntryRequest, userID string) (*models.RosterEntry, error) {
	if req.RosterID == "" {
		return nil, errors.New("roster harus dipilih")
	}
	if req.EmployeeID == "" {
		return nil, errors.New("karyawan harus dipilih")
	}
	if req.Date == "" {
		return nil, errors.New("tanggal harus diisi")
	}
	if req.ShiftID == "" {
		return nil, errors.New("shift harus dipilih")
	}

	entry, err := repository.CreateRosterEntry(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat entri roster: %w", err)
	}
	return entry, nil
}

func (s *RosterService) BulkCreateEntries(ctx context.Context, req *models.BulkRosterEntryRequest, userID string) error {
	if req.RosterID == "" {
		return errors.New("roster harus dipilih")
	}
	if len(req.Entries) == 0 {
		return errors.New("minimal 1 entri harus diisi")
	}

	err := repository.BulkCreateRosterEntries(ctx, req.Entries, req.ClearExisting, userID)
	if err != nil {
		return fmt.Errorf("gagal menyimpan entri roster: %w", err)
	}
	return nil
}

func (s *RosterService) DeleteEntry(ctx context.Context, id, userID string) error {
	return repository.DeleteRosterEntry(ctx, id, userID)
}

func (s *RosterService) DeleteEmployeeEntries(ctx context.Context, rosterID, employeeID, userID string) error {
	if rosterID == "" {
		return errors.New("roster harus dipilih")
	}
	if employeeID == "" {
		return errors.New("karyawan harus dipilih")
	}
	return repository.DeleteEmployeeRosterEntries(ctx, rosterID, employeeID, userID)
}

