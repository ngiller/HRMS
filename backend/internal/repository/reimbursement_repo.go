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

func ListReimbursements(ctx context.Context, page, perPage int, status, employeeID string) ([]models.ReimbursementSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM reimbursements r WHERE r.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND r.status = $%d::leave_status", argIdx)
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
	argIdx++
	query := fmt.Sprintf(`
		SELECT r.id, r.employee_id, COALESCE(e.full_name, ''),
			r.type::text, r.amount, COALESCE(r.description, ''),
			r.status::text, r.created_at
		FROM reimbursements r
		LEFT JOIN employees e ON e.id = r.employee_id
		WHERE r.deleted_at IS NULL
	`)
	searchArgs := []interface{}{}
	if status != "" {
		query += fmt.Sprintf(" AND r.status = $%d::leave_status", len(searchArgs)+1)
		searchArgs = append(searchArgs, status)
	}
	if employeeID != "" {
		query += fmt.Sprintf(" AND r.employee_id::text = $%d", len(searchArgs)+1)
		searchArgs = append(searchArgs, employeeID)
	}
	query += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	allArgs := append(searchArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reimbursements []models.ReimbursementSummary
	for rows.Next() {
		var r models.ReimbursementSummary
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Type, &r.Amount, &r.Description,
			&r.Status, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		reimbursements = append(reimbursements, r)
	}
	return reimbursements, total, nil
}

func GetReimbursementByID(ctx context.Context, id string) (*models.Reimbursement, error) {
	query := `
		SELECT r.id, r.employee_id, COALESCE(e.full_name, ''),
			r.type::text, r.amount, COALESCE(r.description, ''),
			COALESCE(r.receipt_urls, '{}'), COALESCE(r.approval_trail::text, '[]'),
			r.status::text, COALESCE(r.payment_method, ''),
			r.paid_at, r.paid_by, COALESCE(payer.full_name, ''),
			COALESCE(r.rejection_reason, ''),
			r.cancelled_by, r.cancelled_at,
			r.created_at, r.updated_at, r.deleted_at
		FROM reimbursements r
		LEFT JOIN employees e ON e.id = r.employee_id
		LEFT JOIN employees payer ON payer.id = r.paid_by
		WHERE (r.id::text = $1) AND r.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.Reimbursement
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.Type, &r.Amount, &r.Description,
		&r.ReceiptUrls, &r.ApprovalTrail,
		&r.Status, &r.PaymentMethod,
		&r.PaidAt, &r.PaidBy, &r.PaidByName,
		&r.RejectionReason,
		&r.CancelledBy, &r.CancelledAt,
		&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateReimbursement(ctx context.Context, employeeID, userID string, req *models.CreateReimbursementReq) (*models.Reimbursement, error) {
	query := `
		INSERT INTO reimbursements (employee_id, type, amount, description, receipt_urls, status)
		VALUES ($1::uuid, $2::reimbursement_type, $3, $4, $5, 'pending'::leave_status)
		RETURNING id, employee_id, '' as employee_name,
			type::text, amount, COALESCE(description, ''),
			COALESCE(receipt_urls, '{}'), COALESCE(approval_trail::text, '[]'),
			status::text, COALESCE(payment_method, ''),
			paid_at, paid_by, '' as paid_by_name,
			COALESCE(rejection_reason, ''),
			cancelled_by, cancelled_at,
			created_at, updated_at, deleted_at
	`
	var r models.Reimbursement
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			employeeID, req.Type, req.Amount, req.Description, req.ReceiptUrls,
		)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Type, &r.Amount, &r.Description,
			&r.ReceiptUrls, &r.ApprovalTrail,
			&r.Status, &r.PaymentMethod,
			&r.PaidAt, &r.PaidBy, &r.PaidByName,
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

func UpdateReimbursementStatus(ctx context.Context, id, status, rejectionReason, approverID string) (*models.Reimbursement, error) {
	query := `
		UPDATE reimbursements
		SET status = $2::leave_status,
			rejection_reason = $3,
			approval_trail = approval_trail || $4::jsonb
		WHERE id::text = $1 AND deleted_at IS NULL AND status = 'pending'
		RETURNING id, employee_id, '' as employee_name,
			type::text, amount, COALESCE(description, ''),
			COALESCE(receipt_urls, '{}'), COALESCE(approval_trail::text, '[]'),
			status::text, COALESCE(payment_method, ''),
			paid_at, paid_by, '' as paid_by_name,
			COALESCE(rejection_reason, ''),
			cancelled_by, cancelled_at,
			created_at, updated_at, deleted_at
	`
	approvalEntry := fmt.Sprintf(`[{"status":"%s","approver_id":"%s","date":"%s"}]`, status, approverID, time.Now().Format(time.RFC3339))
	var r models.Reimbursement
	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, status, rejectionReason, approvalEntry)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Type, &r.Amount, &r.Description,
			&r.ReceiptUrls, &r.ApprovalTrail,
			&r.Status, &r.PaymentMethod,
			&r.PaidAt, &r.PaidBy, &r.PaidByName,
			&r.RejectionReason,
			&r.CancelledBy, &r.CancelledAt,
			&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("reimbursement tidak ditemukan atau sudah diproses")
		}
		return nil, err
	}
	return &r, nil
}

func PayReimbursement(ctx context.Context, id, paidBy, paymentMethod string) (*models.Reimbursement, error) {
	query := `
		UPDATE reimbursements
		SET status = 'paid'::leave_status,
			payment_method = $2,
			paid_at = NOW(),
			paid_by = $3::uuid,
			approval_trail = approval_trail || $4::jsonb
		WHERE id::text = $1 AND deleted_at IS NULL AND status = 'approved'
		RETURNING id, employee_id, '' as employee_name,
			type::text, amount, COALESCE(description, ''),
			COALESCE(receipt_urls, '{}'), COALESCE(approval_trail::text, '[]'),
			status::text, COALESCE(payment_method, ''),
			paid_at, paid_by, '' as paid_by_name,
			COALESCE(rejection_reason, ''),
			cancelled_by, cancelled_at,
			created_at, updated_at, deleted_at
	`
	approvalEntry := fmt.Sprintf(`[{"status":"paid","paid_by":"%s","date":"%s"}]`, paidBy, time.Now().Format(time.RFC3339))
	var r models.Reimbursement
	err := database.WithUserContext(ctx, paidBy, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, paymentMethod, paidBy, approvalEntry)
		return row.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Type, &r.Amount, &r.Description,
			&r.ReceiptUrls, &r.ApprovalTrail,
			&r.Status, &r.PaymentMethod,
			&r.PaidAt, &r.PaidBy, &r.PaidByName,
			&r.RejectionReason,
			&r.CancelledBy, &r.CancelledAt,
			&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("reimbursement tidak ditemukan atau belum disetujui")
		}
		return nil, err
	}
	return &r, nil
}

func CancelReimbursement(ctx context.Context, id, employeeID string) error {
	query := `UPDATE reimbursements SET status = 'cancelled'::leave_status,
		cancelled_by = $2::uuid, cancelled_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL AND employee_id::text = $2
		AND status IN ('pending', 'approved')`
	return database.WithUserContext(ctx, employeeID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, id, employeeID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("reimbursement tidak ditemukan atau sudah diproses")
		}
		return nil
	})
}
