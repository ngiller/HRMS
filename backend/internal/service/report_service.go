package service

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
)

type ReportService struct{}

func NewReportService() *ReportService {
	return &ReportService{}
}

type HeadcountReport struct {
	TotalEmployees   int            `json:"total_employees"`
	ActiveEmployees  int            `json:"active_employees"`
	ByDepartment     []DeptCount    `json:"by_department"`
	ByGender         map[string]int `json:"by_gender"`
	ByStatus         map[string]int `json:"by_status"`
	NewHiresThisYear int            `json:"new_hires_this_year"`
	ResignedThisYear int            `json:"resigned_this_year"`
}

type DeptCount struct {
	DepartmentID   string `json:"department_id"`
	DepartmentName string `json:"department_name"`
	Count          int    `json:"count"`
}

type PayrollSummaryReport struct {
	TotalPeriods      int     `json:"total_periods"`
	TotalGross        float64 `json:"total_gross"`
	TotalDeductions   float64 `json:"total_deductions"`
	TotalNetSalary    float64 `json:"total_net_salary"`
	TotalCompanyCost  float64 `json:"total_company_cost"`
	AverageNetSalary  float64 `json:"average_net_salary"`
	ByMonth           []MonthStat `json:"by_month"`
}

type MonthStat struct {
	Month     int     `json:"month"`
	Year      int     `json:"year"`
	Count     int     `json:"count"`
	TotalGross float64 `json:"total_gross"`
	TotalNet  float64 `json:"total_net"`
}

type AttendanceSummaryReport struct {
	TotalRecords    int            `json:"total_records"`
	OnTime          int            `json:"on_time"`
	Late            int            `json:"late"`
	LatePercentage  float64        `json:"late_percentage"`
	AverageWorkHours float64        `json:"average_work_hours"`
	ByDepartment    []DeptAttStat  `json:"by_department"`
}

type DeptAttStat struct {
	DepartmentID   string  `json:"department_id"`
	DepartmentName string  `json:"department_name"`
	Total          int     `json:"total"`
	OnTime         int     `json:"on_time"`
	Late           int     `json:"late"`
}

type LoanSummaryReport struct {
	TotalActive  int     `json:"total_active"`
	TotalOutstanding float64 `json:"total_outstanding"`
	TotalDisbursed float64 `json:"total_disbursed"`
	TotalApproved float64 `json:"total_approved"`
	PendingCount int     `json:"pending_count"`
	DefaultCount int     `json:"default_count"`
}

type LeaveSummaryReport struct {
	TotalRequests     int               `json:"total_requests"`
	Approved          int               `json:"approved"`
	Rejected          int               `json:"rejected"`
	Pending           int               `json:"pending"`
	TotalDaysApproved int               `json:"total_days_approved"`
	ByType            []LeaveTypeStat   `json:"by_type"`
}

type LeaveTypeStat struct {
	TypeName string `json:"type_name"`
	Count    int    `json:"count"`
	TotalDays int   `json:"total_days"`
}

type OvertimeSummaryReport struct {
	TotalRequests  int     `json:"total_requests"`
	Approved       int     `json:"approved"`
	TotalHours     float64 `json:"total_hours"`
	TotalCost      float64 `json:"total_cost"`
	PendingCount   int     `json:"pending_count"`
}

func (s *ReportService) Headcount(ctx context.Context, year int) (*HeadcountReport, error) {
	report := &HeadcountReport{}

	// Total employees
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL`).Scan(&report.TotalEmployees)

	// Active employees
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE status='active' AND deleted_at IS NULL`).Scan(&report.ActiveEmployees)

	// By department
	rows, err := database.Pool.Query(ctx, `
		SELECT COALESCE(d.id::text, ''), COALESCE(d.name, 'Tanpa Departemen'), COUNT(*)
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.deleted_at IS NULL
		GROUP BY d.id, d.name
		ORDER BY COUNT(*) DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d DeptCount
			if rows.Scan(&d.DepartmentID, &d.DepartmentName, &d.Count) == nil {
				report.ByDepartment = append(report.ByDepartment, d)
			}
		}
	}

	// By gender
	report.ByGender = map[string]int{}
	rows2, _ := database.Pool.Query(ctx, `SELECT COALESCE(gender, 'unknown'), COUNT(*) FROM employees WHERE deleted_at IS NULL GROUP BY gender`)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var g string
			var c int
			if rows2.Scan(&g, &c) == nil {
				report.ByGender[g] = c
			}
		}
	}

	// By employment status
	report.ByStatus = map[string]int{}
	rows3, _ := database.Pool.Query(ctx, `SELECT COALESCE(employment_status, 'unknown'), COUNT(*) FROM employees WHERE deleted_at IS NULL GROUP BY employment_status`)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var s string
			var c int
			if rows3.Scan(&s, &c) == nil {
				report.ByStatus[s] = c
			}
		}
	}

	// New hires this year
	if year == 0 {
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM NOW()) AND deleted_at IS NULL`).Scan(&report.NewHiresThisYear)
	} else {
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE EXTRACT(YEAR FROM created_at) = $1 AND deleted_at IS NULL`, year).Scan(&report.NewHiresThisYear)
	}

	// Resigned (soft-deleted) this year
	if year == 0 {
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NOT NULL AND EXTRACT(YEAR FROM deleted_at) = EXTRACT(YEAR FROM NOW())`).Scan(&report.ResignedThisYear)
	} else {
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NOT NULL AND EXTRACT(YEAR FROM deleted_at) = $1`, year).Scan(&report.ResignedThisYear)
	}

	return report, nil
}

func (s *ReportService) PayrollSummary(ctx context.Context, year int) (*PayrollSummaryReport, error) {
	report := &PayrollSummaryReport{}
	if year == 0 {
		year = 2026
	}

	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(gross_salary), 0), COALESCE(SUM(total_deductions), 0),
			COALESCE(SUM(net_salary), 0), COALESCE(SUM(company_cost), 0)
		FROM payroll_items pi
		JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
		WHERE pp.year = $1 AND pp.deleted_at IS NULL
	`, year).Scan(&report.TotalPeriods, &report.TotalGross, &report.TotalDeductions, &report.TotalNetSalary, &report.TotalCompanyCost)

	if report.TotalPeriods > 0 {
		report.AverageNetSalary = report.TotalNetSalary / float64(report.TotalPeriods)
	}

	rows, err := database.Pool.Query(ctx, `
		SELECT pp.month, pp.year, COUNT(pi.id), COALESCE(SUM(pi.gross_salary), 0), COALESCE(SUM(pi.net_salary), 0)
		FROM payroll_items pi
		JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
		WHERE pp.year = $1 AND pp.deleted_at IS NULL
		GROUP BY pp.month, pp.year ORDER BY pp.month
	`, year)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m MonthStat
			if rows.Scan(&m.Month, &m.Year, &m.Count, &m.TotalGross, &m.TotalNet) == nil {
				report.ByMonth = append(report.ByMonth, m)
			}
		}
	}

	return report, nil
}

func (s *ReportService) AttendanceSummary(ctx context.Context, year, month int, departmentID string) (*AttendanceSummaryReport, error) {
	report := &AttendanceSummaryReport{}

	where := "WHERE ar.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 0

	if year > 0 {
		argIdx++
		where += fmt.Sprintf(" AND EXTRACT(YEAR FROM ar.check_in) = $%d", argIdx)
		args = append(args, year)
	}
	if month > 0 {
		argIdx++
		where += fmt.Sprintf(" AND EXTRACT(MONTH FROM ar.check_in) = $%d", argIdx)
		args = append(args, month)
	}
	if departmentID != "" {
		argIdx++
		where += fmt.Sprintf(" AND e.department_id::text = $%d", argIdx)
		args = append(args, departmentID)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN ar.status = 'ontime' OR ar.status = 'present' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN ar.status = 'late' THEN 1 ELSE 0 END), 0)
		FROM attendance_records ar
		LEFT JOIN employees e ON e.id = ar.employee_id
		%s
	`, where)

	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&report.TotalRecords, &report.OnTime, &report.Late)
	if err == nil && report.TotalRecords > 0 {
		report.LatePercentage = float64(report.Late) / float64(report.TotalRecords) * 100
	}

	// Average work hours
	hourQuery := fmt.Sprintf(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (check_out - check_in))/3600), 0)
		FROM attendance_records ar
		LEFT JOIN employees e ON e.id = ar.employee_id
		%s AND ar.check_out IS NOT NULL
	`, where)
	database.Pool.QueryRow(ctx, hourQuery, args...).Scan(&report.AverageWorkHours)

	return report, nil
}

func (s *ReportService) LoanSummary(ctx context.Context) (*LoanSummaryReport, error) {
	report := &LoanSummaryReport{}

	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(remaining_balance), 0)
		FROM loans WHERE status IN ('active', 'approved') AND deleted_at IS NULL
	`).Scan(&report.TotalActive, &report.TotalOutstanding)

	database.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM loans WHERE status = 'disbursed' AND deleted_at IS NULL
	`).Scan(&report.TotalDisbursed)

	database.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM loans WHERE status = 'approved' AND deleted_at IS NULL
	`).Scan(&report.TotalApproved)

	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM loans WHERE status = 'pending' AND deleted_at IS NULL
	`).Scan(&report.PendingCount)

	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM loans WHERE status = 'defaulted' AND deleted_at IS NULL
	`).Scan(&report.DefaultCount)

	return report, nil
}

func (s *ReportService) LeaveSummary(ctx context.Context, year int, departmentID string) (*LeaveSummaryReport, error) {
	report := &LeaveSummaryReport{}

	where := "WHERE lr.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 0

	if year > 0 {
		argIdx++
		where += fmt.Sprintf(" AND EXTRACT(YEAR FROM lr.created_at) = $%d", argIdx)
		args = append(args, year)
	}
	if departmentID != "" {
		argIdx++
		where += fmt.Sprintf(" AND e.department_id::text = $%d", argIdx)
		args = append(args, departmentID)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN lr.status = 'approved' OR lr.status = 'paid' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lr.status = 'rejected' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lr.status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lr.status = 'approved' THEN lr.total_days ELSE 0 END), 0)
		FROM leave_requests lr
		LEFT JOIN employees e ON e.id = lr.employee_id
		%s
	`, where)
	database.Pool.QueryRow(ctx, countQuery, args...).Scan(&report.TotalRequests, &report.Approved, &report.Rejected, &report.Pending, &report.TotalDaysApproved)

	// By leave type
	typeQuery := fmt.Sprintf(`
		SELECT COALESCE(lt.name, ''), COUNT(*), COALESCE(SUM(lr.total_days), 0)
		FROM leave_requests lr
		LEFT JOIN employees e ON e.id = lr.employee_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		%s
		GROUP BY lt.name ORDER BY COUNT(*) DESC
	`, where)
	rows, err := database.Pool.Query(ctx, typeQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var l LeaveTypeStat
			if rows.Scan(&l.TypeName, &l.Count, &l.TotalDays) == nil {
				report.ByType = append(report.ByType, l)
			}
		}
	}

	return report, nil
}

func (s *ReportService) OvertimeSummary(ctx context.Context, year, month int) (*OvertimeSummaryReport, error) {
	report := &OvertimeSummaryReport{}

	where := "WHERE otr.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 0

	if year > 0 {
		argIdx++
		where += fmt.Sprintf(" AND EXTRACT(YEAR FROM otr.created_at) = $%d", argIdx)
		args = append(args, year)
	}
	if month > 0 {
		argIdx++
		where += fmt.Sprintf(" AND EXTRACT(MONTH FROM otr.created_at) = $%d", argIdx)
		args = append(args, month)
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN otr.status = 'approved' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN otr.status = 'approved' THEN otr.total_hours ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN otr.status = 'approved' THEN otr.total_pay ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN otr.status = 'pending' THEN 1 ELSE 0 END), 0)
		FROM overtime_requests otr
		%s
	`, where)
	database.Pool.QueryRow(ctx, query, args...).Scan(&report.TotalRequests, &report.Approved, &report.TotalHours, &report.TotalCost, &report.PendingCount)

	return report, nil
}
