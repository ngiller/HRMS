package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

// ─── Leave Types ──────────────────────────────────────────────

func GetAllLeaveTypes(ctx context.Context) ([]models.LeaveTypeSummary, error) {
	query := `SELECT id, name, code, default_quota, is_paid, is_active
		FROM leave_types WHERE deleted_at IS NULL ORDER BY sort_order ASC`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []models.LeaveTypeSummary
	for rows.Next() {
		var t models.LeaveTypeSummary
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.DefaultQuota, &t.IsPaid, &t.IsActive); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	if types == nil {
		types = []models.LeaveTypeSummary{}
	}
	return types, nil
}

func GetLeaveTypeByID(ctx context.Context, id string) (*models.LeaveTypeSummary, error) {
	query := `SELECT id, name, code, default_quota, is_paid, is_active
		FROM leave_types WHERE id::text = $1 AND deleted_at IS NULL`
	var t models.LeaveTypeSummary
	err := database.Pool.QueryRow(ctx, query, id).Scan(&t.ID, &t.Name, &t.Code, &t.DefaultQuota, &t.IsPaid, &t.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// ─── Leave Balances ───────────────────────────────────────────

func GetLeaveBalances(ctx context.Context, employeeID string, year int) ([]models.LeaveBalance, error) {
	query := `SELECT lb.id, lb.employee_id, lb.leave_type_id, lb.year,
		CASE
			WHEN lt.code = 'tahunan' AND lb.year >= EXTRACT(YEAR FROM NOW())::int THEN
				LEAST(COALESCE(lt.default_quota, 12), EXTRACT(MONTH FROM NOW())::int)
			WHEN lt.code = 'tahunan' AND lb.year < EXTRACT(YEAR FROM NOW())::int THEN
				COALESCE(lt.default_quota, 12)
			ELSE lb.total_quota
		END as total_quota,
		lb.used,
		CASE
			WHEN lt.code = 'tahunan' AND lb.year >= EXTRACT(YEAR FROM NOW())::int THEN
				LEAST(COALESCE(lt.default_quota, 12), EXTRACT(MONTH FROM NOW())::int)
				- lb.used + COALESCE(lb.rolled_over_from, 0)
			WHEN lt.code = 'tahunan' AND lb.year < EXTRACT(YEAR FROM NOW())::int THEN
				COALESCE(lt.default_quota, 12) - lb.used + COALESCE(lb.rolled_over_from, 0)
			ELSE lb.remaining
		END as remaining,
		lb.rolled_over_from,
		COALESCE(lt.name, ''), COALESCE(lt.code, '')
		FROM leave_balances lb
		LEFT JOIN leave_types lt ON lt.id = lb.leave_type_id
		WHERE lb.employee_id::text = $1 AND lb.year = $2
		ORDER BY lt.sort_order ASC`
	rows, err := database.Pool.Query(ctx, query, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []models.LeaveBalance
	for rows.Next() {
		var b models.LeaveBalance
		if err := rows.Scan(&b.ID, &b.EmployeeID, &b.LeaveTypeID, &b.Year,
			&b.TotalQuota, &b.Used, &b.Remaining, &b.RolledOverFrom,
			&b.LeaveTypeName, &b.LeaveTypeCode); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	if balances == nil {
		balances = []models.LeaveBalance{}
	}
	return balances, nil
}

func GetLeaveBalancesByEmployee(ctx context.Context, employeeID string) ([]models.LeaveBalance, error) {
	return GetLeaveBalances(ctx, employeeID, time.Now().Year())
}

func GetAllLeaveBalances(ctx context.Context, year int) ([]models.LeaveBalance, error) {
	query := `SELECT lb.id, lb.employee_id, lb.leave_type_id, lb.year,
		CASE
			WHEN lt.code = 'tahunan' AND lb.year >= EXTRACT(YEAR FROM NOW())::int THEN
				LEAST(COALESCE(lt.default_quota, 12), EXTRACT(MONTH FROM NOW())::int)
			WHEN lt.code = 'tahunan' AND lb.year < EXTRACT(YEAR FROM NOW())::int THEN
				COALESCE(lt.default_quota, 12)
			ELSE lb.total_quota
		END as total_quota,
		lb.used,
		CASE
			WHEN lt.code = 'tahunan' AND lb.year >= EXTRACT(YEAR FROM NOW())::int THEN
				LEAST(COALESCE(lt.default_quota, 12), EXTRACT(MONTH FROM NOW())::int)
				- lb.used + COALESCE(lb.rolled_over_from, 0)
			WHEN lt.code = 'tahunan' AND lb.year < EXTRACT(YEAR FROM NOW())::int THEN
				COALESCE(lt.default_quota, 12) - lb.used + COALESCE(lb.rolled_over_from, 0)
			ELSE lb.remaining
		END as remaining,
		lb.rolled_over_from,
		COALESCE(lt.name, ''), COALESCE(lt.code, '')
		FROM leave_balances lb
		LEFT JOIN leave_types lt ON lt.id = lb.leave_type_id
		WHERE lb.year = $1
		ORDER BY lb.employee_id, lt.sort_order ASC`
	rows, err := database.Pool.Query(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []models.LeaveBalance
	for rows.Next() {
		var b models.LeaveBalance
		if err := rows.Scan(&b.ID, &b.EmployeeID, &b.LeaveTypeID, &b.Year,
			&b.TotalQuota, &b.Used, &b.Remaining, &b.RolledOverFrom,
			&b.LeaveTypeName, &b.LeaveTypeCode); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	if balances == nil {
		balances = []models.LeaveBalance{}
	}
	return balances, nil
}

// ─── Leave Requests ───────────────────────────────────────────

func ListLeaveRequests(ctx context.Context, page, perPage int, status, employeeID string) ([]models.LeaveRequestSummary, int, error) {
	args := []interface{}{}
	argIdx := 0

	whereClause := `WHERE lr.deleted_at IS NULL`
	if status != "" {
		argIdx++
		whereClause += fmt.Sprintf(" AND lr.status = $%d::leave_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		whereClause += fmt.Sprintf(" AND lr.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM leave_requests lr %s`, whereClause)
	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data
	offset := (page - 1) * perPage
	argIdx++
	dataQuery := fmt.Sprintf(`
		SELECT lr.id, lr.employee_id, COALESCE(e.full_name, ''),
			COALESCE(lt.name, ''), lr.start_date::text, lr.end_date::text,
			lr.total_days, lr.is_half_day, COALESCE(lr.reason, ''),
			lr.status::text, lr.created_at
		FROM leave_requests lr
		LEFT JOIN employees e ON e.id = lr.employee_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		%s
		ORDER BY lr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	allArgs := append(args, perPage, offset)
	rows, err := database.Pool.Query(ctx, dataQuery, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []models.LeaveRequestSummary
	for rows.Next() {
		var r models.LeaveRequestSummary
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.LeaveTypeName, &r.StartDate, &r.EndDate,
			&r.TotalDays, &r.IsHalfDay, &r.Reason,
			&r.Status, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		requests = append(requests, r)
	}
	if requests == nil {
		requests = []models.LeaveRequestSummary{}
	}
	return requests, total, nil
}

func GetLeaveRequestByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	query := `
		SELECT lr.id, lr.employee_id, COALESCE(e.full_name, ''),
			lr.leave_type_id, COALESCE(lt.name, ''),
			lr.start_date::text, lr.end_date::text,
			lr.total_days, lr.is_half_day, COALESCE(lr.reason, ''),
			COALESCE(lr.document_url, ''), COALESCE(lr.contact_during_leave, ''),
			COALESCE(lr.approval_trail::text, '[]'), lr.status::text,
			lr.cancelled_at, COALESCE(lr.cancel_reason, ''),
			lr.created_at, lr.updated_at, lr.deleted_at
		FROM leave_requests lr
		LEFT JOIN employees e ON e.id = lr.employee_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE (lr.id::text = $1) AND lr.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.LeaveRequest
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.LeaveTypeID, &r.LeaveTypeName,
		&r.StartDate, &r.EndDate,
		&r.TotalDays, &r.IsHalfDay, &r.Reason,
		&r.DocumentURL, &r.ContactDuringLeave,
		&r.ApprovalTrail, &r.Status,
		&r.CancelledAt, &r.CancelReason,
		&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateLeaveRequest(ctx context.Context, employeeID, userID string, req *models.CreateLeaveRequestReq) (*models.LeaveRequest, error) {
	var r *models.LeaveRequest
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO leave_requests (employee_id, leave_type_id, start_date, end_date,
				total_days, is_half_day, reason, document_url, contact_during_leave, status)
			VALUES ($1::uuid, $2::uuid, $3::date, $4::date,
				$5, $6, $7, $8, $9, 'pending'::leave_status)
			RETURNING id, employee_id, '' as employee_name,
				leave_type_id, '' as leave_type_name,
				start_date::text, end_date::text,
				total_days, is_half_day, COALESCE(reason, ''),
				COALESCE(document_url, ''), COALESCE(contact_during_leave, ''),
				COALESCE(approval_trail::text, '[]'), status::text,
				cancelled_at, COALESCE(cancel_reason, ''),
				created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, query,
			employeeID, req.LeaveTypeID, req.StartDate, req.EndDate,
			req.TotalDays, req.IsHalfDay, req.Reason, req.DocumentURL, req.ContactDuringLeave,
		)
		var result models.LeaveRequest
		if err := row.Scan(
			&result.ID, &result.EmployeeID, &result.EmployeeName,
			&result.LeaveTypeID, &result.LeaveTypeName,
			&result.StartDate, &result.EndDate,
			&result.TotalDays, &result.IsHalfDay, &result.Reason,
			&result.DocumentURL, &result.ContactDuringLeave,
			&result.ApprovalTrail, &result.Status,
			&result.CancelledAt, &result.CancelReason,
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

func UpdateLeaveStatus(ctx context.Context, id, status, rejectionReason, approverID string) (*models.LeaveRequest, error) {
	approvalEntry := fmt.Sprintf(`[{"status":"%s","approver_id":"%s","date":"%s"}]`, status, approverID, time.Now().Format(time.RFC3339))

	var r models.LeaveRequest

	// Gunakan transaksi agar update status dan update balance bersifat atomik.
	// Jika update balance gagal → status approved ikut di-rollback.
	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		// 1. Update status leave_request (hanya jika masih 'pending' — optimistic locking)
		updateQuery := `
			UPDATE leave_requests
			SET status = $2::leave_status,
				rejection_reason = $3,
				approval_trail = approval_trail || $4::jsonb,
				rejected_at = CASE WHEN $2::leave_status = 'rejected' THEN NOW() ELSE rejected_at END,
				rejected_by = CASE WHEN $2::leave_status = 'rejected' THEN $5::uuid ELSE rejected_by END
			WHERE id::text = $1 AND deleted_at IS NULL AND status = 'pending'
			RETURNING id, employee_id, '' as employee_name,
				leave_type_id, '' as leave_type_name,
				start_date::text, end_date::text,
				total_days, is_half_day, COALESCE(reason, ''),
				COALESCE(document_url, ''), COALESCE(contact_during_leave, ''),
				COALESCE(approval_trail::text, '[]'), status::text,
				cancelled_at, COALESCE(cancel_reason, ''),
				created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, updateQuery, id, status, rejectionReason, approvalEntry, approverID)
		if err := row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.LeaveTypeID, &r.LeaveTypeName,
			&r.StartDate, &r.EndDate,
			&r.TotalDays, &r.IsHalfDay, &r.Reason,
			&r.DocumentURL, &r.ContactDuringLeave,
			&r.ApprovalTrail, &r.Status,
			&r.CancelledAt, &r.CancelReason,
			&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
		); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("pengajuan cuti tidak ditemukan atau sudah diproses")
			}
			return fmt.Errorf("update leave_requests: %w", err)
		}

		// 2. Jika approved → kurangi jatah cuti (update kolom `used` saja;
		//    kolom `remaining` dihitung dinamis di query SELECT via CASE WHEN).
		if status == "approved" {
			tag, err := tx.Exec(ctx, `
				UPDATE leave_balances
				SET used = used + $1, updated_at = NOW()
				WHERE employee_id = $2::uuid
					AND leave_type_id = $3::uuid
					AND year = EXTRACT(YEAR FROM $4::date)::int
			`, r.TotalDays, r.EmployeeID, r.LeaveTypeID, r.StartDate)
			if err != nil {
				return fmt.Errorf("update leave_balances: %w", err)
			}
			if tag.RowsAffected() == 0 {
				fmt.Printf("[WARN] UpdateLeaveStatus: tidak ada baris leave_balances yang terupdate (employee_id=%s, leave_type_id=%s, year dari start_date=%s)\n",
					r.EmployeeID, r.LeaveTypeID, r.StartDate)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func CancelLeaveRequest(ctx context.Context, id, employeeID, cancelReason string) error {
	query := `UPDATE leave_requests SET status = 'cancelled'::leave_status,
		cancelled_at = NOW(), cancel_reason = $3
		WHERE id::text = $1 AND deleted_at IS NULL AND employee_id::text = $2
		AND status = 'pending'`
	return database.WithUserContext(ctx, employeeID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, id, employeeID, cancelReason)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("pengajuan cuti tidak ditemukan atau sudah diproses")
		}
		return nil
	})
}
