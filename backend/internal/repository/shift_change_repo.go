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

func ListShiftChangeRequests(ctx context.Context, page, perPage int, status, employeeID string) ([]models.ShiftChangeRequestSummary, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM shift_change_requests scr
		WHERE scr.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND scr.status = $%d::shift_change_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND (scr.employee_id::text = $%d OR scr.swap_partner_id::text = $%d)", argIdx, argIdx)
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
		SELECT scr.id, scr.request_type::text,
			scr.employee_id, COALESCE(e.full_name, ''),
			scr.target_date::text,
			COALESCE(ws_current.name, ''),
			COALESCE(ws_requested.name, ''),
			COALESCE(sp.full_name, ''),
			COALESCE(scr.reason, ''),
			scr.status::text,
			scr.created_at
		FROM shift_change_requests scr
		LEFT JOIN employees e ON e.id = scr.employee_id
		LEFT JOIN work_schedules ws_current ON ws_current.id = scr.current_schedule_id
		LEFT JOIN work_schedules ws_requested ON ws_requested.id = scr.requested_schedule_id
		LEFT JOIN employees sp ON sp.id = scr.swap_partner_id
		WHERE scr.deleted_at IS NULL
	`)
	searchArgs := []interface{}{}
	if status != "" {
		query += fmt.Sprintf(" AND scr.status = $%d::shift_change_status", len(searchArgs)+1)
		searchArgs = append(searchArgs, status)
	}
	if employeeID != "" {
		n := len(searchArgs) + 1
		query += fmt.Sprintf(" AND (scr.employee_id::text = $%d OR scr.swap_partner_id::text = $%d)", n, n)
		searchArgs = append(searchArgs, employeeID)
	}
	query += fmt.Sprintf(" ORDER BY scr.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	allArgs := append(searchArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []models.ShiftChangeRequestSummary
	for rows.Next() {
		var r models.ShiftChangeRequestSummary
		if err := rows.Scan(&r.ID, &r.RequestType, &r.EmployeeID, &r.EmployeeName,
			&r.TargetDate, &r.CurrentScheduleName, &r.RequestedScheduleName,
			&r.SwapPartnerName, &r.Reason, &r.Status, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		requests = append(requests, r)
	}
	return requests, total, nil
}

func GetShiftChangeRequestByID(ctx context.Context, id string) (*models.ShiftChangeRequest, error) {
	query := `
		SELECT scr.id, scr.request_type::text,
			scr.employee_id, COALESCE(e.full_name, ''),
			scr.target_date::text,
			scr.current_schedule_id, COALESCE(ws_current.name, ''),
			scr.requested_schedule_id, COALESCE(ws_requested.name, ''),
			scr.swap_partner_id, COALESCE(sp.full_name, ''),
			scr.swap_partner_date::text,
			scr.swap_partner_schedule_id,
			COALESCE(scr.reason, ''),
			scr.swap_partner_confirmed, scr.swap_partner_confirmed_at,
			scr.status::text, COALESCE(scr.approval_trail::text, '[]'),
			COALESCE(scr.rejection_reason, ''),
			scr.cancelled_by, scr.cancelled_at,
			scr.created_at, scr.updated_at, scr.deleted_at
		FROM shift_change_requests scr
		LEFT JOIN employees e ON e.id = scr.employee_id
		LEFT JOIN work_schedules ws_current ON ws_current.id = scr.current_schedule_id
		LEFT JOIN work_schedules ws_requested ON ws_requested.id = scr.requested_schedule_id
		LEFT JOIN employees sp ON sp.id = scr.swap_partner_id
		WHERE (scr.id::text = $1) AND scr.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.ShiftChangeRequest
	err := row.Scan(
		&r.ID, &r.RequestType,
		&r.EmployeeID, &r.EmployeeName,
		&r.TargetDate,
		&r.CurrentScheduleID, &r.CurrentScheduleName,
		&r.RequestedScheduleID, &r.RequestedScheduleName,
		&r.SwapPartnerID, &r.SwapPartnerName,
		&r.SwapPartnerDate,
		&r.SwapPartnerScheduleID,
		&r.Reason,
		&r.SwapPartnerConfirmed, &r.SwapPartnerConfirmedAt,
		&r.Status, &r.ApprovalTrail,
		&r.RejectionReason,
		&r.CancelledBy, &r.CancelledAt,
		&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateShiftChangeRequest(ctx context.Context, employeeID, userID string, req *models.CreateShiftChangeRequestReq) (*models.ShiftChangeRequest, error) {
	status := "pending"
	if req.RequestType == "swap" {
		status = "partner_pending"
	}

	var r *models.ShiftChangeRequest
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO shift_change_requests (request_type, employee_id, target_date,
				current_schedule_id, requested_schedule_id,
				swap_partner_id, swap_partner_date, swap_partner_schedule_id,
				reason, status)
			VALUES ($1::shift_change_type, $2::uuid, $3::date,
				NULLIF(NULLIF($4, ''), 'null')::uuid, $5::uuid,
				NULLIF(NULLIF($6, ''), 'null')::uuid, NULLIF(NULLIF($7, ''), 'null')::date, NULLIF(NULLIF($8, ''), 'null')::uuid,
				$9, $10::shift_change_status)
			RETURNING id, request_type::text,
				employee_id, '' as employee_name,
				target_date::text,
				current_schedule_id, '' as current_schedule_name,
				requested_schedule_id, '' as requested_schedule_name,
				swap_partner_id, '' as swap_partner_name,
				swap_partner_date::text,
				swap_partner_schedule_id,
				COALESCE(reason, ''),
				swap_partner_confirmed, swap_partner_confirmed_at,
				status::text, COALESCE(approval_trail::text, '[]'),
				COALESCE(rejection_reason, ''),
				cancelled_by, cancelled_at,
				created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, query,
			req.RequestType, employeeID, req.TargetDate,
			req.CurrentScheduleID, req.RequestedScheduleID,
			req.SwapPartnerID, req.SwapPartnerDate, req.SwapPartnerScheduleID,
			req.Reason, status,
		)
		var result models.ShiftChangeRequest
		if err := row.Scan(
			&result.ID, &result.RequestType,
			&result.EmployeeID, &result.EmployeeName,
			&result.TargetDate,
			&result.CurrentScheduleID, &result.CurrentScheduleName,
			&result.RequestedScheduleID, &result.RequestedScheduleName,
			&result.SwapPartnerID, &result.SwapPartnerName,
			&result.SwapPartnerDate,
			&result.SwapPartnerScheduleID,
			&result.Reason,
			&result.SwapPartnerConfirmed, &result.SwapPartnerConfirmedAt,
			&result.Status, &result.ApprovalTrail,
			&result.RejectionReason,
			&result.CancelledBy, &result.CancelledAt,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
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

func UpdateShiftChangeStatus(ctx context.Context, id, status, rejectionReason, approverID string) (*models.ShiftChangeRequest, error) {
	var r *models.ShiftChangeRequest
	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		query := `
			UPDATE shift_change_requests
			SET status = $2::shift_change_status,
				rejection_reason = $3,
				approval_trail = approval_trail || $4::jsonb
			WHERE id::text = $1 AND deleted_at IS NULL
			AND status IN ('pending', 'partner_pending')
			RETURNING id, request_type::text,
				employee_id, '' as employee_name,
				target_date::text,
				current_schedule_id, '' as current_schedule_name,
				requested_schedule_id, '' as requested_schedule_name,
				swap_partner_id, '' as swap_partner_name,
				swap_partner_date::text,
				swap_partner_schedule_id,
				COALESCE(reason, ''),
				swap_partner_confirmed, swap_partner_confirmed_at,
				status::text, COALESCE(approval_trail::text, '[]'),
				COALESCE(rejection_reason, ''),
				cancelled_by, cancelled_at,
				created_at, updated_at, deleted_at
		`
		approvalEntry := fmt.Sprintf(`[{"status":"%s","approver_id":"%s","date":"%s"}]`, status, approverID, time.Now().Format(time.RFC3339))
		row := tx.QueryRow(ctx, query, id, status, rejectionReason, approvalEntry)
		var result models.ShiftChangeRequest
		if err := row.Scan(
			&result.ID, &result.RequestType,
			&result.EmployeeID, &result.EmployeeName,
			&result.TargetDate,
			&result.CurrentScheduleID, &result.CurrentScheduleName,
			&result.RequestedScheduleID, &result.RequestedScheduleName,
			&result.SwapPartnerID, &result.SwapPartnerName,
			&result.SwapPartnerDate,
			&result.SwapPartnerScheduleID,
			&result.Reason,
			&result.SwapPartnerConfirmed, &result.SwapPartnerConfirmedAt,
			&result.Status, &result.ApprovalTrail,
			&result.RejectionReason,
			&result.CancelledBy, &result.CancelledAt,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		r = &result
		return nil
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("permintaan shift tidak ditemukan atau sudah diproses")
		}
		return nil, err
	}
	return r, nil
}

func ConfirmSwapPartner(ctx context.Context, id, userID string) (*models.ShiftChangeRequest, error) {
	var r *models.ShiftChangeRequest
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			UPDATE shift_change_requests
			SET swap_partner_confirmed = TRUE,
				swap_partner_confirmed_at = NOW(),
				status = 'pending'::shift_change_status
			WHERE id::text = $1 AND deleted_at IS NULL
			AND status = 'partner_pending'
			RETURNING id, request_type::text,
				employee_id, '' as employee_name,
				target_date::text,
				current_schedule_id, '' as current_schedule_name,
				requested_schedule_id, '' as requested_schedule_name,
				swap_partner_id, '' as swap_partner_name,
				swap_partner_date::text,
				swap_partner_schedule_id,
				COALESCE(reason, ''),
				swap_partner_confirmed, swap_partner_confirmed_at,
				status::text, COALESCE(approval_trail::text, '[]'),
				COALESCE(rejection_reason, ''),
				cancelled_by, cancelled_at,
				created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, query, id)
		var result models.ShiftChangeRequest
		if err := row.Scan(
			&result.ID, &result.RequestType,
			&result.EmployeeID, &result.EmployeeName,
			&result.TargetDate,
			&result.CurrentScheduleID, &result.CurrentScheduleName,
			&result.RequestedScheduleID, &result.RequestedScheduleName,
			&result.SwapPartnerID, &result.SwapPartnerName,
			&result.SwapPartnerDate,
			&result.SwapPartnerScheduleID,
			&result.Reason,
			&result.SwapPartnerConfirmed, &result.SwapPartnerConfirmedAt,
			&result.Status, &result.ApprovalTrail,
			&result.RejectionReason,
			&result.CancelledBy, &result.CancelledAt,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		r = &result
		return nil
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("permintaan shift tidak ditemukan atau sudah diproses")
		}
		return nil, err
	}
	return r, nil
}

func CancelShiftChangeRequest(ctx context.Context, id, employeeID string) error {
	return database.WithUserContext(ctx, employeeID, func(tx pgx.Tx) error {
		query := `UPDATE shift_change_requests SET status = 'cancelled'::shift_change_status,
			cancelled_by = $2::uuid, cancelled_at = NOW()
			WHERE id::text = $1 AND deleted_at IS NULL AND employee_id::text = $2
			AND status IN ('pending', 'partner_pending')`
		tag, err := tx.Exec(ctx, query, id, employeeID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("permintaan shift tidak ditemukan atau sudah diproses")
		}
		return nil
	})
}

func CheckShiftChangeDuplicatePending(ctx context.Context, employeeID, targetDate string) (bool, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM shift_change_requests
		WHERE employee_id::text = $1 AND target_date::text = $2
		AND status IN ('pending', 'partner_pending') AND deleted_at IS NULL`,
		employeeID, targetDate,
	).Scan(&count)
	return count > 0, err
}
