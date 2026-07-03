package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListDailyJournals(ctx context.Context, page, perPage int, departmentID, employeeID, dateFrom, dateTo string) ([]models.DailyJournalSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM daily_journals dj WHERE dj.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if departmentID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND dj.department_id::text = $%d", argIdx)
		args = append(args, departmentID)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND dj.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}
	if dateFrom != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND dj.journal_date >= $%d::date", argIdx)
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND dj.journal_date <= $%d::date", argIdx)
		args = append(args, dateTo)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	listQuery := `
		SELECT dj.id, dj.employee_id, COALESCE(e.full_name, ''),
			dj.journal_date::text,
			LEFT(dj.work_description, 100),
			dj.status::text,
			COALESCE(d.name, ''),
			dj.created_at
		FROM daily_journals dj
		LEFT JOIN employees e ON e.id = dj.employee_id
		LEFT JOIN departments d ON d.id = dj.department_id
		WHERE dj.deleted_at IS NULL
	`
	listArgs := []interface{}{}
	listIdx := 0
	if departmentID != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND dj.department_id::text = $%d", listIdx)
		listArgs = append(listArgs, departmentID)
	}
	if employeeID != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND dj.employee_id::text = $%d", listIdx)
		listArgs = append(listArgs, employeeID)
	}
	if dateFrom != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND dj.journal_date >= $%d::date", listIdx)
		listArgs = append(listArgs, dateFrom)
	}
	if dateTo != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND dj.journal_date <= $%d::date", listIdx)
		listArgs = append(listArgs, dateTo)
	}
	listQuery += fmt.Sprintf(" ORDER BY dj.journal_date DESC, dj.created_at DESC LIMIT $%d OFFSET $%d", listIdx+1, listIdx+2)
	allArgs := append(listArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, listQuery, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var journals []models.DailyJournalSummary
	for rows.Next() {
		var j models.DailyJournalSummary
		if err := rows.Scan(&j.ID, &j.EmployeeID, &j.EmployeeName,
			&j.JournalDate, &j.WorkDescription, &j.Status,
			&j.DepartmentName, &j.CreatedAt); err != nil {
			return nil, 0, err
		}
		journals = append(journals, j)
	}
	return journals, total, nil
}

func GetDailyJournalByID(ctx context.Context, id string) (*models.DailyJournal, error) {
	query := `
		SELECT dj.id, dj.employee_id, COALESCE(e.full_name, ''),
			dj.journal_date::text, dj.work_description,
			COALESCE(dj.achievements, ''), COALESCE(dj.challenges, ''),
			COALESCE(dj.plan_tomorrow, ''),
			dj.status::text, dj.submitted_at,
			dj.acknowledged_by, COALESCE(ack.full_name, ''),
			dj.acknowledged_at,
			dj.department_id, COALESCE(d.name, ''),
			dj.created_at, dj.updated_at
		FROM daily_journals dj
		LEFT JOIN employees e ON e.id = dj.employee_id
		LEFT JOIN employees ack ON ack.id = dj.acknowledged_by
		LEFT JOIN departments d ON d.id = dj.department_id
		WHERE (dj.id::text = $1) AND dj.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var j models.DailyJournal
	err := row.Scan(
		&j.ID, &j.EmployeeID, &j.EmployeeName,
		&j.JournalDate, &j.WorkDescription,
		&j.Achievements, &j.Challenges,
		&j.PlanTomorrow,
		&j.Status, &j.SubmittedAt,
		&j.AcknowledgedBy, &j.AcknowledgedByName,
		&j.AcknowledgedAt,
		&j.DepartmentID, &j.DepartmentName,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func CreateDailyJournal(ctx context.Context, employeeID string, req *models.CreateDailyJournalRequest) (*models.DailyJournal, error) {
	query := `
		INSERT INTO daily_journals (employee_id, journal_date, work_description,
			achievements, challenges, plan_tomorrow, status, department_id)
		SELECT $1::uuid, $2::date, $3, $4, $5, $6, 'submitted'::journal_status, e.department_id
		FROM employees e WHERE e.id = $1::uuid AND e.deleted_at IS NULL
		ON CONFLICT (employee_id, journal_date) DO UPDATE SET
			work_description = EXCLUDED.work_description,
			achievements = EXCLUDED.achievements,
			challenges = EXCLUDED.challenges,
			plan_tomorrow = EXCLUDED.plan_tomorrow,
			status = 'submitted'::journal_status,
			submitted_at = NOW(),
			deleted_at = NULL
		RETURNING id, employee_id, '' as employee_name,
			journal_date::text, work_description,
			COALESCE(achievements, ''), COALESCE(challenges, ''),
			COALESCE(plan_tomorrow, ''),
			status::text, submitted_at,
			acknowledged_by, '' as acknowledged_by_name,
			acknowledged_at,
			department_id, '' as department_name,
			created_at, updated_at
	`
	var j models.DailyJournal
	err := database.Pool.QueryRow(ctx, query,
		employeeID, req.JournalDate, req.WorkDescription,
		req.Achievements, req.Challenges, req.PlanTomorrow,
	).Scan(
		&j.ID, &j.EmployeeID, &j.EmployeeName,
		&j.JournalDate, &j.WorkDescription,
		&j.Achievements, &j.Challenges,
		&j.PlanTomorrow,
		&j.Status, &j.SubmittedAt,
		&j.AcknowledgedBy, &j.AcknowledgedByName,
		&j.AcknowledgedAt,
		&j.DepartmentID, &j.DepartmentName,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func AcknowledgeJournal(ctx context.Context, id, managerID string) (*models.DailyJournal, error) {
	query := `
		UPDATE daily_journals
		SET status = 'acknowledged'::journal_status,
			acknowledged_by = $2::uuid,
			acknowledged_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL
		AND status = 'submitted'::journal_status
		RETURNING id, employee_id, '' as employee_name,
			journal_date::text, work_description,
			COALESCE(achievements, ''), COALESCE(challenges, ''),
			COALESCE(plan_tomorrow, ''),
			status::text, submitted_at,
			acknowledged_by, '' as acknowledged_by_name,
			acknowledged_at,
			department_id, '' as department_name,
			created_at, updated_at
	`
	var j models.DailyJournal
	err := database.Pool.QueryRow(ctx, query, id, managerID).Scan(
		&j.ID, &j.EmployeeID, &j.EmployeeName,
		&j.JournalDate, &j.WorkDescription,
		&j.Achievements, &j.Challenges,
		&j.PlanTomorrow,
		&j.Status, &j.SubmittedAt,
		&j.AcknowledgedBy, &j.AcknowledgedByName,
		&j.AcknowledgedAt,
		&j.DepartmentID, &j.DepartmentName,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("jurnal tidak ditemukan atau sudah diakui")
		}
		return nil, err
	}
	return &j, nil
}
