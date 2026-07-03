package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListPositions(ctx context.Context, page, perPage int, search string) ([]models.PositionSummary, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM positions p
		WHERE p.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND (LOWER(p.name) LIKE LOWER($%d) OR LOWER(d.name) LIKE LOWER($%d))", argIdx, argIdx)
		// Need to join for search on department name — simplified: search position name only
		args = append(args, "%"+search+"%")
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	argIdx++
	query := fmt.Sprintf(`
		SELECT p.id, p.name,
			COALESCE(d.name, '') as department_name,
			p.department_id::text,
			COALESCE(pg.name, '') as grade_name,
			COALESCE(p.grade_id::text, '') as grade_id,
			COALESCE(p.description, '') as description,
			p.is_active,
			p.created_at
		FROM positions p
		LEFT JOIN departments d ON p.department_id = d.id
		LEFT JOIN position_grades pg ON p.grade_id = pg.id
		WHERE p.deleted_at IS NULL
	`)
	if search != "" {
		query += fmt.Sprintf(" AND (LOWER(p.name) LIKE LOWER($%d))", argIdx-1)
	}
	query += fmt.Sprintf(" ORDER BY p.name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var positions []models.PositionSummary
	for rows.Next() {
		var pos models.PositionSummary
		if err := rows.Scan(&pos.ID, &pos.Name, &pos.DepartmentName, &pos.DepartmentID,
			&pos.GradeName, &pos.GradeID, &pos.Description, &pos.IsActive, &pos.CreatedAt); err != nil {
			return nil, 0, err
		}
		positions = append(positions, pos)
	}
	return positions, total, nil
}

func GetAllPositions(ctx context.Context) ([]models.PositionSummary, error) {
	query := `
		SELECT p.id, p.name,
			COALESCE(d.name, '') as department_name,
			p.department_id::text,
			COALESCE(pg.name, '') as grade_name,
			COALESCE(p.grade_id::text, '') as grade_id,
			COALESCE(p.description, '') as description,
			p.is_active,
			p.created_at
		FROM positions p
		LEFT JOIN departments d ON p.department_id = d.id
		LEFT JOIN position_grades pg ON p.grade_id = pg.id
		WHERE p.deleted_at IS NULL AND p.is_active = TRUE
		ORDER BY p.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []models.PositionSummary
	for rows.Next() {
		var pos models.PositionSummary
		if err := rows.Scan(&pos.ID, &pos.Name, &pos.DepartmentName, &pos.DepartmentID,
			&pos.GradeName, &pos.GradeID, &pos.Description, &pos.IsActive, &pos.CreatedAt); err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}
	return positions, nil
}

func GetPositionByID(ctx context.Context, id string) (*models.Position, error) {
	query := `
		SELECT p.id, p.name, p.department_id, p.grade_id,
			COALESCE(p.description, '') as description,
			p.is_active, p.created_at, p.updated_at, p.deleted_at
		FROM positions p
		WHERE (p.id::text = $1) AND p.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var pos models.Position
	err := row.Scan(&pos.ID, &pos.Name, &pos.DepartmentID, &pos.GradeID,
		&pos.Description, &pos.IsActive, &pos.CreatedAt, &pos.UpdatedAt, &pos.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func CreatePosition(ctx context.Context, req *models.CreatePositionRequest, userID string) (*models.Position, error) {
	query := `
		INSERT INTO positions (name, department_id, grade_id, description)
		VALUES ($1, $2::uuid,
			NULLIF(NULLIF($3, ''), 'null')::uuid,
			$4)
		RETURNING id, name, department_id, grade_id,
			COALESCE(description, '') as description,
			is_active, created_at, updated_at, deleted_at
	`
	var pos *models.Position
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.Name, req.DepartmentID, coalesceStr(&req.GradeID), req.Description,
		)
		var result models.Position
		if err := row.Scan(&result.ID, &result.Name, &result.DepartmentID, &result.GradeID,
			&result.Description, &result.IsActive, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		pos = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pos, nil
}

func UpdatePosition(ctx context.Context, id string, req *models.UpdatePositionRequest, userID string) (*models.Position, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.DepartmentID != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("department_id = $%d::uuid", argIdx))
		args = append(args, *req.DepartmentID)
	}
	if req.GradeID != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("grade_id = NULLIF(NULLIF($%d, ''), 'null')::uuid", argIdx))
		args = append(args, *req.GradeID)
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

	if len(setClauses) == 0 {
		return GetPositionByID(ctx, id)
	}

	var pos *models.Position
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE positions SET %s
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, name, department_id, grade_id,
				COALESCE(description, '') as description,
				is_active, created_at, updated_at, deleted_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.Position
		if err := row.Scan(&result.ID, &result.Name, &result.DepartmentID, &result.GradeID,
			&result.Description, &result.IsActive, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		pos = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pos, nil
}

func DeletePosition(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE positions SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

func CheckPositionNameExists(ctx context.Context, name string, departmentID string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM positions WHERE name = $1 AND department_id::text = $2 AND deleted_at IS NULL`
	args := []interface{}{name, departmentID}
	if excludeID != "" {
		query += ` AND id::text != $3`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func CheckPositionHasEmployees(ctx context.Context, id string) (bool, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM employees WHERE position_id::text = $1 AND deleted_at IS NULL`, id,
	).Scan(&count)
	return count > 0, err
}
