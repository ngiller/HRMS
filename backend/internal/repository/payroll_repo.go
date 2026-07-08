package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ==================== Payroll Periods ====================

func ListPayrollPeriods(ctx context.Context, page, perPage int) ([]models.PayrollPeriodSummary, int, error) {
	var total int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM payroll_periods WHERE deleted_at IS NULL`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT pp.id, pp.month, pp.year, pp.period_name,
			pp.start_date, pp.end_date, pp.status::text,
			pp.total_employee, pp.total_gross, pp.total_net,
			COALESCE((SELECT full_name FROM employees WHERE id = pp.approved_by), ''),
			COALESCE((SELECT full_name FROM employees WHERE id = pp.paid_by), ''),
			pp.created_at
		FROM payroll_periods pp
		WHERE pp.deleted_at IS NULL
		ORDER BY pp.year DESC, pp.month DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := database.Pool.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var periods []models.PayrollPeriodSummary
	for rows.Next() {
		var p models.PayrollPeriodSummary
		var status string
		if err := rows.Scan(&p.ID, &p.Month, &p.Year, &p.PeriodName,
			&p.StartDate, &p.EndDate, &status,
			&p.TotalEmployee, &p.TotalGross, &p.TotalNet,
			&p.ApprovedByName, &p.PaidByName,
			&p.CreatedAt); err != nil {
			return nil, 0, err
		}
		p.Status = models.PayrollPeriodStatus(status)
		periods = append(periods, p)
	}

	return periods, total, nil
}

func GetPayrollPeriod(ctx context.Context, periodID string) (*models.PayrollPeriod, error) {
	query := `
		SELECT id, month, year, period_name, start_date, end_date, status::text,
			approved_by, approved_at, paid_by, paid_at,
			total_employee, total_gross, total_deductions, total_net, total_company_cost,
			created_by, created_at, updated_at
		FROM payroll_periods
		WHERE id::text = $1 AND deleted_at IS NULL
	`

	row := database.Pool.QueryRow(ctx, query, periodID)
	var p models.PayrollPeriod
	var status string
	err := row.Scan(&p.ID, &p.Month, &p.Year, &p.PeriodName, &p.StartDate, &p.EndDate, &status,
		&p.ApprovedBy, &p.ApprovedAt, &p.PaidBy, &p.PaidAt,
		&p.TotalEmployee, &p.TotalGross, &p.TotalDeductions, &p.TotalNet, &p.TotalCompanyCost,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.Status = models.PayrollPeriodStatus(status)
	return &p, nil
}

func CreatePayrollPeriod(ctx context.Context, req *models.CreatePayrollPeriodRequest, userID string) (*models.PayrollPeriod, error) {
	var startDate, endDate string
	var query string
	if req.StartDate != "" && req.EndDate != "" {
		query = `
			INSERT INTO payroll_periods (month, year, period_name, start_date, end_date, created_by)
			VALUES ($1, $2, $3, $5::date, $6::date, $4::uuid)
			RETURNING id, month, year, period_name, start_date, end_date, status::text,
				approved_by, approved_at, paid_by, paid_at,
				total_employee, total_gross, total_deductions, total_net, total_company_cost,
				created_by, created_at, updated_at
		`
		startDate = req.StartDate
		endDate = req.EndDate
	} else {
		query = `
			INSERT INTO payroll_periods (month, year, period_name, start_date, end_date, created_by)
			VALUES ($1, $2, $3,
				MAKE_DATE($2, $1, 1),
				(MAKE_DATE($2, $1, 1) + INTERVAL '1 month' - INTERVAL '1 day')::date,
				$4::uuid)
			RETURNING id, month, year, period_name, start_date, end_date, status::text,
				approved_by, approved_at, paid_by, paid_at,
				total_employee, total_gross, total_deductions, total_net, total_company_cost,
				created_by, created_at, updated_at
		`
	}

	var p models.PayrollPeriod
	var status string
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		var args []interface{}
		if startDate != "" && endDate != "" {
			args = []interface{}{req.Month, req.Year, req.PeriodName, userID, startDate, endDate}
		} else {
			args = []interface{}{req.Month, req.Year, req.PeriodName, userID}
		}
		return tx.QueryRow(ctx, query, args...).Scan(
			&p.ID, &p.Month, &p.Year, &p.PeriodName, &p.StartDate, &p.EndDate, &status,
			&p.ApprovedBy, &p.ApprovedAt, &p.PaidBy, &p.PaidAt,
			&p.TotalEmployee, &p.TotalGross, &p.TotalDeductions, &p.TotalNet, &p.TotalCompanyCost,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	})
	if err != nil {
		return nil, err
	}
	p.Status = models.PayrollPeriodStatus(status)
	return &p, nil
}

func UpdatePayrollPeriodStatus(ctx context.Context, periodID string, status models.PayrollPeriodStatus, userID string) error {
	var col string
	switch status {
	case models.PayrollStatusCalculated:
		col = "status = 'calculated'::payroll_status"
	case models.PayrollStatusApproved:
		col = "status = 'approved'::payroll_status, approved_by = $2::uuid, approved_at = NOW()"
	case models.PayrollStatusPaid:
		col = "status = 'paid'::payroll_status, paid_by = $2::uuid, paid_at = NOW()"
	default:
		return errors.New("status tidak valid")
	}

	query := fmt.Sprintf(`UPDATE payroll_periods SET %s, updated_at = NOW() WHERE id::text = $1`, col)
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		var tag pgconn.CommandTag
		var err error
		if status == models.PayrollStatusCalculated {
			tag, err = tx.Exec(ctx, query, periodID)
		} else {
			tag, err = tx.Exec(ctx, query, periodID, userID)
		}
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("periode penggajian tidak ditemukan")
		}
		return nil
	})
}

// ==================== Payroll Items ====================

func GetActiveSalaryComponents(ctx context.Context, employeeID string) (allowances []models.AllowanceItem, deductions []models.AllowanceItem, err error) {
	query := `
		SELECT component_name, amount, component_type::text
		FROM employee_salary_components
		WHERE employee_id::text = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY component_name
	`

	rows, err := database.Pool.Query(ctx, query, employeeID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.AllowanceItem
		var compType string
		if err := rows.Scan(&item.Name, &item.Amount, &compType); err != nil {
			return nil, nil, err
		}
		if compType == "allowance" {
			allowances = append(allowances, item)
		} else {
			deductions = append(deductions, item)
		}
	}
	if allowances == nil {
		allowances = []models.AllowanceItem{}
	}
	if deductions == nil {
		deductions = []models.AllowanceItem{}
	}
	return allowances, deductions, nil
}

func GetLatestBaseSalary(ctx context.Context, employeeID string) (float64, error) {
	var salary float64
	err := database.Pool.QueryRow(ctx, `
		SELECT base_salary FROM employee_salary_histories
		WHERE employee_id::text = $1
		ORDER BY effective_date DESC NULLS LAST, created_at DESC
		LIMIT 1
	`, employeeID).Scan(&salary)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return salary, nil
}

func GetEmployeeDetailsForPayroll(ctx context.Context, employeeID string) (dailyWage float64, employmentStatus string, err error) {
	err = database.Pool.QueryRow(ctx, `
		SELECT COALESCE(daily_wage, 0), COALESCE(employment_status::text, '')
		FROM employees WHERE id::text = $1 AND deleted_at IS NULL
	`, employeeID).Scan(&dailyWage, &employmentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", errors.New("karyawan tidak ditemukan")
		}
		return 0, "", err
	}
	return dailyWage, employmentStatus, nil
}

// CallCalculateEmployeePayroll calls the PL/pgSQL function calculate_employee_payroll()
// to calculate and upsert a payroll item for a single employee.
// Wrapped in WithUserContext so audit triggers have the real user ID.
func CallCalculateEmployeePayroll(
	ctx context.Context,
	userID string,
	periodID string,
	employeeID string,
	baseSalary float64,
	dailyWage float64,
	totalDaysWorked int,
	allowances []models.AllowanceItem,
	overtimePay float64,
	thrAmount float64,
	bonusAmount float64,
	loanDeduction float64,
	otherDeductions float64,
) (string, error) {
	allowancesJSON, _ := json.Marshal(allowances)

	var resultID string
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT calculate_employee_payroll(
				$1::uuid, $2::uuid, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11
			)::text
		`, periodID, employeeID, baseSalary, dailyWage, totalDaysWorked,
			string(allowancesJSON), overtimePay, thrAmount, bonusAmount, loanDeduction, otherDeductions,
		).Scan(&resultID)
	})
	if err != nil {
		return "", err
	}
	return resultID, nil
}

func ListPayrollItems(ctx context.Context, periodID string, page, perPage int) ([]models.PayrollItem, int, error) {
	// Count
	var total int
	err := database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM payroll_items
		WHERE payroll_period_id::text = $1
	`, periodID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT
			pi.id, pi.payroll_period_id, pi.employee_id,
			pi.base_salary, pi.daily_wage, pi.total_days_worked,
			pi.allowances, pi.overtime_pay, pi.overtime_hours,
			pi.thr_amount, pi.bonus_amount, pi.gross_salary,
			pi.deductions, pi.pph21_amount,
			pi.bpjs_kesehatan, pi.bpjs_jht, pi.bpjs_jp,
			pi.loan_deduction, pi.other_deductions, pi.total_deductions,
			pi.net_salary,
			pi.company_cost,
			pi.bpjs_kesehatan_company, pi.bpjs_jht_company, pi.bpjs_jp_company,
			pi.bpjs_jkk, pi.bpjs_jkm,
			e.full_name, e.employee_id,
			COALESCE(d.name, ''), COALESCE(pos.name, ''),
			COALESCE(e.employment_status::text, ''),
			pi.status::text, COALESCE(pi.notes, ''),
			pi.created_at, pi.updated_at
		FROM payroll_items pi
		JOIN employees e ON e.id = pi.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions pos ON pos.id = e.position_id
		WHERE pi.payroll_period_id::text = $1
		ORDER BY d.name, e.full_name
		LIMIT $2 OFFSET $3
	`

	rows, err := database.Pool.Query(ctx, query, periodID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.PayrollItem
	for rows.Next() {
		var item models.PayrollItem
		var allowancesBytes, deductionsBytes, companyCostBytes []byte
		var status string
		var notes string

		if err := rows.Scan(
			&item.ID, &item.PayrollPeriodID, &item.EmployeeID,
			&item.BaseSalary, &item.DailyWage, &item.TotalDaysWorked,
			&allowancesBytes, &item.OvertimePay, &item.OvertimeHours,
			&item.THRAmount, &item.BonusAmount, &item.GrossSalary,
			&deductionsBytes, &item.PPh21Amount,
			&item.BPJSKesehatan, &item.BPJSJHT, &item.BPJSJP,
			&item.LoanDeduction, &item.OtherDeductions, &item.TotalDeductions,
			&item.NetSalary,
			&companyCostBytes,
			&item.BPJSKesehatanCompany, &item.BPJSJHTCompany, &item.BPJSJPCompany,
			&item.BPJSJKK, &item.BPJSJKM,
			&item.EmployeeName, &item.EmployeeIDCode,
			&item.DepartmentName, &item.PositionName,
			&item.EmploymentStatus,
			&status, &notes,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		// Parse JSONB
		if len(allowancesBytes) > 0 {
			json.Unmarshal(allowancesBytes, &item.Allowances)
		}
		if len(deductionsBytes) > 0 {
			json.Unmarshal(deductionsBytes, &item.Deductions)
		}
		if len(companyCostBytes) > 0 {
			json.Unmarshal(companyCostBytes, &item.CompanyCost)
		}
		if item.Allowances == nil {
			item.Allowances = []models.AllowanceItem{}
		}
		if item.Deductions == nil {
			item.Deductions = []models.AllowanceItem{}
		}
		if item.CompanyCost == nil {
			item.CompanyCost = []models.AllowanceItem{}
		}

		item.Status = models.PayrollPeriodStatus(status)
		item.Notes = notes

		items = append(items, item)
	}

	return items, total, nil
}

// GetPayslip returns a single payroll item with period info
func GetPayslip(ctx context.Context, periodID, employeeID string) (*models.PayslipResponse, error) {
	query := `
		SELECT
			pi.id, pi.payroll_period_id, pi.employee_id,
			pi.base_salary, pi.daily_wage, pi.total_days_worked,
			pi.allowances, pi.overtime_pay, pi.overtime_hours,
			pi.thr_amount, pi.bonus_amount, pi.gross_salary,
			pi.deductions, pi.pph21_amount,
			pi.bpjs_kesehatan, pi.bpjs_jht, pi.bpjs_jp,
			pi.loan_deduction, pi.other_deductions, pi.total_deductions,
			pi.net_salary,
			pi.company_cost,
			pi.bpjs_kesehatan_company, pi.bpjs_jht_company, pi.bpjs_jp_company,
			pi.bpjs_jkk, pi.bpjs_jkm,
			e.full_name, e.employee_id,
			COALESCE(d.name, ''), COALESCE(pos.name, ''),
			COALESCE(e.employment_status::text, ''),
			pi.status::text, COALESCE(pi.notes, ''),
			pi.created_at, pi.updated_at,
			pp.period_name, pp.status::text
		FROM payroll_items pi
		JOIN employees e ON e.id = pi.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions pos ON pos.id = e.position_id
		JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
		WHERE pi.payroll_period_id::text = $1
			AND pi.employee_id::text = $2
			AND pp.deleted_at IS NULL
	`

	var item models.PayslipResponse
	var allowancesBytes, deductionsBytes, companyCostBytes []byte
	var periodStatus string
	var status string

	err := database.Pool.QueryRow(ctx, query, periodID, employeeID).Scan(
		&item.ID, &item.PayrollPeriodID, &item.EmployeeID,
		&item.BaseSalary, &item.DailyWage, &item.TotalDaysWorked,
		&allowancesBytes, &item.OvertimePay, &item.OvertimeHours,
		&item.THRAmount, &item.BonusAmount, &item.GrossSalary,
		&deductionsBytes, &item.PPh21Amount,
		&item.BPJSKesehatan, &item.BPJSJHT, &item.BPJSJP,
		&item.LoanDeduction, &item.OtherDeductions, &item.TotalDeductions,
		&item.NetSalary,
		&companyCostBytes,
		&item.BPJSKesehatanCompany, &item.BPJSJHTCompany, &item.BPJSJPCompany,
		&item.BPJSJKK, &item.BPJSJKM,
		&item.EmployeeName, &item.EmployeeIDCode,
		&item.DepartmentName, &item.PositionName,
		&item.EmploymentStatus,
		&status, &item.Notes,
		&item.CreatedAt, &item.UpdatedAt,
		&item.PeriodName, &periodStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Parse JSONB
	if len(allowancesBytes) > 0 {
		json.Unmarshal(allowancesBytes, &item.Allowances)
	}
	if len(deductionsBytes) > 0 {
		json.Unmarshal(deductionsBytes, &item.Deductions)
	}
	if len(companyCostBytes) > 0 {
		json.Unmarshal(companyCostBytes, &item.CompanyCost)
	}
	if item.Allowances == nil {
		item.Allowances = []models.AllowanceItem{}
	}
	if item.Deductions == nil {
		item.Deductions = []models.AllowanceItem{}
	}
	if item.CompanyCost == nil {
		item.CompanyCost = []models.AllowanceItem{}
	}

	item.Status = models.PayrollPeriodStatus(status)

	return &item, nil
}

// UpdatePayrollPeriodSummary refreshes the summary counts in payroll_periods
func UpdatePayrollPeriodSummary(ctx context.Context, periodID string) error {
	query := `
		UPDATE payroll_periods pp
		SET
			total_employee = sub.cnt,
			total_gross = sub.gross,
			total_deductions = sub.deductions,
			total_net = sub.net,
			total_company_cost = sub.company
		FROM (
			SELECT
				COUNT(*) AS cnt,
				COALESCE(SUM(gross_salary), 0) AS gross,
				COALESCE(SUM(total_deductions), 0) AS deductions,
				COALESCE(SUM(net_salary), 0) AS net,
				COALESCE(SUM(bpjs_kesehatan_company + bpjs_jht_company + bpjs_jp_company + bpjs_jkk + bpjs_jkm), 0) AS company
			FROM payroll_items
			WHERE payroll_period_id::text = $1 AND status = 'calculated'
		) sub
		WHERE pp.id::text = $1 AND pp.deleted_at IS NULL
	`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, periodID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// GetAllActiveEmployeeIDs returns all active employee IDs
func GetAllActiveEmployeeIDs(ctx context.Context) ([]string, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT id::text FROM employees WHERE is_active = TRUE AND deleted_at IS NULL ORDER BY full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ListMyPayslips returns all payslips for a given employee
func ListMyPayslips(ctx context.Context, employeeID string) ([]models.PayslipResponse, error) {
	query := `
		SELECT
			pi.id, pi.payroll_period_id, pi.employee_id,
			pi.base_salary, pi.daily_wage, pi.total_days_worked,
			pi.allowances, pi.overtime_pay, pi.overtime_hours,
			pi.thr_amount, pi.bonus_amount, pi.gross_salary,
			pi.deductions, pi.pph21_amount,
			pi.bpjs_kesehatan, pi.bpjs_jht, pi.bpjs_jp,
			pi.loan_deduction, pi.other_deductions, pi.total_deductions,
			pi.net_salary,
			pi.company_cost,
			pi.bpjs_kesehatan_company, pi.bpjs_jht_company, pi.bpjs_jp_company,
			pi.bpjs_jkk, pi.bpjs_jkm,
			e.full_name, e.employee_id,
			COALESCE(d.name, ''), COALESCE(pos.name, ''),
			COALESCE(e.employment_status::text, ''),
			pi.status::text, COALESCE(pi.notes, ''),
			pi.created_at, pi.updated_at,
			pp.period_name, pp.status::text
		FROM payroll_items pi
		JOIN employees e ON e.id = pi.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions pos ON pos.id = e.position_id
		JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
		WHERE pi.employee_id::text = $1
			AND pp.deleted_at IS NULL
		ORDER BY pp.year DESC, pp.month DESC
	`

	rows, err := database.Pool.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PayslipResponse
	for rows.Next() {
		var item models.PayslipResponse
		var allowancesBytes, deductionsBytes, companyCostBytes []byte
		var periodStatus string
		var status string

		err := rows.Scan(
			&item.ID, &item.PayrollPeriodID, &item.EmployeeID,
			&item.BaseSalary, &item.DailyWage, &item.TotalDaysWorked,
			&allowancesBytes, &item.OvertimePay, &item.OvertimeHours,
			&item.THRAmount, &item.BonusAmount, &item.GrossSalary,
			&deductionsBytes, &item.PPh21Amount,
			&item.BPJSKesehatan, &item.BPJSJHT, &item.BPJSJP,
			&item.LoanDeduction, &item.OtherDeductions, &item.TotalDeductions,
			&item.NetSalary,
			&companyCostBytes,
			&item.BPJSKesehatanCompany, &item.BPJSJHTCompany, &item.BPJSJPCompany,
			&item.BPJSJKK, &item.BPJSJKM,
			&item.EmployeeName, &item.EmployeeIDCode,
			&item.DepartmentName, &item.PositionName,
			&item.EmploymentStatus,
			&status, &item.Notes,
			&item.CreatedAt, &item.UpdatedAt,
			&item.PeriodName, &periodStatus,
		)
		if err != nil {
			return nil, err
		}

		if len(allowancesBytes) > 0 {
			json.Unmarshal(allowancesBytes, &item.Allowances)
		}
		if len(deductionsBytes) > 0 {
			json.Unmarshal(deductionsBytes, &item.Deductions)
		}
		if len(companyCostBytes) > 0 {
			json.Unmarshal(companyCostBytes, &item.CompanyCost)
		}
		if item.Allowances == nil {
			item.Allowances = []models.AllowanceItem{}
		}
		if item.Deductions == nil {
			item.Deductions = []models.AllowanceItem{}
		}
		if item.CompanyCost == nil {
			item.CompanyCost = []models.AllowanceItem{}
		}

		item.Status = models.PayrollPeriodStatus(status)

		items = append(items, item)
	}

	return items, nil
}
