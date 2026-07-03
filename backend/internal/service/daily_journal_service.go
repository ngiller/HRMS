package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type DailyJournalService struct{}

func NewDailyJournalService() *DailyJournalService {
	return &DailyJournalService{}
}

func (s *DailyJournalService) List(ctx context.Context, page, perPage int, departmentID, employeeID, dateFrom, dateTo string) (*models.DailyJournalListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	journals, total, err := repository.ListDailyJournals(ctx, page, perPage, departmentID, employeeID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat jurnal harian: %w", err)
	}
	if journals == nil {
		journals = []models.DailyJournalSummary{}
	}

	return &models.DailyJournalListResponse{
		Journals: journals,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}, nil
}

func (s *DailyJournalService) Get(ctx context.Context, id string) (*models.DailyJournal, error) {
	j, err := repository.GetDailyJournalByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat jurnal harian")
	}
	if j == nil {
		return nil, errors.New("jurnal harian tidak ditemukan")
	}
	return j, nil
}

func (s *DailyJournalService) Create(ctx context.Context, employeeID string, req *models.CreateDailyJournalRequest) (*models.DailyJournal, error) {
	if req.WorkDescription == "" {
		return nil, errors.New("deskripsi pekerjaan harus diisi")
	}
	if req.JournalDate == "" {
		req.JournalDate = time.Now().Format("2006-01-02")
	}

	journal, err := repository.CreateDailyJournal(ctx, employeeID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat jurnal harian: %w", err)
	}
	return journal, nil
}

func (s *DailyJournalService) Acknowledge(ctx context.Context, id, managerID string) (*models.DailyJournal, error) {
	j, err := repository.AcknowledgeJournal(ctx, id, managerID)
	if err != nil {
		return nil, err
	}
	return j, nil
}
