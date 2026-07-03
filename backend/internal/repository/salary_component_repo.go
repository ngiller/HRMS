package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListSalaryComponents(ctx context.Context, employeeID string, page, perPage int) ([]models.SalaryComponentSummary, int, error) {
	// Count
	var total int
	err := database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_salary_components WHERE employee_id::text = $1 AND deleted_at IS NULL`, employeeID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch
	offset := (page - 1) * perPage
	query := `
		SELECT id, employee_id, component_name, component_type::text, amount, is_active, effective_date, created_at
		FROM employee_salary_components
		WHERE employee_id::text = $1 AND deleted_at IS NULL
		ORDER BY component_type, component_name
		LIMIT $2 OFFSET $3
	`

	rows, err := database.Pool.Query(ctx, query, employeeID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var components []models.SalaryComponentSummary
	for rows.Next() {
		var comp models.SalaryComponentSummary
		if err := rows.Scan(&comp.ID, &comp.EmployeeID, &comp.ComponentName, &comp.ComponentType,
			&comp.Amount, &comp.IsActive, &comp.EffectiveDate, &comp.CreatedAt); err != nil {
			return nil, 0, err
		}
		components = append(components, comp)
	}

	return components, total, nil
}

func GetSalaryComponent(ctx context.Context, componentID string) (*models.SalaryComponent, error) {
	query := `
		SELECT id, employee_id, component_name, component_type::text, amount, is_active,
			effective_date, created_by, updated_by, created_at, updated_at
		FROM employee_salary_components
		WHERE id::text = $1 AND deleted_at IS NULL
	`

	row := database.Pool.QueryRow(ctx, query, componentID)
	var comp models.SalaryComponent
	err := row.Scan(&comp.ID, &comp.EmployeeID, &comp.ComponentName, &comp.ComponentType,
		&comp.Amount, &comp.IsActive, &comp.EffectiveDate,
		&comp.CreatedBy, &comp.UpdatedBy, &comp.CreatedAt, &comp.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &comp, nil
}

func CreateSalaryComponent(ctx context.Context, employeeID string, req *models.CreateSalaryComponentRequest, userID string) (*models.SalaryComponent, error) {
	query := `
		INSERT INTO employee_salary_components (employee_id, component_name, component_type, amount, effective_date, created_by, updated_by)
		VALUES ($1::uuid, $2, $3::salary_component_type, $4, NULLIF($5, '')::date, NULLIF($6, '')::uuid, NULLIF($6, '')::uuid)
		RETURNING id, employee_id, component_name, component_type::text, amount, is_active,
			effective_date, created_by, updated_by, created_at, updated_at
	`

	var comp models.SalaryComponent
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			employeeID, req.ComponentName, req.ComponentType, req.Amount, req.EffectiveDate, userID,
		).Scan(&comp.ID, &comp.EmployeeID, &comp.ComponentName, &comp.ComponentType,
			&comp.Amount, &comp.IsActive, &comp.EffectiveDate,
			&comp.CreatedBy, &comp.UpdatedBy, &comp.CreatedAt, &comp.UpdatedAt)
	})
	if err != nil {
		return nil, err
	}

	return &comp, nil
}

func UpdateSalaryComponent(ctx context.Context, componentID string, req *models.UpdateSalaryComponentRequest, userID string) (*models.SalaryComponent, error) {
	sets := []string{}
	args := []interface{}{}
	argIdx := 0

	addSet := func(col string, val interface{}) {
		argIdx++
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
	}

	if req.ComponentName != nil && *req.ComponentName != "" {
		addSet("component_name", *req.ComponentName)
	}
	if req.ComponentType != nil && *req.ComponentType != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("component_type = $%d::salary_component_type", argIdx))
		args = append(args, *req.ComponentType)
	}
	if req.Amount != nil {
		addSet("amount", *req.Amount)
	}
	if req.IsActive != nil {
		addSet("is_active", *req.IsActive)
	}
	if req.EffectiveDate != nil && *req.EffectiveDate != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("effective_date = NULLIF($%d, '')::date", argIdx))
		args = append(args, *req.EffectiveDate)
	}

	if len(sets) == 0 {
		return nil, errors.New("tidak ada data yang diubah")
	}

	argIdx++
	sets = append(sets, fmt.Sprintf("updated_by = $%d::uuid", argIdx))
	args = append(args, userID)

	argIdx++
	query := fmt.Sprintf(`
		UPDATE employee_salary_components SET %s, updated_at = NOW()
		WHERE id::text = $%d AND deleted_at IS NULL
		RETURNING id, employee_id, component_name, component_type::text, amount, is_active,
			effective_date, created_by, updated_by, created_at, updated_at
	`, joinStrings(sets, ", "), argIdx)
	args = append(args, componentID)

	var comp models.SalaryComponent
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, args...).Scan(
			&comp.ID, &comp.EmployeeID, &comp.ComponentName, &comp.ComponentType,
			&comp.Amount, &comp.IsActive, &comp.EffectiveDate,
			&comp.CreatedBy, &comp.UpdatedBy, &comp.CreatedAt, &comp.UpdatedAt)
	})
	if err != nil {
		return nil, err
	}

	return &comp, nil
}

func DeleteSalaryComponent(ctx context.Context, componentID, userID string) error {
	query := `UPDATE employee_salary_components SET deleted_at = NOW(), updated_at = NOW(), updated_by = $2::uuid WHERE id::text = $1 AND deleted_at IS NULL`

	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, componentID, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("komponen gaji tidak ditemukan")
		}
		return nil
	})
}
