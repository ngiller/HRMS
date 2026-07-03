package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type AttendanceLocationService struct{}

func NewAttendanceLocationService() *AttendanceLocationService {
	return &AttendanceLocationService{}
}

func (s *AttendanceLocationService) List(ctx context.Context, page, perPage int, search string) (*models.AttendanceLocationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	locations, total, err := repository.ListAttendanceLocations(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data lokasi absensi: %w", err)
	}
	if locations == nil {
		locations = []models.AttendanceLocationSummary{}
	}

	return &models.AttendanceLocationListResponse{
		AttendanceLocations: locations,
		Total:               total,
		Page:                page,
		PerPage:             perPage,
	}, nil
}

func (s *AttendanceLocationService) GetAll(ctx context.Context) ([]models.AttendanceLocationSummary, error) {
	locations, err := repository.GetAllAttendanceLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar lokasi absensi: %w", err)
	}
	if locations == nil {
		locations = []models.AttendanceLocationSummary{}
	}
	return locations, nil
}

func (s *AttendanceLocationService) Get(ctx context.Context, id string) (*models.AttendanceLocation, error) {
	loc, err := repository.GetAttendanceLocationByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data lokasi absensi")
	}
	if loc == nil {
		return nil, errors.New("lokasi absensi tidak ditemukan")
	}
	return loc, nil
}

func (s *AttendanceLocationService) Create(ctx context.Context, req *models.CreateAttendanceLocationRequest, userID string) (*models.AttendanceLocation, error) {
	if req.Name == "" {
		return nil, errors.New("nama lokasi absensi harus diisi")
	}
	if req.Latitude == 0 && req.Longitude == 0 {
		return nil, errors.New("koordinat latitude dan longitude harus diisi")
	}
	if req.RadiusMeters == 0 {
		req.RadiusMeters = 100
	}

	exists, err := repository.CheckAttendanceLocationNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi lokasi absensi: %w", err)
	}
	if exists {
		return nil, errors.New("nama lokasi absensi sudah digunakan")
	}

	loc, err := repository.CreateAttendanceLocation(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat lokasi absensi: %w", err)
	}
	return loc, nil
}

func (s *AttendanceLocationService) Update(ctx context.Context, id string, req *models.UpdateAttendanceLocationRequest, userID string) (*models.AttendanceLocation, error) {
	existing, err := repository.GetAttendanceLocationByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data lokasi absensi")
	}
	if existing == nil {
		return nil, errors.New("lokasi absensi tidak ditemukan")
	}

	if req.Name != nil && *req.Name != existing.Name {
		exists, err := repository.CheckAttendanceLocationNameExists(ctx, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("gagal validasi lokasi absensi: %w", err)
		}
		if exists {
			return nil, errors.New("nama lokasi absensi sudah digunakan")
		}
	}

	loc, err := repository.UpdateAttendanceLocation(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui lokasi absensi: %w", err)
	}
	return loc, nil
}

func (s *AttendanceLocationService) Delete(ctx context.Context, id, userID string) error {
	existing, err := repository.GetAttendanceLocationByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data lokasi absensi")
	}
	if existing == nil {
		return errors.New("lokasi absensi tidak ditemukan")
	}

	err = repository.DeleteAttendanceLocation(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus lokasi absensi: %w", err)
	}
	return nil
}
