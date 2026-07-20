package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListShifts(ctx context.Context, page, perPage int, search string, departmentID string) ([]models.ShiftSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM shifts WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND LOWER(name) LIKE LOWER($%d)", argIdx)
		args = append(args, "%"+search+"%")
	}
	if departmentID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND (department_id::text = $%d OR department_id IS NULL)", argIdx)
		args = append(args, departmentID)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT id, department_id, name, code, start_time::text, end_time::text,
			color, is_active, COALESCE(description,''), created_at
		FROM shifts WHERE deleted_at IS NULL
	`
	queryArgs := []interface{}{}
	qArgIdx := 0
	if search != "" {
		qArgIdx++
		query += fmt.Sprintf(" AND LOWER(name) LIKE LOWER($%d)", qArgIdx)
		queryArgs = append(queryArgs, "%"+search+"%")
	}
	if departmentID != "" {
		qArgIdx++
		query += fmt.Sprintf(" AND (department_id::text = $%d OR department_id IS NULL)", qArgIdx)
		queryArgs = append(queryArgs, departmentID)
	}

	qArgIdx++
	query += fmt.Sprintf(" ORDER BY start_time ASC LIMIT $%d", qArgIdx)
	queryArgs = append(queryArgs, perPage)

	qArgIdx++
	query += fmt.Sprintf(" OFFSET $%d", qArgIdx)
	queryArgs = append(queryArgs, offset)

	rows, err := database.Pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var shifts []models.ShiftSummary
	for rows.Next() {
		var s models.ShiftSummary
		if err := rows.Scan(&s.ID, &s.DepartmentID, &s.Name, &s.Code, &s.StartTime, &s.EndTime, &s.Color, &s.IsActive, &s.Description, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		shifts = append(shifts, s)
	}
	return shifts, total, nil
}

func GetAllShifts(ctx context.Context, departmentID string) ([]models.ShiftSummary, error) {
	query := `
		SELECT id, department_id, name, code, start_time::text, end_time::text,
			color, is_active, COALESCE(description,''), created_at
		FROM shifts WHERE deleted_at IS NULL AND is_active = TRUE
	`
	args := []interface{}{}
	if departmentID != "" {
		query += ` AND (department_id::text = $1 OR department_id IS NULL)`
		args = append(args, departmentID)
	}
	query += ` ORDER BY start_time ASC`
	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.ShiftSummary
	for rows.Next() {
		var s models.ShiftSummary
		if err := rows.Scan(&s.ID, &s.DepartmentID, &s.Name, &s.Code, &s.StartTime, &s.EndTime, &s.Color, &s.IsActive, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, nil
}

func GetShiftByID(ctx context.Context, id string) (*models.Shift, error) {
	query := `
		SELECT id, department_id, name, code, start_time::text, end_time::text,
			break_start::text, break_end::text, color, COALESCE(description,''),
			is_active, created_at, updated_at, deleted_at
		FROM shifts WHERE id::text = $1 AND deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var s models.Shift
	err := row.Scan(&s.ID, &s.DepartmentID, &s.Name, &s.Code, &s.StartTime, &s.EndTime,
		&s.BreakStart, &s.BreakEnd, &s.Color, &s.Description,
		&s.IsActive, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func CreateShift(ctx context.Context, req *models.CreateShiftRequest, userID string) (*models.Shift, error) {
	var s *models.Shift
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
			VALUES (NULLIF($1, '')::uuid, $2, $3, $4::time, $5::time, NULLIF($6,'')::time, NULLIF($7,'')::time, $8, $9)
			RETURNING id, department_id, name, code, start_time::text, end_time::text,
				break_start::text, break_end::text, color, COALESCE(description,''),
				is_active, created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, query, req.DepartmentID, req.Name, req.Code, req.StartTime, req.EndTime,
			req.BreakStart, req.BreakEnd, req.Color, req.Description)
		var result models.Shift
		if err := row.Scan(&result.ID, &result.DepartmentID, &result.Name, &result.Code, &result.StartTime, &result.EndTime,
			&result.BreakStart, &result.BreakEnd, &result.Color, &result.Description,
			&result.IsActive, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		s = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func UpdateShift(ctx context.Context, id string, req *models.UpdateShiftRequest, userID string) (*models.Shift, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.DepartmentID != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("department_id = NULLIF($%d,'')::uuid", argIdx)); args = append(args, *req.DepartmentID)
	}
	if req.Name != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx)); args = append(args, *req.Name)
	}
	if req.Code != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("code = $%d", argIdx)); args = append(args, *req.Code)
	}
	if req.StartTime != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("start_time = $%d::time", argIdx)); args = append(args, *req.StartTime)
	}
	if req.EndTime != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("end_time = $%d::time", argIdx)); args = append(args, *req.EndTime)
	}
	if req.BreakStart != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("break_start = NULLIF($%d,'')::time", argIdx)); args = append(args, *req.BreakStart)
	}
	if req.BreakEnd != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("break_end = NULLIF($%d,'')::time", argIdx)); args = append(args, *req.BreakEnd)
	}
	if req.Color != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("color = $%d", argIdx)); args = append(args, *req.Color)
	}
	if req.Description != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx)); args = append(args, *req.Description)
	}
	if req.IsActive != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx)); args = append(args, *req.IsActive)
	}

	if len(setClauses) == 0 {
		return GetShiftByID(ctx, id)
	}

	var s *models.Shift
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE shifts SET %s
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, department_id, name, code, start_time::text, end_time::text,
				break_start::text, break_end::text, color, COALESCE(description,''),
				is_active, created_at, updated_at, deleted_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.Shift
		if err := row.Scan(&result.ID, &result.DepartmentID, &result.Name, &result.Code, &result.StartTime, &result.EndTime,
			&result.BreakStart, &result.BreakEnd, &result.Color, &result.Description,
			&result.IsActive, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		s = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func DeleteShift(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE shifts SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

func CheckShiftCodeExists(ctx context.Context, code string, departmentID string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM shifts WHERE code = $1 AND deleted_at IS NULL`
	args := []interface{}{code}
	argIdx := 1
	if departmentID != "" {
		argIdx++
		query += fmt.Sprintf(" AND department_id::text = $%d", argIdx)
		args = append(args, departmentID)
	} else {
		query += " AND department_id IS NULL"
	}
	if excludeID != "" {
		argIdx++
		query += fmt.Sprintf(" AND id::text != $%d", argIdx)
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func CheckShiftNameExists(ctx context.Context, name string, departmentID string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM shifts WHERE name = $1 AND deleted_at IS NULL`
	args := []interface{}{name}
	argIdx := 1
	if departmentID != "" {
		argIdx++
		query += fmt.Sprintf(" AND department_id::text = $%d", argIdx)
		args = append(args, departmentID)
	} else {
		query += " AND department_id IS NULL"
	}
	if excludeID != "" {
		argIdx++
		query += fmt.Sprintf(" AND id::text != $%d", argIdx)
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}
