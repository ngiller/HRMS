package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type HolidayService struct{}

func NewHolidayService() *HolidayService {
	return &HolidayService{}
}

func (s *HolidayService) List(ctx context.Context, page, perPage int, year int, holidayType string) (*models.HolidayListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	if year <= 0 {
		year = time.Now().Year()
	}

	holidays, total, err := repository.ListHolidays(ctx, page, perPage, year, holidayType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data hari libur: %w", err)
	}
	if holidays == nil {
		holidays = []models.CompanyHoliday{}
	}

	return &models.HolidayListResponse{
		Holidays: holidays,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}, nil
}

func (s *HolidayService) Get(ctx context.Context, id string) (*models.CompanyHoliday, error) {
	h, err := repository.GetHolidayByID(ctx, id)
	if err != nil {
		return nil, errors.New("hari libur tidak ditemukan")
	}
	if h == nil {
		return nil, errors.New("hari libur tidak ditemukan")
	}
	return h, nil
}

func (s *HolidayService) Create(ctx context.Context, createdBy string, req *models.CreateHolidayReq) (*models.CompanyHoliday, error) {
	if req.Name == "" {
		return nil, errors.New("nama hari libur harus diisi")
	}
	if req.Date == "" {
		return nil, errors.New("tanggal hari libur harus diisi")
	}
	if req.HolidayType == "" {
		req.HolidayType = "national"
	}

	h, err := repository.CreateHoliday(ctx, createdBy, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat hari libur: %w", err)
	}
	return h, nil
}

func (s *HolidayService) Update(ctx context.Context, id, userID string, req *models.UpdateHolidayReq) (*models.CompanyHoliday, error) {
	existing, err := repository.GetHolidayByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("hari libur tidak ditemukan")
	}

	h, err := repository.UpdateHoliday(ctx, id, userID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate hari libur: %w", err)
	}
	return h, nil
}

func (s *HolidayService) Delete(ctx context.Context, id, userID string) error {
	return repository.DeleteHoliday(ctx, id, userID)
}

// GetByYear returns all holidays for a given year
func (s *HolidayService) GetByYear(ctx context.Context, year int) (*models.HolidayYearResponse, error) {
	if year <= 0 {
		year = time.Now().Year()
	}

	holidays, err := repository.GetHolidaysByYear(ctx, year)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat hari libur tahun %d: %w", year, err)
	}
	if holidays == nil {
		holidays = []models.CompanyHoliday{}
	}

	return &models.HolidayYearResponse{
		Year:     year,
		Holidays: holidays,
	}, nil
}
