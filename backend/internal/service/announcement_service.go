package service

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type AnnouncementService struct{}

func NewAnnouncementService() *AnnouncementService {
	return &AnnouncementService{}
}

func (s *AnnouncementService) List(ctx context.Context, page, perPage int, announcementType string) (*models.AnnouncementListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	announcements, total, err := repository.ListAnnouncements(ctx, page, perPage, announcementType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data pengumuman: %w", err)
	}
	if announcements == nil {
		announcements = []models.AnnouncementSummary{}
	}

	return &models.AnnouncementListResponse{
		Announcements: announcements,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
	}, nil
}

func (s *AnnouncementService) Get(ctx context.Context, id string) (*models.Announcement, error) {
	a, err := repository.GetAnnouncementByID(ctx, id)
	if err != nil {
		return nil, errors.New("pengumuman tidak ditemukan")
	}
	if a == nil {
		return nil, errors.New("pengumuman tidak ditemukan")
	}
	return a, nil
}

func (s *AnnouncementService) Create(ctx context.Context, createdBy string, req *models.CreateAnnouncementReq) (*models.Announcement, error) {
	if req.Title == "" {
		return nil, errors.New("judul pengumuman harus diisi")
	}
	if req.Content == "" {
		return nil, errors.New("konten pengumuman harus diisi")
	}
	if req.AnnouncementType == "" {
		req.AnnouncementType = "general"
	}

	a, err := repository.CreateAnnouncement(ctx, createdBy, req)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengumuman: %w", err)
	}
	return a, nil
}

func (s *AnnouncementService) Update(ctx context.Context, id, userID string, req *models.UpdateAnnouncementReq) (*models.Announcement, error) {
	existing, err := repository.GetAnnouncementByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("pengumuman tidak ditemukan")
	}

	a, err := repository.UpdateAnnouncement(ctx, id, userID, req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate pengumuman: %w", err)
	}
	return a, nil
}

func (s *AnnouncementService) Delete(ctx context.Context, id, userID string) error {
	return repository.DeleteAnnouncement(ctx, id, userID)
}

func (s *AnnouncementService) MarkAsRead(ctx context.Context, announcementID, employeeID string) error {
	return repository.MarkAnnouncementRead(ctx, announcementID, employeeID)
}
