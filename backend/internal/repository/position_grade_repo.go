package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListPositionGrades(ctx context.Context, page, perPage int, search string) ([]models.PositionGradeSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM position_grades`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" WHERE LOWER(name) LIKE LOWER($%d)", argIdx)
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
		SELECT id, name, level, min_salary, max_salary, COALESCE(description, '') as description, is_active, created_at
		FROM position_grades
	`)
	if search != "" {
		query += fmt.Sprintf(" WHERE LOWER(name) LIKE LOWER($%d)", argIdx-1)
	}
	query += fmt.Sprintf(" ORDER BY level ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var grades []models.PositionGradeSummary
	for rows.Next() {
		var g models.PositionGradeSummary
		if err := rows.Scan(&g.ID, &g.Name, &g.Level, &g.MinSalary, &g.MaxSalary, &g.Description, &g.IsActive, &g.CreatedAt); err != nil {
			return nil, 0, err
		}
		grades = append(grades, g)
	}
	return grades, total, nil
}

func GetAllPositionGrades(ctx context.Context) ([]models.PositionGradeSummary, error) {
	query := `SELECT id, name, level, min_salary, max_salary, COALESCE(description, '') as description, is_active, created_at FROM position_grades WHERE is_active = TRUE ORDER BY level ASC`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []models.PositionGradeSummary
	for rows.Next() {
		var g models.PositionGradeSummary
		if err := rows.Scan(&g.ID, &g.Name, &g.Level, &g.MinSalary, &g.MaxSalary, &g.Description, &g.IsActive, &g.CreatedAt); err != nil {
			return nil, err
		}
		grades = append(grades, g)
	}
	return grades, nil
}

func GetPositionGradeByID(ctx context.Context, id string) (*models.PositionGrade, error) {
	query := `SELECT id, name, level, min_salary, max_salary, COALESCE(description, ''), is_active, created_at, updated_at FROM position_grades WHERE id::text = $1`
	row := database.Pool.QueryRow(ctx, query, id)

	var g models.PositionGrade
	err := row.Scan(&g.ID, &g.Name, &g.Level, &g.MinSalary, &g.MaxSalary, &g.Description, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func CreatePositionGrade(ctx context.Context, req *models.CreatePositionGradeRequest, userID string) (*models.PositionGrade, error) {
	query := `
		INSERT INTO position_grades (name, level, description, min_salary, max_salary)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, level, min_salary, max_salary, COALESCE(description, ''), is_active, created_at, updated_at
	`
	var g *models.PositionGrade
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, req.Name, req.Level, req.Description, req.MinSalary, req.MaxSalary)
		var result models.PositionGrade
		if err := row.Scan(&result.ID, &result.Name, &result.Level, &result.MinSalary, &result.MaxSalary, &result.Description, &result.IsActive, &result.CreatedAt, &result.UpdatedAt); err != nil {
			return err
		}
		g = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func UpdatePositionGrade(ctx context.Context, id string, req *models.UpdatePositionGradeRequest, userID string) (*models.PositionGrade, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.Level != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("level = $%d", argIdx))
		args = append(args, *req.Level)
	}
	if req.MinSalary != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("min_salary = $%d", argIdx))
		args = append(args, *req.MinSalary)
	}
	if req.MaxSalary != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("max_salary = $%d", argIdx))
		args = append(args, *req.MaxSalary)
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
		return GetPositionGradeByID(ctx, id)
	}

	var g *models.PositionGrade
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE position_grades SET %s
			WHERE id::text = $%d
			RETURNING id, name, level, min_salary, max_salary, COALESCE(description, ''), is_active, created_at, updated_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.PositionGrade
		if err := row.Scan(&result.ID, &result.Name, &result.Level, &result.MinSalary, &result.MaxSalary, &result.Description, &result.IsActive, &result.CreatedAt, &result.UpdatedAt); err != nil {
			return err
		}
		g = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func DeletePositionGrade(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM position_grades WHERE id::text = $1`, id)
		return err
	})
}

// GetPositionGradeByName looks up a position grade ID by name (case-insensitive).
func GetPositionGradeByName(ctx context.Context, name string) (*string, error) {
	var id string
	err := database.Pool.QueryRow(ctx, `SELECT id::text FROM position_grades WHERE LOWER(name) = LOWER($1) LIMIT 1`, name).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CheckPositionGradeNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM position_grades WHERE name = $1`
	args := []interface{}{name}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func CheckPositionGradeLevelExists(ctx context.Context, level int, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM position_grades WHERE level = $1`
	args := []interface{}{level}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func CheckPositionGradeHasPositions(ctx context.Context, id string) (bool, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM positions WHERE grade_id::text = $1 AND deleted_at IS NULL`, id,
	).Scan(&count)
	return count > 0, err
}
