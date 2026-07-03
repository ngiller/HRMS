package service

import (
	"context"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

// ActivityLogService handles business logic for activity logs
type ActivityLogService struct {
	repo *repository.ActivityLogRepo
}

// NewActivityLogService creates a new ActivityLogService
func NewActivityLogService() *ActivityLogService {
	return &ActivityLogService{
		repo: repository.NewActivityLogRepo(),
	}
}

// ListActivityLogs returns paginated activity logs with filters
func (s *ActivityLogService) ListActivityLogs(ctx context.Context, filter *models.ActivityLogFilter) (*models.ActivityLogListResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 25
	}
	return s.repo.ListActivityLogs(ctx, filter)
}

// GetActivityLog returns a single activity log by ID
func (s *ActivityLogService) GetActivityLog(ctx context.Context, id string) (*models.ActivityLog, error) {
	return s.repo.GetActivityLog(ctx, id)
}

// GetDistinctEntityTypes returns distinct entity types for filter dropdown
func (s *ActivityLogService) GetDistinctEntityTypes(ctx context.Context) ([]string, error) {
	return s.repo.GetDistinctEntityTypes(ctx)
}

// GetDistinctActions returns distinct action types for filter dropdown
func (s *ActivityLogService) GetDistinctActions(ctx context.Context) ([]string, error) {
	return s.repo.GetDistinctActions(ctx)
}
