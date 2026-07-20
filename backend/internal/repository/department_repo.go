package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// ListDepartments returns paginated department list.
func ListDepartments(ctx context.Context, page, perPage int, search string) ([]models.DepartmentSummary, int, error) {
	// Count total
	countQuery := `
		SELECT COUNT(*) FROM departments d
		WHERE d.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND (LOWER(d.name) LIKE LOWER($%d) OR LOWER(d.code) LIKE LOWER($%d))", argIdx, argIdx)
		args = append(args, "%"+search+"%")
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch page
	offset := (page - 1) * perPage
	argIdx++
	query := fmt.Sprintf(`
		SELECT d.id, d.name, d.code,
			COALESCE(dp.name, '') as parent_name,
			COALESCE(e.full_name, '') as head_name,
			COALESCE(ws.name, '') as work_schedule_name,
			COALESCE(d.description, '') as description,
			d.is_active, d.sort_order,
			(SELECT COUNT(*) FROM employees emp WHERE emp.department_id = d.id AND emp.deleted_at IS NULL) as employee_count,
			d.created_at
		FROM departments d
		LEFT JOIN departments dp ON d.parent_id = dp.id
		LEFT JOIN employees e ON d.head_id = e.id
		LEFT JOIN work_schedules ws ON d.work_schedule_id = ws.id
		WHERE d.deleted_at IS NULL
	`)

	if search != "" {
		query += fmt.Sprintf(" AND (LOWER(d.name) LIKE LOWER($%d) OR LOWER(d.code) LIKE LOWER($%d))", argIdx-1, argIdx-1)
	}

	query += fmt.Sprintf(" ORDER BY d.sort_order ASC, d.name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var departments []models.DepartmentSummary
	for rows.Next() {
		var dept models.DepartmentSummary
		err := rows.Scan(
			&dept.ID, &dept.Name, &dept.Code,
			&dept.ParentName, &dept.HeadName,
			&dept.WorkScheduleName,
			&dept.Description, &dept.IsActive, &dept.SortOrder,
			&dept.EmployeeCnt, &dept.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		departments = append(departments, dept)
	}

	return departments, total, nil
}

// GetAllDepartments returns all active departments (for dropdown, no pagination).
func GetAllDepartments(ctx context.Context) ([]models.DepartmentSummary, error) {
	query := `
		SELECT d.id, d.name, d.code,
			COALESCE(dp.name, '') as parent_name,
			COALESCE(ws.name, '') as work_schedule_name
		FROM departments d
		LEFT JOIN departments dp ON d.parent_id = dp.id
		LEFT JOIN work_schedules ws ON d.work_schedule_id = ws.id
		WHERE d.deleted_at IS NULL AND d.is_active = TRUE
		ORDER BY d.sort_order ASC, d.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []models.DepartmentSummary
	for rows.Next() {
		var dept models.DepartmentSummary
		if err := rows.Scan(&dept.ID, &dept.Name, &dept.Code, &dept.ParentName, &dept.WorkScheduleName); err != nil {
			return nil, err
		}
		departments = append(departments, dept)
	}
	return departments, nil
}

// GetDepartmentByID returns a single department by ID.
func GetDepartmentByID(ctx context.Context, id string) (*models.Department, error) {
	query := `
		SELECT d.id, d.name, d.code, d.parent_id, d.head_id, d.work_schedule_id,
			COALESCE(d.description, '') as description,
			d.is_active, d.sort_order,
			d.created_at, d.updated_at, d.deleted_at
		FROM departments d
		WHERE (d.id::text = $1) AND d.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)

	var dept models.Department
	err := row.Scan(
		&dept.ID, &dept.Name, &dept.Code, &dept.ParentID, &dept.HeadID, &dept.WorkScheduleID,
		&dept.Description, &dept.IsActive, &dept.SortOrder,
		&dept.CreatedAt, &dept.UpdatedAt, &dept.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// CreateDepartment inserts a new department.
func CreateDepartment(ctx context.Context, req *models.CreateDepartmentRequest, userID string) (*models.Department, error) {
	query := `
		INSERT INTO departments (name, code, parent_id, head_id, work_schedule_id, description, sort_order)
		VALUES ($1, $2,
			NULLIF(NULLIF($3, ''), 'null')::uuid,
			NULLIF(NULLIF($4, ''), 'null')::uuid,
			NULLIF(NULLIF($5, ''), 'null')::uuid,
			$6, $7)
		RETURNING id, name, code, parent_id, head_id, work_schedule_id,
			COALESCE(description, '') as description,
			is_active, sort_order, created_at, updated_at, deleted_at
	`
	var dept *models.Department
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.Name,
			req.Code,
			coalesceStr(req.ParentID),
			coalesceStr(req.HeadID),
			coalesceStr(req.WorkScheduleID),
			req.Description,
			req.SortOrder,
		)

		var result models.Department
		if err := row.Scan(
			&result.ID, &result.Name, &result.Code, &result.ParentID, &result.HeadID, &result.WorkScheduleID,
			&result.Description, &result.IsActive, &result.SortOrder,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		dept = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dept, nil
}

// UpdateDepartment updates an existing department.
func UpdateDepartment(ctx context.Context, id string, req *models.UpdateDepartmentRequest, userID string) (*models.Department, error) {
	// Build dynamic SET clause
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.Code != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("code = $%d", argIdx))
		args = append(args, *req.Code)
	}
	if req.ParentID != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("parent_id = NULLIF(NULLIF($%d, ''), 'null')::uuid", argIdx))
		args = append(args, *req.ParentID)
	}
	if req.HeadID != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("head_id = NULLIF(NULLIF($%d, ''), 'null')::uuid", argIdx))
		args = append(args, *req.HeadID)
	}
	if req.Description != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
	}
	if req.IsActive != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
	}
	if req.SortOrder != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("sort_order = $%d", argIdx))
		args = append(args, *req.SortOrder)
	}
	if req.WorkScheduleID != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("work_schedule_id = NULLIF(NULLIF($%d, ''), 'null')::uuid", argIdx))
		args = append(args, *req.WorkScheduleID)
	}

	if len(setClauses) == 0 {
		return GetDepartmentByID(ctx, id)
	}

	var dept *models.Department
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE departments SET %s
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, name, code, parent_id, head_id, work_schedule_id,
				COALESCE(description, '') as description,
				is_active, sort_order, created_at, updated_at, deleted_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)

		var result models.Department
		if err := row.Scan(
			&result.ID, &result.Name, &result.Code, &result.ParentID, &result.HeadID, &result.WorkScheduleID,
			&result.Description, &result.IsActive, &result.SortOrder,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		dept = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dept, nil
}

// UpdateDepartmentWorkSchedule updates only the work_schedule_id for a department.
func UpdateDepartmentWorkSchedule(ctx context.Context, id, workScheduleID, userID string) (*models.Department, error) {
	query := `
		UPDATE departments SET work_schedule_id = NULLIF(NULLIF($1, ''), 'null')::uuid
		WHERE id::text = $2 AND deleted_at IS NULL
		RETURNING id, name, code, parent_id, head_id, work_schedule_id,
			COALESCE(description, '') as description,
			is_active, sort_order, created_at, updated_at, deleted_at
	`
	var dept *models.Department
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, workScheduleID, id)

		var result models.Department
		if err := row.Scan(
			&result.ID, &result.Name, &result.Code, &result.ParentID, &result.HeadID, &result.WorkScheduleID,
			&result.Description, &result.IsActive, &result.SortOrder,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		dept = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dept, nil
}

// DeleteDepartment soft-deletes a department.
func DeleteDepartment(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE departments SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

// GetDepartmentByName looks up a department ID by name (case-insensitive).
func GetDepartmentByName(ctx context.Context, name string) (*string, error) {
	var id string
	err := database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE LOWER(name) = LOWER($1) AND deleted_at IS NULL LIMIT 1`, name).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

// CheckDepartmentCodeExists checks if a department code already exists (excluding a given ID).
func CheckDepartmentCodeExists(ctx context.Context, code string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM departments WHERE code = $1 AND deleted_at IS NULL`
	args := []interface{}{code}

	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}

	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

// CheckDepartmentNameExists checks if a department name already exists (excluding a given ID).
func CheckDepartmentNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM departments WHERE name = $1 AND deleted_at IS NULL`
	args := []interface{}{name}

	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}

	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

// CheckDepartmentHasChildren checks if a department has child departments.
func CheckDepartmentHasChildren(ctx context.Context, id string) (bool, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM departments WHERE parent_id::text = $1 AND deleted_at IS NULL`, id,
	).Scan(&count)
	return count > 0, err
}

// CheckDepartmentHasEmployees checks if a department has employees.
func CheckDepartmentHasEmployees(ctx context.Context, id string) (bool, error) {
	// Department ID is stored on the position, but employees have department_id directly
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM employees WHERE department_id::text = $1 AND deleted_at IS NULL`, id,
	).Scan(&count)
	return count > 0, err
}

// --- work_schedules ---

// GetAllWorkSchedules returns all active work schedules (for dropdown).
func GetAllWorkSchedules(ctx context.Context) ([]models.WorkScheduleSummary, error) {
	query := `
		SELECT ws.id, ws.name, ws.schedule_type::text,
			COALESCE(ws.description, '') as description,
			ws.weekly_hours, ws.is_active, ws.created_at
		FROM work_schedules ws
		WHERE ws.deleted_at IS NULL AND ws.is_active = TRUE
		ORDER BY ws.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.WorkScheduleSummary
	for rows.Next() {
		var ws models.WorkScheduleSummary
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.ScheduleType, &ws.Description, &ws.WeeklyHours, &ws.IsActive, &ws.CreatedAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, ws)
	}
	return schedules, nil
}

// --- helpers ---

func coalesceStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func joinStrings(elems []string, sep string) string {
	result := ""
	for i, e := range elems {
		if i > 0 {
			result += sep
		}
		result += e
	}
	return result
}
