package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListOvertimeRequests(ctx context.Context, page, perPage int, status, employeeID string) ([]models.OvertimeRequestSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM overtime_requests otr WHERE otr.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND otr.status = $%d::leave_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND otr.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	argIdx++
	query := fmt.Sprintf(`
		SELECT otr.id, otr.employee_id, COALESCE(e.full_name, ''),
			otr.date::text, otr.total_hours, otr.overtime_type::text,
			COALESCE(otr.reason, ''), otr.status::text, otr.created_at
		FROM overtime_requests otr
		LEFT JOIN employees e ON e.id = otr.employee_id
		WHERE otr.deleted_at IS NULL
	`)
	searchArgs := []interface{}{}
	if status != "" {
		query += fmt.Sprintf(" AND otr.status = $%d::leave_status", len(searchArgs)+1)
		searchArgs = append(searchArgs, status)
	}
	if employeeID != "" {
		query += fmt.Sprintf(" AND otr.employee_id::text = $%d", len(searchArgs)+1)
		searchArgs = append(searchArgs, employeeID)
	}
	query += fmt.Sprintf(" ORDER BY otr.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	allArgs := append(searchArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []models.OvertimeRequestSummary
	for rows.Next() {
		var r models.OvertimeRequestSummary
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Date, &r.TotalHours, &r.OvertimeType,
			&r.Reason, &r.Status, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		requests = append(requests, r)
	}
	return requests, total, nil
}

func GetOvertimeRequestByID(ctx context.Context, id string) (*models.OvertimeRequest, error) {
	query := `
		SELECT otr.id, otr.employee_id, COALESCE(e.full_name, ''),
			otr.date::text, otr.start_time, otr.end_time,
			otr.total_hours, otr.overtime_type::text,
			COALESCE(otr.reason, ''), otr.is_mandatory,
			COALESCE(otr.approval_trail::text, '[]'), otr.status::text,
			COALESCE(otr.rejection_reason, ''),
			otr.cancelled_by, otr.cancelled_at,
			otr.created_at, otr.updated_at, otr.deleted_at
		FROM overtime_requests otr
		LEFT JOIN employees e ON e.id = otr.employee_id
		WHERE (otr.id::text = $1) AND otr.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.OvertimeRequest
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.Date, &r.StartTime, &r.EndTime,
		&r.TotalHours, &r.OvertimeType,
		&r.Reason, &r.IsMandatory,
		&r.ApprovalTrail, &r.Status,
		&r.RejectionReason,
		&r.CancelledBy, &r.CancelledAt,
		&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateOvertimeRequest(ctx context.Context, employeeID string, req *models.CreateOvertimeRequestReq) (*models.OvertimeRequest, error) {
	query := `
		INSERT INTO overtime_requests (employee_id, date, start_time, end_time,
			total_hours, overtime_type, reason, is_mandatory, status)
		VALUES ($1::uuid, $2::date, $3::timestamptz, $4::timestamptz,
			$5, $6::overtime_type, $7, $8, 'pending'::leave_status)
		RETURNING id, employee_id, '' as employee_name,
			date::text, start_time, end_time,
			total_hours, overtime_type::text,
			COALESCE(reason, ''), is_mandatory,
			COALESCE(approval_trail::text, '[]'), status::text,
			COALESCE(rejection_reason, ''),
			cancelled_by, cancelled_at,
			created_at, updated_at, deleted_at
	`
	var r models.OvertimeRequest
	err := database.WithUserContext(ctx, employeeID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			employeeID, req.Date, req.StartTime, req.EndTime,
			req.TotalHours, req.OvertimeType, req.Reason, req.IsMandatory,
		)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Date, &r.StartTime, &r.EndTime,
			&r.TotalHours, &r.OvertimeType,
			&r.Reason, &r.IsMandatory,
			&r.ApprovalTrail, &r.Status,
			&r.RejectionReason,
			&r.CancelledBy, &r.CancelledAt,
			&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func UpdateOvertimeStatus(ctx context.Context, id, status, rejectionReason, approverID string) (*models.OvertimeRequest, error) {
	query := `
		UPDATE overtime_requests
		SET status = $2::leave_status,
			rejection_reason = $3,
			approval_trail = approval_trail || $4::jsonb,
			rejected_at = CASE WHEN $2::leave_status = 'rejected' THEN NOW() ELSE rejected_at END,
			rejected_by = CASE WHEN $2::leave_status = 'rejected' THEN $5::uuid ELSE rejected_by END
		WHERE id::text = $1 AND deleted_at IS NULL
		AND status = 'pending'
		RETURNING id, employee_id, '' as employee_name,
			date::text, start_time, end_time,
			total_hours, overtime_type::text,
			COALESCE(reason, ''), is_mandatory,
			COALESCE(approval_trail::text, '[]'), status::text,
			COALESCE(rejection_reason, ''),
			cancelled_by, cancelled_at,
			created_at, updated_at, deleted_at
	`
	approvalEntry := fmt.Sprintf(`[{"status":"%s","approver_id":"%s","date":"%s"}]`, status, approverID, time.Now().Format(time.RFC3339))
	var r models.OvertimeRequest
	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, status, rejectionReason, approvalEntry, approverID)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Date, &r.StartTime, &r.EndTime,
			&r.TotalHours, &r.OvertimeType,
			&r.Reason, &r.IsMandatory,
			&r.ApprovalTrail, &r.Status,
			&r.RejectionReason,
			&r.CancelledBy, &r.CancelledAt,
			&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("pengajuan lembur tidak ditemukan atau sudah diproses")
		}
		return nil, err
	}
	return &r, nil
}

func CancelOvertimeRequest(ctx context.Context, id, employeeID string) error {
	query := `UPDATE overtime_requests SET status = 'cancelled'::leave_status,
		cancelled_by = $2::uuid, cancelled_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL AND employee_id::text = $2
		AND status = 'pending'`
	return database.WithUserContext(ctx, employeeID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, id, employeeID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("pengajuan lembur tidak ditemukan atau sudah diproses")
		}
		return nil
	})
}

func GetOvertimeCalculation(ctx context.Context, id string) (*models.OvertimeCalculationResponse, error) {
	query := `SELECT id, employee_id, date::text, total_hours,
		overtime_type::text, COALESCE(base_salary, 0), COALESCE(hourly_rate, 0),
		COALESCE(rate_segments::text, '[]'), COALESCE(overtime_pay, 0)
		FROM overtime_calculation WHERE id::text = $1`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.OvertimeCalculationResponse
	err := row.Scan(&r.ID, &r.EmployeeID, &r.Date, &r.TotalHours,
		&r.OvertimeType, &r.BaseSalary, &r.HourlyRate,
		&r.RateSegments, &r.OvertimePay)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
