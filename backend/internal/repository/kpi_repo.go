package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ListKPITemplates(ctx context.Context, page, perPage int, year int) ([]models.KPITemplate, int, error) {
	countQuery := `SELECT COUNT(*) FROM kpi_templates kt WHERE kt.deleted_at IS NULL`
	args := []interface{}{}
	if year > 0 {
		countQuery += " AND kt.year = $1"
		args = append(args, year)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	allArgs := append(args, perPage, offset)
	query := fmt.Sprintf(`
		SELECT kt.id, kt.title, kt.position_id, COALESCE(p.name, ''),
			kt.department_id, COALESCE(d.name, ''),
			kt.period_type, kt.year, COALESCE(kt.description, ''),
			kt.is_active, kt.created_by, COALESCE(e.full_name, ''),
			kt.created_at, kt.updated_at
		FROM kpi_templates kt
		LEFT JOIN positions p ON p.id = kt.position_id
		LEFT JOIN departments d ON d.id = kt.department_id
		LEFT JOIN employees e ON e.id = kt.created_by
		WHERE kt.deleted_at IS NULL
		ORDER BY kt.created_at DESC
		LIMIT $%d OFFSET $%d
	`, len(args)+1, len(args)+2)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var templates []models.KPITemplate
	for rows.Next() {
		var t models.KPITemplate
		if err := rows.Scan(&t.ID, &t.Title, &t.PositionID, &t.PositionName,
			&t.DepartmentID, &t.DeptName,
			&t.PeriodType, &t.Year, &t.Description,
			&t.IsActive, &t.CreatedBy, &t.CreatedByName,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		templates = append(templates, t)
	}
	return templates, total, nil
}

func GetKPITemplateByID(ctx context.Context, id string) (*models.KPITemplateDetail, error) {
	query := `
		SELECT kt.id, kt.title, kt.position_id, COALESCE(p.name, ''),
			kt.department_id, COALESCE(d.name, ''),
			kt.period_type, kt.year, COALESCE(kt.description, ''),
			kt.is_active, kt.created_by, COALESCE(e.full_name, ''),
			kt.created_at, kt.updated_at
		FROM kpi_templates kt
		LEFT JOIN positions p ON p.id = kt.position_id
		LEFT JOIN departments d ON d.id = kt.department_id
		LEFT JOIN employees e ON e.id = kt.created_by
		WHERE (kt.id::text = $1) AND kt.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var t models.KPITemplateDetail
	err := row.Scan(&t.ID, &t.Title, &t.PositionID, &t.PositionName,
		&t.DepartmentID, &t.DeptName,
		&t.PeriodType, &t.Year, &t.Description,
		&t.IsActive, &t.CreatedBy, &t.CreatedByName,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get indicators
	indicatorQuery := `
		SELECT id, kpi_template_id, name, COALESCE(description, ''),
			target, weight, COALESCE(measurement_unit, ''), sort_order
		FROM kpi_indicators
		WHERE kpi_template_id = $1
		ORDER BY sort_order ASC
	`
	rows, err := database.Pool.Query(ctx, indicatorQuery, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ind models.KPIIndicator
		if err := rows.Scan(&ind.ID, &ind.KPITemplateID, &ind.Name, &ind.Description,
			&ind.Target, &ind.Weight, &ind.MeasurementUnit, &ind.SortOrder); err != nil {
			return nil, err
		}
		t.Indicators = append(t.Indicators, ind)
	}
	if t.Indicators == nil {
		t.Indicators = []models.KPIIndicator{}
	}

	return &t, nil
}

func ListKPIReviews(ctx context.Context, page, perPage int, status, employeeID string) ([]models.KPIReviewSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM kpi_reviews kr WHERE kr.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND kr.status = $%d::kpi_review_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND kr.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}


	offset := (page - 1) * perPage
	finalQuery := `
		SELECT kr.id, kr.employee_id, COALESCE(e.full_name, ''),
			COALESCE(kt.title, ''), kr.period, kr.year,
			kr.final_score, COALESCE(kr.final_category, ''),
			kr.status::text, kr.created_at
		FROM kpi_reviews kr
		LEFT JOIN employees e ON e.id = kr.employee_id
		LEFT JOIN kpi_templates kt ON kt.id = kr.kpi_template_id
		WHERE kr.deleted_at IS NULL
	`
	finalArgs := []interface{}{}
	if status != "" {
		finalArgs = append(finalArgs, status)
		finalQuery += fmt.Sprintf(" AND kr.status = $%d::kpi_review_status", len(finalArgs))
	}
	if employeeID != "" {
		finalArgs = append(finalArgs, employeeID)
		finalQuery += fmt.Sprintf(" AND kr.employee_id::text = $%d", len(finalArgs))
	}
	finalQuery += " ORDER BY kr.created_at DESC"
	finalArgs = append(finalArgs, perPage, offset)
	finalQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(finalArgs)-1, len(finalArgs))

	rows, err := database.Pool.Query(ctx, finalQuery, finalArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []models.KPIReviewSummary
	for rows.Next() {
		var r models.KPIReviewSummary
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.TemplateTitle, &r.Period, &r.Year,
			&r.FinalScore, &r.FinalCategory,
			&r.Status, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, r)
	}
	return reviews, total, nil
}

func GetKPIReviewByID(ctx context.Context, id string) (*models.KPIReview, error) {
	query := `
		SELECT kr.id, kr.employee_id, COALESCE(e.full_name, ''),
			kr.kpi_template_id, COALESCE(kt.title, ''),
			kr.period, kr.year,
			COALESCE(kr.self_rating::text, '[]'), kr.self_score,
			COALESCE(kr.self_note, ''), kr.self_submitted_at,
			COALESCE(kr.manager_rating::text, '[]'), kr.manager_score,
			COALESCE(kr.manager_note, ''), kr.manager_id,
			COALESCE(m.full_name, ''), kr.manager_reviewed_at,
			COALESCE(kr.hr_rating::text, '[]'), kr.final_score,
			COALESCE(kr.final_category, ''), COALESCE(kr.hr_note, ''),
			kr.hr_id, COALESCE(h.full_name, ''), kr.hr_reviewed_at,
			kr.status::text, kr.salary_increase, kr.bonus_amount,
			kr.created_at, kr.updated_at
		FROM kpi_reviews kr
		LEFT JOIN employees e ON e.id = kr.employee_id
		LEFT JOIN kpi_templates kt ON kt.id = kr.kpi_template_id
		LEFT JOIN employees m ON m.id = kr.manager_id
		LEFT JOIN employees h ON h.id = kr.hr_id
		WHERE (kr.id::text = $1) AND kr.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.KPIReview
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.KPITemplateID, &r.TemplateTitle,
		&r.Period, &r.Year,
		&r.SelfRating, &r.SelfScore,
		&r.SelfNote, &r.SelfSubmittedAt,
		&r.ManagerRating, &r.ManagerScore,
		&r.ManagerNote, &r.ManagerID,
		&r.ManagerName, &r.ManagerReviewedAt,
		&r.HRRating, &r.FinalScore,
		&r.FinalCategory, &r.HRNote,
		&r.HRID, &r.HRName, &r.HRReviewedAt,
		&r.Status, &r.SalaryIncrease, &r.BonusAmount,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateKPIReview(ctx context.Context, req *models.CreateKPIReviewRequest, userID string) (*models.KPIReview, error) {
	var r *models.KPIReview
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO kpi_reviews (employee_id, kpi_template_id, period, year, status)
			VALUES ($1::uuid, $2::uuid, $3, $4, 'draft'::kpi_review_status)
			RETURNING id, employee_id, '' as employee_name,
				kpi_template_id, '' as template_title,
				period, year,
				'[]', NULL, '', NULL,
				'[]', NULL, '', NULL, '', NULL,
				'[]', NULL, '', '', NULL, '', NULL,
				status::text, NULL, NULL,
				created_at, updated_at
		`
		row := tx.QueryRow(ctx, query,
			req.EmployeeID, req.KPITemplateID, req.Period, req.Year)
		var result models.KPIReview
		if err := row.Scan(
			&result.ID, &result.EmployeeID, &result.EmployeeName,
			&result.KPITemplateID, &result.TemplateTitle,
			&result.Period, &result.Year,
			&result.SelfRating, &result.SelfScore, &result.SelfNote, &result.SelfSubmittedAt,
			&result.ManagerRating, &result.ManagerScore, &result.ManagerNote, &result.ManagerID, &result.ManagerName, &result.ManagerReviewedAt,
			&result.HRRating, &result.FinalScore, &result.FinalCategory, &result.HRNote, &result.HRID, &result.HRName, &result.HRReviewedAt,
			&result.Status, &result.SalaryIncrease, &result.BonusAmount,
			&result.CreatedAt, &result.UpdatedAt,
		); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return errors.New("review KPI untuk karyawan pada periode dan tahun tersebut sudah ada")
			}
			return err
		}
		r = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func CreateKPITemplate(ctx context.Context, req *models.CreateKPITemplateRequest, createdBy string) (*models.KPITemplateDetail, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Insert template
	var templateID string
	err = tx.QueryRow(ctx, `
		INSERT INTO kpi_templates (title, position_id, department_id, period_type, year, description, is_active, created_by)
		VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6, TRUE, $7::uuid)
		RETURNING id::text
	`, req.Title, req.PositionID, req.DepartmentID, req.PeriodType, req.Year, req.Description, createdBy).Scan(&templateID)
	if err != nil {
		return nil, err
	}

	// Insert indicators
	for _, ind := range req.Indicators {
		_, err = tx.Exec(ctx, `
			INSERT INTO kpi_indicators (kpi_template_id, name, description, target, weight, measurement_unit, sort_order)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
		`, templateID, ind.Name, ind.Description, ind.Target, ind.Weight, ind.MeasurementUnit, ind.SortOrder)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Return full detail
	return GetKPITemplateByID(ctx, templateID)
}

func UpdateKPITemplate(ctx context.Context, id string, req *models.UpdateKPITemplateRequest) (*models.KPITemplateDetail, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	isActiveClause := ""
	args := []interface{}{req.Title, req.PositionID, req.DepartmentID, req.PeriodType, req.Year, req.Description, id}
	argIdx := 7
	if req.IsActive != nil {
		argIdx++
		isActiveClause = fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *req.IsActive)
	}

	query := fmt.Sprintf(`
		UPDATE kpi_templates
		SET title = $1, position_id = $2::uuid, department_id = $3::uuid,
			period_type = $4, year = $5, description = $6%s
		WHERE id::text = $7 AND deleted_at IS NULL
	`, isActiveClause)

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("template KPI tidak ditemukan")
	}

	// Replace indicators: delete old, insert new
	_, err = tx.Exec(ctx, `DELETE FROM kpi_indicators WHERE kpi_template_id = $1::uuid`, id)
	if err != nil {
		return nil, err
	}

	for _, ind := range req.Indicators {
		_, err = tx.Exec(ctx, `
			INSERT INTO kpi_indicators (kpi_template_id, name, description, target, weight, measurement_unit, sort_order)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
		`, id, ind.Name, ind.Description, ind.Target, ind.Weight, ind.MeasurementUnit, ind.SortOrder)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return GetKPITemplateByID(ctx, id)
}

func DeleteKPITemplate(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE kpi_templates SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL
		`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("template KPI tidak ditemukan")
		}
		return nil
	})
}
