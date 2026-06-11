package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

func ListEmployees(ctx context.Context, page, perPage int, search string) ([]models.EmployeeSummary, int, error) {
	// Count total
	countQuery := `
		SELECT COUNT(*) FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON p.department_id = d.id
		WHERE e.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND (LOWER(e.full_name) LIKE LOWER($%d) OR LOWER(e.email) LIKE LOWER($%d))", argIdx, argIdx)
		args = append(args, "%"+search+"%")
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch page
	offset := (page - 1) * perPage
	argIdx++
	query := fmt.Sprintf(`
		SELECT e.id, e.employee_id, e.full_name, e.email,
			e.gender, e.employment_status, e.is_active,
			COALESCE(r.name, '') as role_name,
			COALESCE(p.name, '') as position_name,
			COALESCE(d.name, '') as department_name,
			e.join_date, e.phone
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON p.department_id = d.id
		WHERE e.deleted_at IS NULL
	`)

	if search != "" {
		query += fmt.Sprintf(" AND (LOWER(e.full_name) LIKE LOWER($%d) OR LOWER(e.email) LIKE LOWER($%d))", argIdx-1, argIdx-1)
	}

	query += fmt.Sprintf(" ORDER BY e.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var employees []models.EmployeeSummary
	for rows.Next() {
		var emp models.EmployeeSummary
		err := rows.Scan(
			&emp.ID, &emp.EmployeeID, &emp.FullName, &emp.Email,
			&emp.Gender, &emp.EmploymentStatus, &emp.IsActive,
			&emp.RoleName, &emp.PositionName, &emp.DepartmentName,
			&emp.JoinDate, &emp.Phone,
		)
		if err != nil {
			return nil, 0, err
		}
		employees = append(employees, emp)
	}

	return employees, total, nil
}

func GetEmployeeByIDRepo(ctx context.Context, id string) (*models.Employee, error) {
	query := `
		SELECT e.id, e.employee_id, e.full_name, e.email, e.password_hash,
			e.gender::text, COALESCE(e.birth_place, ''), e.birth_date, COALESCE(e.religion::text, ''),
			COALESCE(e.marital_status::text, ''), e.join_date, e.employment_status::text, e.is_active,
			e.role_id, COALESCE(r.slug, ''), COALESCE(r.name, ''),
			e.position_id, COALESCE(p.name, ''),
			p.department_id, COALESCE(d.name, ''),
			COALESCE(e.phone, ''), COALESCE(e.address_domicile, ''), e.last_login_at, e.is_locked, e.locked_until,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON p.department_id = d.id
		WHERE (e.id::text = $1 OR e.employee_id = $1) AND e.deleted_at IS NULL
	`

	row := database.Pool.QueryRow(ctx, query, id)
	return scanEmployee(row)
}

func GetDashboardStats(ctx context.Context) (*models.DashboardResponse, error) {
	resp := &models.DashboardResponse{}

	// Total employees
	err := database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL AND is_active = TRUE`).Scan(&resp.TotalEmployees)
	if err != nil {
		return nil, err
	}

	// Present today (attendance records for today)
	err = database.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT employee_id) FROM attendance_records
		WHERE date = CURRENT_DATE AND status != 'tanpa_keterangan'
	`).Scan(&resp.PresentToday)
	if err != nil {
		resp.PresentToday = 0
	}

	// Pending approvals (approximate - count various pending requests)
	pendingCount := 0
	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT id FROM leave_requests WHERE status = 'pending' AND deleted_at IS NULL
			UNION ALL
			SELECT id FROM reimbursements WHERE status = 'pending' AND deleted_at IS NULL
			UNION ALL
			SELECT id FROM overtime_requests WHERE status = 'pending' AND deleted_at IS NULL
		) p
	`).Scan(&pendingCount)
	resp.PendingApprovals = pendingCount

	// Payroll this month placeholder
	resp.PayrollThisMonth = "Rp0"

	// Attendance by day (last 7 days)
	rows, err := database.Pool.Query(ctx, `
		SELECT to_char(date, 'Day') as day_name, COUNT(DISTINCT employee_id) as cnt
		FROM attendance_records
		WHERE date >= CURRENT_DATE - 6 AND date <= CURRENT_DATE
		GROUP BY date
		ORDER BY date
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day models.AttendanceDay
			if err := rows.Scan(&day.Day, &day.Count); err == nil {
				resp.AttendanceByDay = append(resp.AttendanceByDay, day)
			}
		}
	}

	// Employee composition by status
	compRows, err := database.Pool.Query(ctx, `
		SELECT employment_status::text, COUNT(*) as cnt
		FROM employees
		WHERE deleted_at IS NULL AND is_active = TRUE
		GROUP BY employment_status
		ORDER BY cnt DESC
	`)
	if err == nil {
		defer compRows.Close()
		for compRows.Next() {
			var comp models.EmployeeComposition
			if err := compRows.Scan(&comp.Status, &comp.Count); err == nil {
				resp.Composition = append(resp.Composition, comp)
			}
		}
	}

	// Recent employees
	recentQuery := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			e.gender, e.employment_status, e.is_active,
			COALESCE(r.name, '') as role_name,
			COALESCE(p.name, '') as position_name,
			COALESCE(d.name, '') as department_name,
			e.join_date, e.phone
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON p.department_id = d.id
		WHERE e.deleted_at IS NULL
		ORDER BY e.created_at DESC LIMIT 5
	`
	recentRows, err := database.Pool.Query(ctx, recentQuery)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var emp models.EmployeeSummary
			if err := recentRows.Scan(
				&emp.ID, &emp.EmployeeID, &emp.FullName, &emp.Email,
				&emp.Gender, &emp.EmploymentStatus, &emp.IsActive,
				&emp.RoleName, &emp.PositionName, &emp.DepartmentName,
				&emp.JoinDate, &emp.Phone,
			); err == nil {
				resp.RecentEmployees = append(resp.RecentEmployees, emp)
			}
		}
	}

	return resp, nil
}

func ListEmployeesForExport(ctx context.Context) ([]models.EmployeeSummary, error) {
	query := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			e.gender, e.employment_status, e.is_active,
			COALESCE(r.name, '') as role_name,
			COALESCE(p.name, '') as position_name,
			COALESCE(d.name, '') as department_name,
			e.join_date, e.phone
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON p.department_id = d.id
		WHERE e.deleted_at IS NULL
		ORDER BY e.full_name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.EmployeeSummary
	for rows.Next() {
		var emp models.EmployeeSummary
		err := rows.Scan(
			&emp.ID, &emp.EmployeeID, &emp.FullName, &emp.Email,
			&emp.Gender, &emp.EmploymentStatus, &emp.IsActive,
			&emp.RoleName, &emp.PositionName, &emp.DepartmentName,
			&emp.JoinDate, &emp.Phone,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}
