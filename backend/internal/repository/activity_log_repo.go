package repository

import (
	"context"
	"fmt"
	"strings"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

// ActivityLogRepo handles database operations for activity logs
type ActivityLogRepo struct{}

// NewActivityLogRepo creates a new ActivityLogRepo
func NewActivityLogRepo() *ActivityLogRepo {
	return &ActivityLogRepo{}
}

// ListActivityLogs returns paginated activity logs with filters
func (r *ActivityLogRepo) ListActivityLogs(ctx context.Context, filter *models.ActivityLogFilter) (*models.ActivityLogListResponse, error) {
	where := []string{"1=1"} // activity_logs is append-only, no soft delete
	args := make([]any, 0)
	argIdx := 1

	if filter.Action != "" {
		where = append(where, fmt.Sprintf("al.action = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.EntityType != "" {
		where = append(where, fmt.Sprintf("al.entity_type = $%d", argIdx))
		args = append(args, filter.EntityType)
		argIdx++
	}
	if filter.UserID != "" {
		where = append(where, fmt.Sprintf("al.user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.StartDate != "" {
		where = append(where, fmt.Sprintf("al.created_at >= $%d::timestamptz", argIdx))
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != "" {
		where = append(where, fmt.Sprintf("al.created_at <= $%d::timestamptz", argIdx))
		args = append(args, filter.EndDate)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM activity_logs al WHERE %s`, whereClause)
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count activity logs: %w", err)
	}

	// Get paginated logs with employee name
	offset := (filter.Page - 1) * filter.PerPage
	selectQuery := fmt.Sprintf(`
		SELECT al.id, al.user_id, al.action, al.entity_type, 
		       COALESCE(al.entity_name, ''), al.created_at,
		       COALESCE(e.full_name, '') as employee_name
		FROM activity_logs al
		LEFT JOIN employees e ON e.id = al.user_id
		WHERE %s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, filter.PerPage, offset)

	rows, err := database.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query activity logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ActivityLogSummary, 0)
	for rows.Next() {
		var l models.ActivityLogSummary
		var empName string
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityName, &l.CreatedAt, &empName); err != nil {
			return nil, fmt.Errorf("scan activity log: %w", err)
		}
		if empName != "" {
			l.EmployeeName = &empName
		} else {
			l.EmployeeName = nil
		}
		logs = append(logs, l)
	}

	return &models.ActivityLogListResponse{
		Logs:     logs,
		Total:    total,
		Page:     filter.Page,
		PerPage:  filter.PerPage,
	}, nil
}

// GetActivityLog returns a single activity log by ID
func (r *ActivityLogRepo) GetActivityLog(ctx context.Context, id string) (*models.ActivityLog, error) {
	var l models.ActivityLog
	var empName, empEmail string
	err := database.Pool.QueryRow(ctx, `
		SELECT al.id, al.user_id, al.action, al.entity_type, al.entity_id,
		       COALESCE(al.entity_name, ''), al.old_values, al.new_values,
		       COALESCE(al.ip_address, ''), COALESCE(al.user_agent, ''),
		       al.created_at,
		       COALESCE(e.full_name, ''), COALESCE(e.email, '')
		FROM activity_logs al
		LEFT JOIN employees e ON e.id = al.user_id
		WHERE al.id = $1
	`, id).Scan(
		&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID,
		&l.EntityName, &l.OldValues, &l.NewValues,
		&l.IPAddress, &l.UserAgent,
		&l.CreatedAt,
		&empName, &empEmail,
	)
	if err != nil {
		return nil, fmt.Errorf("get activity log: %w", err)
	}

	if empName != "" {
		l.EmployeeName = &empName
	}
	if empEmail != "" {
		l.EmployeeEmail = &empEmail
	}

	return &l, nil
}

// GetDistinctEntityTypes returns distinct entity types for filter dropdown
func (r *ActivityLogRepo) GetDistinctEntityTypes(ctx context.Context) ([]string, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT DISTINCT entity_type FROM activity_logs 
		ORDER BY entity_type
	`)
	if err != nil {
		return nil, fmt.Errorf("get entity types: %w", err)
	}
	defer rows.Close()

	types := make([]string, 0)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan entity type: %w", err)
		}
		types = append(types, t)
	}
	return types, nil
}

// GetDistinctActions returns distinct action types for filter dropdown
func (r *ActivityLogRepo) GetDistinctActions(ctx context.Context) ([]string, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT DISTINCT action FROM activity_logs 
		ORDER BY action
	`)
	if err != nil {
		return nil, fmt.Errorf("get actions: %w", err)
	}
	defer rows.Close()

	actions := make([]string, 0)
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		actions = append(actions, a)
	}
	return actions, nil
}
