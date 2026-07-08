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

func ListLoans(ctx context.Context, page, perPage int, status, employeeID string) ([]models.LoanSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM loans l WHERE l.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND l.status = $%d::loan_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND l.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	// Build main query with independent arg index
	listQuery := `
		SELECT l.id, l.employee_id, COALESCE(e.full_name, ''),
			l.loan_type::text, l.amount, l.installment_count,
			l.installment_amount, l.remaining_balance,
			l.status::text, l.created_at
		FROM loans l
		LEFT JOIN employees e ON e.id = l.employee_id
		WHERE l.deleted_at IS NULL
	`
	listArgs := []interface{}{}
	listIdx := 0
	if status != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND l.status = $%d::loan_status", listIdx)
		listArgs = append(listArgs, status)
	}
	if employeeID != "" {
		listIdx++
		listQuery += fmt.Sprintf(" AND l.employee_id::text = $%d", listIdx)
		listArgs = append(listArgs, employeeID)
	}
	listQuery += fmt.Sprintf(" ORDER BY l.created_at DESC LIMIT $%d OFFSET $%d", listIdx+1, listIdx+2)
	allArgs := append(listArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, listQuery, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var loans []models.LoanSummary
	for rows.Next() {
		var l models.LoanSummary
		if err := rows.Scan(&l.ID, &l.EmployeeID, &l.EmployeeName,
			&l.LoanType, &l.Amount, &l.InstallmentCount,
			&l.InstallmentAmount, &l.RemainingBalance,
			&l.Status, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		loans = append(loans, l)
	}
	return loans, total, nil
}

func GetLoanByID(ctx context.Context, id string) (*models.Loan, error) {
	query := `
		SELECT l.id, l.employee_id, COALESCE(e.full_name, ''),
			l.loan_type::text, l.amount, l.interest_rate,
			COALESCE(l.total_interest, 0), l.total_amount,
			l.installment_count, l.installment_amount,
			l.payment_method::text, l.remaining_balance,
			COALESCE(l.purpose, ''), COALESCE(l.approval_trail::text, '[]'),
			l.status::text, l.disbursed_at, l.disbursed_by,
			l.settled_at, l.created_at, l.updated_at, l.deleted_at
		FROM loans l
		LEFT JOIN employees e ON e.id = l.employee_id
		WHERE (l.id::text = $1) AND l.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var l models.Loan
	err := row.Scan(
		&l.ID, &l.EmployeeID, &l.EmployeeName,
		&l.LoanType, &l.Amount, &l.InterestRate,
		&l.TotalInterest, &l.TotalAmount,
		&l.InstallmentCount, &l.InstallmentAmount,
		&l.PaymentMethod, &l.RemainingBalance,
		&l.Purpose, &l.ApprovalTrail,
		&l.Status, &l.DisbursedAt, &l.DisbursedBy,
		&l.SettledAt, &l.CreatedAt, &l.UpdatedAt, &l.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func CreateLoan(ctx context.Context, employeeID string, req *models.CreateLoanRequest, userID string) (*models.Loan, error) {
	totalAmount := req.Amount
	var totalInterest float64
	if req.InterestRate > 0 {
		totalInterest = req.Amount * req.InterestRate / 100
		totalAmount = req.Amount + totalInterest
	}
	installmentAmount := totalAmount / float64(req.InstallmentCount)

	query := `
		INSERT INTO loans (employee_id, loan_type, amount, interest_rate,
			total_interest, total_amount, installment_count, installment_amount,
			payment_method, remaining_balance, purpose, status)
		VALUES ($1::uuid, $2::loan_type, $3, $4,
			$5, $6, $7, $8,
			$9::loan_payment_method, $6, $10, 'pending'::loan_status)
		RETURNING id, employee_id, '' as employee_name,
			loan_type::text, amount, interest_rate,
			COALESCE(total_interest, 0), total_amount,
			installment_count, installment_amount,
			payment_method::text, remaining_balance,
			COALESCE(purpose, ''), COALESCE(approval_trail::text, '[]'),
			status::text, disbursed_at, disbursed_by,
			settled_at, created_at, updated_at, deleted_at
	`
	var l models.Loan
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.EmployeeID, req.LoanType, req.Amount, req.InterestRate,
			totalInterest, totalAmount, req.InstallmentCount, installmentAmount,
			req.PaymentMethod, req.Purpose,
		)
		return row.Scan(
			&l.ID, &l.EmployeeID, &l.EmployeeName,
			&l.LoanType, &l.Amount, &l.InterestRate,
			&l.TotalInterest, &l.TotalAmount,
			&l.InstallmentCount, &l.InstallmentAmount,
			&l.PaymentMethod, &l.RemainingBalance,
			&l.Purpose, &l.ApprovalTrail,
			&l.Status, &l.DisbursedAt, &l.DisbursedBy,
			&l.SettledAt, &l.CreatedAt, &l.UpdatedAt, &l.DeletedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func UpdateLoanStatus(ctx context.Context, id, status, rejectionReason, approverID string) (*models.Loan, error) {
	query := `
		UPDATE loans
		SET status = $2::loan_status,
			approval_trail = approval_trail || $3::jsonb
		WHERE id::text = $1 AND deleted_at IS NULL
		AND status = 'pending'
		RETURNING id, employee_id, '' as employee_name,
			loan_type::text, amount, interest_rate,
			COALESCE(total_interest, 0), total_amount,
			installment_count, installment_amount,
			payment_method::text, remaining_balance,
			COALESCE(purpose, ''), COALESCE(approval_trail::text, '[]'),
			status::text, disbursed_at, disbursed_by,
			settled_at, created_at, updated_at, deleted_at
	`
	approvalEntry := fmt.Sprintf(`[{"status":"%s","approver_id":"%s","date":"%s"}]`,
		status, approverID, time.Now().UTC().Format(time.RFC3339))
	var l models.Loan
	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, status, approvalEntry)
		return row.Scan(
			&l.ID, &l.EmployeeID, &l.EmployeeName,
			&l.LoanType, &l.Amount, &l.InterestRate,
			&l.TotalInterest, &l.TotalAmount,
			&l.InstallmentCount, &l.InstallmentAmount,
			&l.PaymentMethod, &l.RemainingBalance,
			&l.Purpose, &l.ApprovalTrail,
			&l.Status, &l.DisbursedAt, &l.DisbursedBy,
			&l.SettledAt, &l.CreatedAt, &l.UpdatedAt, &l.DeletedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("pinjaman tidak ditemukan atau sudah diproses")
		}
		return nil, err
	}
	return &l, nil
}

func DisburseLoan(ctx context.Context, id, approverID, date string) (*models.Loan, error) {
	query := `
		UPDATE loans
		SET status = 'active'::loan_status,
			disbursed_at = $2::date,
			disbursed_by = $3::uuid,
			approval_trail = approval_trail || $4::jsonb
		WHERE id::text = $1 AND deleted_at IS NULL
		AND status IN ('approved', 'pending')
		RETURNING id, employee_id, '' as employee_name,
			loan_type::text, amount, interest_rate,
			COALESCE(total_interest, 0), total_amount,
			installment_count, installment_amount,
			payment_method::text, remaining_balance,
			COALESCE(purpose, ''), COALESCE(approval_trail::text, '[]'),
			status::text, disbursed_at, disbursed_by,
			settled_at, created_at, updated_at, deleted_at
	`
	approvalEntry := fmt.Sprintf(`[{"status":"disbursed","approver_id":"%s","date":"%s"}]`,
		approverID, time.Now().UTC().Format(time.RFC3339))
	var l models.Loan
	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, date, approverID, approvalEntry)
		return row.Scan(
			&l.ID, &l.EmployeeID, &l.EmployeeName,
			&l.LoanType, &l.Amount, &l.InterestRate,
			&l.TotalInterest, &l.TotalAmount,
			&l.InstallmentCount, &l.InstallmentAmount,
			&l.PaymentMethod, &l.RemainingBalance,
			&l.Purpose, &l.ApprovalTrail,
			&l.Status, &l.DisbursedAt, &l.DisbursedBy,
			&l.SettledAt, &l.CreatedAt, &l.UpdatedAt, &l.DeletedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("pinjaman tidak ditemukan atau sudah dicairkan")
		}
		return nil, err
	}
	return &l, nil
}

func CancelLoan(ctx context.Context, id, employeeID string) error {
	query := `
		UPDATE loans
		SET status = 'cancelled'::loan_status,
			deleted_at = NOW()
		WHERE id::text = $1 AND employee_id::text = $2
		AND deleted_at IS NULL AND status = 'pending'::loan_status
	`
	_, err := database.Pool.Exec(ctx, query, id, employeeID)
	if err != nil {
		return fmt.Errorf("gagal membatalkan pinjaman: %w", err)
	}
	return nil
}

func GetLoanStats(ctx context.Context) (*models.LoanStatsResponse, error) {
	stats := &models.LoanStatsResponse{}

	// Total active loans with outstanding balance
	err := database.Pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(remaining_balance), 0)
		FROM loans WHERE status IN ('approved','active') AND deleted_at IS NULL
	`).Scan(&stats.ActiveLoans, &stats.TotalOutstanding)
	if err != nil {
		return nil, err
	}

	// Total loans and disbursed amount
	err = database.Pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM loans WHERE deleted_at IS NULL
	`).Scan(&stats.TotalLoans, &stats.TotalDisbursed)
	if err != nil {
		return nil, err
	}

	// By status
	stats.ByStatus = make(map[string]int)
	rows, err := database.Pool.Query(ctx, `
		SELECT status::text, COUNT(*) FROM loans WHERE deleted_at IS NULL GROUP BY status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		var c int
		if err := rows.Scan(&s, &c); err != nil {
			return nil, err
		}
		stats.ByStatus[s] = c
	}

	// By payment method
	rows2, err := database.Pool.Query(ctx, `
		SELECT payment_method::text, COUNT(*), COALESCE(SUM(amount), 0)
		FROM loans WHERE deleted_at IS NULL GROUP BY payment_method
	`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var pm models.PaymentMethodStat
		if err := rows2.Scan(&pm.Method, &pm.Count, &pm.Total); err != nil {
			return nil, err
		}
		stats.ByMethod = append(stats.ByMethod, pm)
	}

	return stats, nil
}
