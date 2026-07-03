package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListHolidays(ctx context.Context, page, perPage int, year int, holidayType string) ([]models.CompanyHoliday, int, error) {
	countQuery := `SELECT COUNT(*) FROM company_holidays WHERE is_active = TRUE`
	args := []interface{}{}
	argIdx := 0

	if year > 0 {
		argIdx++
		countQuery += fmt.Sprintf(" AND EXTRACT(YEAR FROM date) = $%d", argIdx)
		args = append(args, year)
	}
	if holidayType != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND holiday_type = $%d::holiday_type", argIdx)
		args = append(args, holidayType)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT h.id, to_char(h.date, 'YYYY-MM-DD'), h.name,
			h.holiday_type::text, h.is_recurring_yearly, COALESCE(h.description, ''),
			h.is_active, h.created_by, COALESCE(e.full_name, ''),
			h.created_at, h.updated_at
		FROM company_holidays h
		LEFT JOIN employees e ON e.id = h.created_by
		WHERE h.is_active = TRUE
	`
	searchArgs := []interface{}{}
	if year > 0 {
		query += fmt.Sprintf(" AND EXTRACT(YEAR FROM h.date) = $%d", len(searchArgs)+1)
		searchArgs = append(searchArgs, year)
	}
	if holidayType != "" {
		query += fmt.Sprintf(" AND h.holiday_type = $%d::holiday_type", len(searchArgs)+1)
		searchArgs = append(searchArgs, holidayType)
	}
	query += fmt.Sprintf(" ORDER BY h.date ASC LIMIT $%d OFFSET $%d", len(searchArgs)+1, len(searchArgs)+2)
	allArgs := append(searchArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var holidays []models.CompanyHoliday
	for rows.Next() {
		var h models.CompanyHoliday
		if err := rows.Scan(&h.ID, &h.Date, &h.Name,
			&h.HolidayType, &h.IsRecurringYearly, &h.Description,
			&h.IsActive, &h.CreatedBy, &h.CreatedByName,
			&h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, 0, err
		}
		holidays = append(holidays, h)
	}
	return holidays, total, nil
}

func GetHolidayByID(ctx context.Context, id string) (*models.CompanyHoliday, error) {
	query := `
		SELECT h.id, to_char(h.date, 'YYYY-MM-DD'), h.name,
			h.holiday_type::text, h.is_recurring_yearly, COALESCE(h.description, ''),
			h.is_active, h.created_by, COALESCE(e.full_name, ''),
			h.created_at, h.updated_at
		FROM company_holidays h
		LEFT JOIN employees e ON e.id = h.created_by
		WHERE h.id::text = $1
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var h models.CompanyHoliday
	err := row.Scan(
		&h.ID, &h.Date, &h.Name,
		&h.HolidayType, &h.IsRecurringYearly, &h.Description,
		&h.IsActive, &h.CreatedBy, &h.CreatedByName,
		&h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func CreateHoliday(ctx context.Context, createdBy string, req *models.CreateHolidayReq) (*models.CompanyHoliday, error) {
	query := `
		INSERT INTO company_holidays (date, name, holiday_type, is_recurring_yearly, description, created_by)
		VALUES ($1::date, $2, $3::holiday_type, $4, $5, $6::uuid)
		RETURNING id, to_char(date, 'YYYY-MM-DD'), name,
			holiday_type::text, is_recurring_yearly, COALESCE(description, ''),
			is_active, created_by, '' as created_by_name,
			created_at, updated_at
	`
	var h models.CompanyHoliday
	err := database.WithUserContext(ctx, createdBy, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.Date, req.Name, req.HolidayType, req.IsRecurringYearly, req.Description, createdBy,
		)
		return row.Scan(
			&h.ID, &h.Date, &h.Name,
			&h.HolidayType, &h.IsRecurringYearly, &h.Description,
			&h.IsActive, &h.CreatedBy, &h.CreatedByName,
			&h.CreatedAt, &h.UpdatedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func UpdateHoliday(ctx context.Context, id, userID string, req *models.UpdateHolidayReq) (*models.CompanyHoliday, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Date != nil && *req.Date != "" {
		setClauses = append(setClauses, fmt.Sprintf("date = $%d::date", argIdx))
		args = append(args, *req.Date)
		argIdx++
	}
	if req.Name != nil && *req.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.HolidayType != nil && *req.HolidayType != "" {
		setClauses = append(setClauses, fmt.Sprintf("holiday_type = $%d::holiday_type", argIdx))
		args = append(args, *req.HolidayType)
		argIdx++
	}
	if req.IsRecurringYearly != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_recurring_yearly = $%d", argIdx))
		args = append(args, *req.IsRecurringYearly)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(setClauses) == 0 {
		return GetHolidayByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE company_holidays SET %s, updated_at = NOW()
		WHERE id::text = $%d
		RETURNING id, to_char(date, 'YYYY-MM-DD'), name,
			holiday_type::text, is_recurring_yearly, COALESCE(description, ''),
			is_active, created_by, '' as created_by_name,
			created_at, updated_at
	`, joinStrings(setClauses, ", "), argIdx)
	args = append(args, id)

	var h models.CompanyHoliday
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, args...)
		return row.Scan(
			&h.ID, &h.Date, &h.Name,
			&h.HolidayType, &h.IsRecurringYearly, &h.Description,
			&h.IsActive, &h.CreatedBy, &h.CreatedByName,
			&h.CreatedAt, &h.UpdatedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func DeleteHoliday(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE company_holidays SET deleted_at = NOW() WHERE id::text = $1`, id)
		return err
	})
}

// GetHolidaysByYear returns all holidays for a specific year
func GetHolidaysByYear(ctx context.Context, year int) ([]models.CompanyHoliday, error) {
	query := `
		SELECT h.id, to_char(h.date, 'YYYY-MM-DD'), h.name,
			h.holiday_type::text, h.is_recurring_yearly, COALESCE(h.description, ''),
			h.is_active, h.created_by, COALESCE(e.full_name, ''),
			h.created_at, h.updated_at
		FROM company_holidays h
		LEFT JOIN employees e ON e.id = h.created_by
		WHERE EXTRACT(YEAR FROM h.date) = $1 AND h.is_active = TRUE
		ORDER BY h.date ASC
	`
	rows, err := database.Pool.Query(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holidays []models.CompanyHoliday
	for rows.Next() {
		var h models.CompanyHoliday
		if err := rows.Scan(&h.ID, &h.Date, &h.Name,
			&h.HolidayType, &h.IsRecurringYearly, &h.Description,
			&h.IsActive, &h.CreatedBy, &h.CreatedByName,
			&h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, nil
}
