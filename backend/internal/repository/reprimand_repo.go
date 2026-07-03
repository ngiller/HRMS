package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListReprimands(ctx context.Context, page, perPage int, status, employeeID string) ([]models.ReprimandSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM reprimands r WHERE r.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND r.status = $%d::reprimand_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND r.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	listQuery := `
		SELECT r.id, r.employee_id, COALESCE(e.full_name, ''),
			r.reprimand_type::text, r.title, r.status::text,
			r.issued_date, r.created_at
		FROM reprimands r
		LEFT JOIN employees e ON e.id = r.employee_id
		WHERE r.deleted_at IS NULL
	`
	listArgs := []interface{}{}
	listIdx := 0
	if status != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND r.status = $%d::reprimand_status", listIdx)
		listArgs = append(listArgs, status)
	}
	if employeeID != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND r.employee_id::text = $%d", listIdx)
		listArgs = append(listArgs, employeeID)
	}
	listQuery += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", listIdx+1, listIdx+2)
	allArgs := append(listArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, listQuery, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reprimands []models.ReprimandSummary
	for rows.Next() {
		var r models.ReprimandSummary
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.ReprimandType, &r.Title, &r.Status,
			&r.IssuedDate, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		reprimands = append(reprimands, r)
	}
	return reprimands, total, nil
}

func GetReprimandByID(ctx context.Context, id string) (*models.Reprimand, error) {
	query := `
		SELECT r.id, r.employee_id, COALESCE(e.full_name, ''),
			r.reprimand_type::text, r.title, COALESCE(r.description, ''),
			r.violation_date, COALESCE(r.violation_details, ''),
			r.issued_by, COALESCE(iss.full_name, ''),
			r.issued_date, r.acknowledgment_date,
			COALESCE(r.acknowledgment_note, ''),
			COALESCE(r.document_url, ''),
			r.effective_period_months, r.status::text,
			r.expired_at, r.escalated_from_id,
			r.created_at, r.updated_at
		FROM reprimands r
		LEFT JOIN employees e ON e.id = r.employee_id
		LEFT JOIN employees iss ON iss.id = r.issued_by
		WHERE (r.id::text = $1) AND r.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.Reprimand
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.ReprimandType, &r.Title, &r.Description,
		&r.ViolationDate, &r.ViolationDetails,
		&r.IssuedBy, &r.IssuedByName,
		&r.IssuedDate, &r.AcknowledgmentDate,
		&r.AcknowledgmentNote,
		&r.DocumentURL,
		&r.EffectivePeriodMonths, &r.Status,
		&r.ExpiredAt, &r.EscalatedFromID,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateReprimand(ctx context.Context, req *models.CreateReprimandRequest, issuerID string) (*models.Reprimand, error) {
	expiredAt := fmt.Sprintf("NOW() + INTERVAL '%d months'", req.EffectivePeriodMonths)
	if req.EffectivePeriodMonths <= 0 {
		expiredAt = "NULL"
	}

	query := fmt.Sprintf(`
		INSERT INTO reprimands (employee_id, reprimand_type, title, description,
			violation_date, violation_details, issued_by, issued_date,
			document_url, effective_period_months, status,
			expired_at, escalated_from_id)
		VALUES ($1::uuid, $2::reprimand_type, $3, $4,
			$5::date, $6, $7::uuid, CURRENT_DATE,
			$8, $9, 'issued'::reprimand_status,
			%s, NULL)
		RETURNING id, employee_id, '' as employee_name,
			reprimand_type::text, title, COALESCE(description, ''),
			violation_date, COALESCE(violation_details, ''),
			issued_by, '' as issued_by_name,
			issued_date, acknowledgment_date,
			COALESCE(acknowledgment_note, ''),
			COALESCE(document_url, ''),
			effective_period_months, status::text,
			expired_at, escalated_from_id,
			created_at, updated_at
	`, expiredAt)

	var r models.Reprimand
	err := database.WithUserContext(ctx, issuerID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.EmployeeID, req.ReprimandType, req.Title, req.Description,
			req.ViolationDate, req.ViolationDetails, issuerID,
			req.DocumentURL, req.EffectivePeriodMonths,
		)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.ReprimandType, &r.Title, &r.Description,
			&r.ViolationDate, &r.ViolationDetails,
			&r.IssuedBy, &r.IssuedByName,
			&r.IssuedDate, &r.AcknowledgmentDate,
			&r.AcknowledgmentNote,
			&r.DocumentURL,
			&r.EffectivePeriodMonths, &r.Status,
			&r.ExpiredAt, &r.EscalatedFromID,
			&r.CreatedAt, &r.UpdatedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func AcknowledgeReprimand(ctx context.Context, id, employeeID, note string) (*models.Reprimand, error) {
	query := `
		UPDATE reprimands
		SET status = 'acknowledged'::reprimand_status,
			acknowledgment_date = NOW(),
			acknowledgment_note = $3
		WHERE id::text = $1 AND employee_id::text = $2
		AND deleted_at IS NULL AND status = 'issued'::reprimand_status
		RETURNING id, employee_id, '' as employee_name,
			reprimand_type::text, title, COALESCE(description, ''),
			violation_date, COALESCE(violation_details, ''),
			issued_by, '' as issued_by_name,
			issued_date, acknowledgment_date,
			COALESCE(acknowledgment_note, ''),
			COALESCE(document_url, ''),
			effective_period_months, status::text,
			expired_at, escalated_from_id,
			created_at, updated_at
	`
	var r models.Reprimand
	err := database.WithUserContext(ctx, employeeID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, employeeID, note)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.ReprimandType, &r.Title, &r.Description,
			&r.ViolationDate, &r.ViolationDetails,
			&r.IssuedBy, &r.IssuedByName,
			&r.IssuedDate, &r.AcknowledgmentDate,
			&r.AcknowledgmentNote,
			&r.DocumentURL,
			&r.EffectivePeriodMonths, &r.Status,
			&r.ExpiredAt, &r.EscalatedFromID,
			&r.CreatedAt, &r.UpdatedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("surat peringatan tidak ditemukan atau sudah diakui")
		}
		return nil, err
	}
	return &r, nil
}
