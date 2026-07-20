package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// GetEmployeeByName looks up an employee by full_name (case-insensitive) and returns their ID.
func GetEmployeeByName(ctx context.Context, name string) (*string, error) {
	var id string
	err := database.Pool.QueryRow(ctx, `SELECT id::text FROM employees WHERE LOWER(full_name) = LOWER($1) AND deleted_at IS NULL LIMIT 1`, name).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func ListEmployees(ctx context.Context, page, perPage int, search, departmentID, status string, includeDeleted bool) ([]models.EmployeeSummary, int, error) {
	// Count total
	deletedClause := "e.deleted_at IS NULL"
	if includeDeleted {
		deletedClause = "(e.deleted_at IS NULL OR e.deleted_at IS NOT NULL)"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE %s
	`, deletedClause)
	args := []interface{}{}
	argIdx := 0
	filters := []string{}

	if search != "" {
		argIdx++
		filters = append(filters, fmt.Sprintf("(LOWER(e.full_name) LIKE LOWER($%d) OR LOWER(e.email) LIKE LOWER($%d))", argIdx, argIdx))
		args = append(args, "%"+search+"%")
	}
	if departmentID != "" {
		argIdx++
		filters = append(filters, fmt.Sprintf("e.department_id::text = $%d", argIdx))
		args = append(args, departmentID)
	}
	if status != "" {
		if status == "active" {
			filters = append(filters, "e.is_active = true")
		} else if status == "inactive" {
			filters = append(filters, "e.is_active = false")
		} else {
			argIdx++
			filters = append(filters, fmt.Sprintf("e.employment_status::text = $%d", argIdx))
			args = append(args, status)
		}
	}

	for _, f := range filters {
		countQuery += " AND " + f
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
			COALESCE(e.gender::text, '') as gender,
			COALESCE(e.employment_status::text, '') as employment_status,
			e.is_active,
			COALESCE(r.name, '') as role_name,
			COALESCE(p.name, '') as position_name,
			COALESCE(d.name, '') as department_name,
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE %s
	`, deletedClause)

	for _, f := range filters {
		query += " AND " + f
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
			&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
			&emp.BaseSalary,
		)
		if err != nil {
			return nil, 0, err
		}
		employees = append(employees, emp)
	}

	return employees, total, nil
}

// UpdateEmployeeWorkSchedule updates only the work_schedule_id for an employee (individual override).
func UpdateEmployeeWorkSchedule(ctx context.Context, id, workScheduleID, userID string) (*models.Employee, error) {
	query := `
		UPDATE employees SET work_schedule_id = NULLIF(NULLIF($1, ''), 'null')::uuid, updated_at = NOW()
		WHERE id::text = $2 AND deleted_at IS NULL
		RETURNING id, employee_id, full_name, email, password_hash,
			gender::text, COALESCE(birth_place, ''), birth_date, COALESCE(religion::text, ''),
			COALESCE(marital_status::text, ''), join_date, employment_status::text, is_active,
			role_id, COALESCE((SELECT slug FROM roles WHERE id = role_id), ''), COALESCE((SELECT name FROM roles WHERE id = role_id), ''),
			position_id, COALESCE((SELECT name FROM positions WHERE id = position_id), ''),
			department_id,
			COALESCE((SELECT name FROM departments WHERE id = department_id), ''),
			work_schedule_id, COALESCE((SELECT name FROM work_schedules WHERE id = work_schedule_id), ''),
			approval_line_id,
			COALESCE((SELECT full_name FROM employees WHERE id = approval_line_id), '') as approval_line_name,
			COALESCE(phone, ''), COALESCE(address_domicile, ''), COALESCE(photo_url, ''),
			COALESCE(decrypt_sensitive(encrypted_nik), ''),
			COALESCE(decrypt_sensitive(encrypted_npwp), ''),
			COALESCE(decrypt_sensitive(encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(encrypted_bank_account), ''),
			COALESCE(decrypt_sensitive(encrypted_address_ktp), ''),
			COALESCE(ptkp_status::text, ''),
			is_pregnant,
			base_salary,
			daily_wage,
			last_login_at, is_locked, locked_until, created_at, updated_at
	`

	var employee *models.Employee
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, workScheduleID, id)
		var scanErr error
		employee, scanErr = scanEmployee(row)
		return scanErr
	})
	if err != nil {
		return nil, err
	}
	return employee, nil
}

func RestoreEmployee(ctx context.Context, id string, userID string) error {
	// Set deleted_at to NULL, updated_at to NOW, reactivate
	query := `UPDATE employees SET deleted_at = NULL, is_active = TRUE, updated_at = NOW() WHERE id::text = $1 AND deleted_at IS NOT NULL`

	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		tag, execErr := tx.Exec(ctx, query, id)
		if execErr != nil {
			return execErr
		}
		if tag.RowsAffected() == 0 {
			return errors.New("karyawan tidak ditemukan atau belum dinonaktifkan")
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func GetEmployeeByIDRepo(ctx context.Context, id string) (*models.Employee, error) {
	query := `
		SELECT e.id, e.employee_id, e.full_name, e.email, e.password_hash,
			e.gender::text, COALESCE(e.birth_place, ''), e.birth_date, COALESCE(e.religion::text, ''),
			COALESCE(e.marital_status::text, ''), e.join_date, e.employment_status::text, e.is_active,
			e.role_id, COALESCE(r.slug, ''), COALESCE(r.name, ''),
			e.position_id, COALESCE(p.name, ''),
			e.department_id, COALESCE(d.name, ''),
			e.work_schedule_id, COALESCE(ws.name, ''),
			e.approval_line_id,
			COALESCE((SELECT full_name FROM employees WHERE id = e.approval_line_id), '') as approval_line_name,
			COALESCE(e.phone, ''), COALESCE(e.address_domicile, ''), COALESCE(e.photo_url, ''),
			COALESCE(decrypt_sensitive(e.encrypted_nik), ''),
			COALESCE(decrypt_sensitive(e.encrypted_npwp), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_account), ''),
			COALESCE(decrypt_sensitive(e.encrypted_address_ktp), ''),
			COALESCE(e.ptkp_status::text, ''),
			e.is_pregnant,
			e.base_salary,
			e.daily_wage,
			e.last_login_at, e.is_locked, e.locked_until,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		LEFT JOIN work_schedules ws ON e.work_schedule_id = ws.id
		WHERE (e.id::text = $1 OR e.employee_id = $1) AND e.deleted_at IS NULL
	`

	row := database.Pool.QueryRow(ctx, query, id)
	return scanEmployee(row)
}

func GetDashboardStats(ctx context.Context) (*models.DashboardResponse, error) {
	resp := &models.DashboardResponse{}

	// Total employees (including non-active)
	err := database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL`).Scan(&resp.TotalEmployees)
	if err != nil {
		return nil, err
	}

	// Active employees
	err = database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL AND is_active = TRUE`).Scan(&resp.ActiveEmployees)
	if err != nil {
		resp.ActiveEmployees = 0
	}

	// Present today (attendance records for today)
	err = database.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT employee_id) FROM attendance_records
		WHERE date = CURRENT_DATE AND status != 'tanpa_keterangan'
	`).Scan(&resp.PresentToday)
	if err != nil {
		resp.PresentToday = 0
	}

	// Attendance rate (present / active employees)
	if resp.ActiveEmployees > 0 {
		resp.AttendanceRate = float64(resp.PresentToday) / float64(resp.ActiveEmployees) * 100
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

	// Monthly hiring trend (last 6 months)
	trendRows, err := database.Pool.Query(ctx, `
		SELECT to_char(join_date, 'Mon') as month_name, COUNT(*) as cnt
		FROM employees
		WHERE deleted_at IS NULL AND join_date >= CURRENT_DATE - INTERVAL '6 months'
		GROUP BY to_char(join_date, 'Mon'), date_trunc('month', join_date)
		ORDER BY date_trunc('month', join_date)
	`)
	if err == nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var trend models.MonthlyTrend
			if err := trendRows.Scan(&trend.Month, &trend.Count); err == nil {
				resp.MonthlyTrend = append(resp.MonthlyTrend, trend)
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

	// Gender breakdown
	genderRows, err := database.Pool.Query(ctx, `
		SELECT gender::text, COUNT(*) as cnt
		FROM employees
		WHERE deleted_at IS NULL AND is_active = TRUE
		GROUP BY gender
		ORDER BY cnt DESC
	`)
	if err == nil {
		defer genderRows.Close()
		for genderRows.Next() {
			var g models.GenderBreakdown
			if err := genderRows.Scan(&g.Gender, &g.Count); err == nil {
				resp.GenderBreakdown = append(resp.GenderBreakdown, g)
			}
		}
	}

	// Department stats
	deptRows, err := database.Pool.Query(ctx, `
		SELECT COALESCE(d.name, 'Tanpa Departemen') as dept_name, COUNT(*) as cnt
		FROM employees e
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		GROUP BY d.name
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err == nil {
		defer deptRows.Close()
		for deptRows.Next() {
			var ds models.DepartmentStat
			if err := deptRows.Scan(&ds.DepartmentName, &ds.EmployeeCount); err == nil {
				resp.DepartmentStats = append(resp.DepartmentStats, ds)
			}
		}
	}

	// Recent employees
	recentQuery := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			COALESCE(e.gender::text, '') as gender,
			COALESCE(e.employment_status::text, '') as employment_status,
			e.is_active,
			COALESCE(r.name, '') as role_name,
			COALESCE(p.name, '') as position_name,
			COALESCE(d.name, '') as department_name,
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
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
				&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
				&emp.BaseSalary,
			); err == nil {
				resp.RecentEmployees = append(resp.RecentEmployees, emp)
			}
		}
	}

	// Absent today: active employees with no check-in today + leave info
	absentRows, err := database.Pool.Query(ctx, `
		SELECT e.id::text, e.full_name, COALESCE(d.name, '') as dept_name,
			CASE
				WHEN lt.name IS NOT NULL THEN LOWER(lt.name)
				ELSE 'tanpa_keterangan'
			END as absence_reason,
			COALESCE(lr.reason, '')
		FROM employees e
		LEFT JOIN departments d ON e.department_id = d.id
		LEFT JOIN leave_requests lr ON lr.employee_id = e.id
			AND CURRENT_DATE BETWEEN lr.start_date AND lr.end_date
			AND lr.status = 'approved'
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		AND NOT EXISTS (
			SELECT 1 FROM attendance_records ar
			WHERE ar.employee_id = e.id AND ar.date = CURRENT_DATE
		)
		ORDER BY e.full_name ASC
	`)
	if err == nil {
		defer absentRows.Close()
		for absentRows.Next() {
			var ae models.AbsentEmployee
			if err := absentRows.Scan(&ae.EmployeeID, &ae.FullName, &ae.DepartmentName, &ae.AbsenceReason, &ae.LeaveReason); err == nil {
				resp.AbsentToday = append(resp.AbsentToday, ae)
			}
		}
	}
	if resp.AbsentToday == nil {
		resp.AbsentToday = []models.AbsentEmployee{}
	}

	return resp, nil
}

func CreateEmployee(ctx context.Context, req *models.CreateEmployeeRequest, userID string) (*models.Employee, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi password: %w", err)
	}

	query := `
		INSERT INTO employees (employee_id, full_name, email, password_hash, gender,
			birth_place, birth_date, religion, marital_status, join_date,
			employment_status, base_salary, daily_wage, phone, address_domicile,
			role_id, position_id, department_id,
			encrypted_nik, encrypted_npwp, encrypted_bank_name, encrypted_bank_account, encrypted_address_ktp,
			blood_type, ptkp_status, end_date, position_grade_id, approval_line_id)
		VALUES ($1, $2, $3, $4, $5::gender_type,
			NULLIF($6, ''), NULLIF(NULLIF($7, ''), 'null')::date, NULLIF($8, '')::religion_type,
			NULLIF($9, '')::marital_status, NULLIF($10, '')::date,
			$11::employment_status, $12, $13, NULLIF($14, ''), NULLIF($15, ''),
			NULLIF(NULLIF($16, ''), 'null')::uuid, NULLIF(NULLIF($17, ''), 'null')::uuid,
			NULLIF(NULLIF($18, ''), 'null')::uuid,
			encrypt_sensitive(NULLIF($19, '')), encrypt_sensitive(NULLIF($20, '')),
			encrypt_sensitive(NULLIF($21, '')), encrypt_sensitive(NULLIF($22, '')),
			encrypt_sensitive(NULLIF($23, '')),
			NULLIF($24, ''), NULLIF($25, '')::ptkp_status, NULLIF(NULLIF($26, ''), 'null')::date,
			NULLIF(NULLIF($27, ''), 'null')::uuid, NULLIF(NULLIF($28, ''), 'null')::uuid)
		RETURNING id, employee_id, full_name, email, password_hash,
			gender::text, COALESCE(birth_place, ''), birth_date, COALESCE(religion::text, ''),
			COALESCE(marital_status::text, ''), join_date, employment_status::text, is_active,
			role_id, COALESCE((SELECT slug FROM roles WHERE id = role_id), ''), COALESCE((SELECT name FROM roles WHERE id = role_id), ''),
			position_id, COALESCE((SELECT name FROM positions WHERE id = position_id), ''),
			department_id,
			COALESCE((SELECT name FROM departments WHERE id = department_id), ''),
			work_schedule_id, COALESCE((SELECT name FROM work_schedules WHERE id = work_schedule_id), ''),
			approval_line_id,
			COALESCE((SELECT full_name FROM employees WHERE id = approval_line_id), '') as approval_line_name,
			COALESCE(phone, ''), COALESCE(address_domicile, ''), COALESCE(photo_url, ''),
			COALESCE(decrypt_sensitive(encrypted_nik), ''),
			COALESCE(decrypt_sensitive(encrypted_npwp), ''),
			COALESCE(decrypt_sensitive(encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(encrypted_bank_account), ''),
			COALESCE(decrypt_sensitive(encrypted_address_ktp), ''),
			COALESCE(ptkp_status::text, ''),
			is_pregnant,
			base_salary,
			daily_wage,
			last_login_at, is_locked, locked_until, created_at, updated_at
	`

	var employee *models.Employee
	err = database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		var baseSalaryArg interface{} = nil
		if req.BaseSalary != nil {
			baseSalaryArg = fmt.Sprintf("%.2f", *req.BaseSalary)
		}
		var dailyWageArg interface{} = nil
		if req.DailyWage != nil {
			dailyWageArg = fmt.Sprintf("%.2f", *req.DailyWage)
		}
		row := tx.QueryRow(ctx, query,
			req.EmployeeID, req.FullName, req.Email, string(passwordHash),
			req.Gender, req.PlaceOfBirth, req.DateOfBirth, req.Religion,
			req.MaritalStatus, req.JoinDate, req.EmploymentStatus,
			baseSalaryArg, dailyWageArg,
			req.Phone, req.Address, req.RoleID, req.PositionID, req.DepartmentID,
			req.NIK, req.NPWP, req.BankName, req.BankAccount, req.AddressKTP,
			req.BloodType, req.PTKPStatus, req.EndDate, req.PositionGradeID, req.ApprovalLineID,
		)
		var scanErr error
		employee, scanErr = scanEmployee(row)
		return scanErr
	})
	if err != nil {
		return nil, err
	}

	// If base_salary provided, also insert into employee_salary_histories
	if req.BaseSalary != nil && *req.BaseSalary > 0 {
		if err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
			dailyWageVal := float64(0)
			if req.DailyWage != nil {
				dailyWageVal = *req.DailyWage
			}
			_, histErr := tx.Exec(ctx, `
				INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason, changed_by)
				VALUES ($1::uuid, $2, $3, CURRENT_DATE, 'Initial salary from employee creation', $4::uuid)
			`, employee.ID, *req.BaseSalary, dailyWageVal, userID)
			return histErr
		}); err != nil {
			log.Printf("WARNING: Gagal insert salary history: %v", err)
		}
	}

	return employee, nil
}

func UpdateEmployee(ctx context.Context, id string, req *models.UpdateEmployeeRequest, userID string) (*models.Employee, error) {
	sets := []string{}
	args := []interface{}{}
	argIdx := 0

	addSet := func(col string, val interface{}) {
		argIdx++
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
	}

	if req.FullName != "" {
		addSet("full_name", req.FullName)
	}
	if req.Email != "" {
		addSet("email", req.Email)
	}
	if req.Gender != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("gender = $%d::gender_type", argIdx))
		args = append(args, req.Gender)
	}
	if req.PlaceOfBirth != "" {
		addSet("birth_place", req.PlaceOfBirth)
	}
	if req.DateOfBirth != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("birth_date = NULLIF($%d::text, '')::date", argIdx))
		args = append(args, req.DateOfBirth)
	}
	if req.Religion != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("religion = $%d::religion_type", argIdx))
		args = append(args, req.Religion)
	}
	if req.MaritalStatus != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("marital_status = $%d::marital_status", argIdx))
		args = append(args, req.MaritalStatus)
	}
	if req.JoinDate != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("join_date = NULLIF($%d::text, '')::date", argIdx))
		args = append(args, req.JoinDate)
	}
	if req.EmploymentStatus != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("employment_status = $%d::employment_status", argIdx))
		args = append(args, req.EmploymentStatus)
	}
	if req.IsActive != nil {
		addSet("is_active", *req.IsActive)
	}
	if req.BaseSalary != nil {
		argIdx++
		sets = append(sets, fmt.Sprintf("base_salary = $%d", argIdx))
		args = append(args, *req.BaseSalary)
	}
	if req.DailyWage != nil {
		argIdx++
		sets = append(sets, fmt.Sprintf("daily_wage = $%d", argIdx))
		args = append(args, *req.DailyWage)
	}
	if req.Phone != "" {
		addSet("phone", req.Phone)
	}
	if req.Address != "" {
		addSet("address_domicile", req.Address)
	}
	if req.NIK != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("encrypted_nik = encrypt_sensitive(NULLIF($%d, ''))", argIdx))
		args = append(args, req.NIK)
	}
	if req.NPWP != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("encrypted_npwp = encrypt_sensitive(NULLIF($%d, ''))", argIdx))
		args = append(args, req.NPWP)
	}
	if req.BankName != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("encrypted_bank_name = encrypt_sensitive(NULLIF($%d, ''))", argIdx))
		args = append(args, req.BankName)
	}
	if req.BankAccount != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("encrypted_bank_account = encrypt_sensitive(NULLIF($%d, ''))", argIdx))
		args = append(args, req.BankAccount)
	}
	if req.AddressKTP != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("encrypted_address_ktp = encrypt_sensitive(NULLIF($%d, ''))", argIdx))
		args = append(args, req.AddressKTP)
	}
	if req.RoleID != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("role_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, req.RoleID)
	}
	if req.PositionID != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("position_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, req.PositionID)
	}
	if req.DepartmentID != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("department_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, req.DepartmentID)
	}
	if req.WorkScheduleID != "" {
		argIdx++
		sets = append(sets, fmt.Sprintf("work_schedule_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, req.WorkScheduleID)
	}
	if req.ApprovalLineID != nil {
		argIdx++
		sets = append(sets, fmt.Sprintf("approval_line_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, *req.ApprovalLineID)
	}
	if req.IsPregnant != nil {
		addSet("is_pregnant", *req.IsPregnant)
	}

	if len(sets) == 0 {
		return nil, errors.New("tidak ada data yang diubah")
	}

	argIdx++
	query := fmt.Sprintf(`
		UPDATE employees SET %s, updated_at = NOW()
		WHERE id::text = $%d AND deleted_at IS NULL
		RETURNING id, employee_id, full_name, email, password_hash,
			gender::text, COALESCE(birth_place, ''), birth_date, COALESCE(religion::text, ''),
			COALESCE(marital_status::text, ''), join_date, employment_status::text, is_active,
			role_id, COALESCE((SELECT slug FROM roles WHERE id = role_id), ''), COALESCE((SELECT name FROM roles WHERE id = role_id), ''),
			position_id, COALESCE((SELECT name FROM positions WHERE id = position_id), ''),
			department_id,
			COALESCE((SELECT name FROM departments WHERE id = department_id), ''),
			work_schedule_id, COALESCE((SELECT name FROM work_schedules WHERE id = work_schedule_id), ''),
			approval_line_id,
			COALESCE((SELECT full_name FROM employees WHERE id = approval_line_id), '') as approval_line_name,
			COALESCE(phone, ''), COALESCE(address_domicile, ''), COALESCE(photo_url, ''),
			COALESCE(decrypt_sensitive(encrypted_nik), ''),
			COALESCE(decrypt_sensitive(encrypted_npwp), ''),
			COALESCE(decrypt_sensitive(encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(encrypted_bank_account), ''),
			COALESCE(decrypt_sensitive(encrypted_address_ktp), ''),
			COALESCE(ptkp_status::text, ''),
			is_pregnant,
			base_salary,
			daily_wage,
			last_login_at, is_locked, locked_until, created_at, updated_at
	`, strings.Join(sets, ", "), argIdx)
	args = append(args, id)

	var employee *models.Employee
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, args...)
		var scanErr error
		employee, scanErr = scanEmployee(row)
		return scanErr
	})
	if err != nil {
		return nil, err
	}

	// If base_salary changed, insert into employee_salary_histories
	if req.BaseSalary != nil && *req.BaseSalary > 0 {
		if err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
			_, histErr := tx.Exec(ctx, `
				INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason, changed_by)
				VALUES ($1::uuid, $2, COALESCE((SELECT daily_wage FROM employees WHERE id = $1::uuid), 0), CURRENT_DATE, 'Salary update from employee edit', $3::uuid)
			`, employee.ID, *req.BaseSalary, userID)
			return histErr
		}); err != nil {
			log.Printf("WARNING: Gagal insert salary history: %v", err)
		}
	}

	return employee, nil
}

func DeleteEmployee(ctx context.Context, id string, userID string) error {
	query := `UPDATE employees SET deleted_at = NOW(), is_active = FALSE, updated_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`

	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		tag, execErr := tx.Exec(ctx, query, id)
		if execErr != nil {
			return execErr
		}
		if tag.RowsAffected() == 0 {
			return errors.New("karyawan tidak ditemukan")
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func CheckEmployeeExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := database.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE id::text = $1 AND deleted_at IS NULL)`, id).Scan(&exists)
	return exists, err
}

func CheckEmailExists(ctx context.Context, email, excludeID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM employees WHERE email = $1 AND deleted_at IS NULL`
	args := []interface{}{email}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	query += `)`
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

func CheckEmployeeIDExists(ctx context.Context, empID, excludeID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM employees WHERE employee_id = $1 AND deleted_at IS NULL`
	args := []interface{}{empID}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	query += `)`
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

func GetEmployeeHistory(ctx context.Context, employeeID string, page, perPage int) ([]models.EmployeeHistory, int, error) {
	countQuery := `SELECT COUNT(*) FROM employee_histories WHERE employee_id::text = $1`
	var total int
	err := database.Pool.QueryRow(ctx, countQuery, employeeID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT eh.id::text, eh.employee_id::text, COALESCE(e.full_name, '') as employee_name,
			eh.change_type, COALESCE(eh.old_value::text, ''), COALESCE(eh.new_value::text, ''),
			COALESCE(eh.reason, ''),
			COALESCE(eh.changed_by::text, ''), COALESCE(c.full_name, '') as changed_by_name,
			eh.changed_at
		FROM employee_histories eh
		LEFT JOIN employees e ON eh.employee_id = e.id
		LEFT JOIN employees c ON eh.changed_by = c.id
		WHERE eh.employee_id::text = $1
		ORDER BY eh.changed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := database.Pool.Query(ctx, query, employeeID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []models.EmployeeHistory
	for rows.Next() {
		var h models.EmployeeHistory
		var oldValStr, newValStr string
		if err := rows.Scan(&h.ID, &h.EmployeeID, &h.EmployeeName,
			&h.ChangeType, &oldValStr, &newValStr,
			&h.Reason, &h.ChangedBy, &h.ChangedByName, &h.ChangedAt); err != nil {
			return nil, 0, err
		}
		if oldValStr != "" && oldValStr != "null" {
			json.Unmarshal([]byte(oldValStr), &h.OldValue)
		}
		if newValStr != "" && newValStr != "null" {
			json.Unmarshal([]byte(newValStr), &h.NewValue)
		}
		histories = append(histories, h)
	}

	return histories, total, nil
}

func CreateEmployeeHistory(ctx context.Context, employeeID, changeType string, oldValue, newValue map[string]any, reason, changedBy string) error {
	query := `
		INSERT INTO employee_histories (employee_id, change_type, old_value, new_value, reason, changed_by)
		VALUES ($1::uuid, $2,
			CASE WHEN $3::text IS NOT NULL AND $3::text != '' THEN $3::jsonb ELSE NULL END,
			CASE WHEN $4::text IS NOT NULL AND $4::text != '' THEN $4::jsonb ELSE NULL END,
			NULLIF($5, ''), NULLIF($6, '')::uuid)
	`

	var oldJSON, newJSON *string
	if oldValue != nil {
		b, err := json.Marshal(oldValue)
		if err == nil {
			s := string(b)
			oldJSON = &s
		}
	}
	if newValue != nil {
		b, err := json.Marshal(newValue)
		if err == nil {
			s := string(b)
			newJSON = &s
		}
	}

	return database.WithUserContext(ctx, changedBy, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, employeeID, changeType, oldJSON, newJSON, reason, changedBy)
		return err
	})

}

func UpdateEmployeePhoto(ctx context.Context, id, photoURL, userID string) error {
	query := `UPDATE employees SET photo_url = $1, updated_at = NOW() WHERE id::text = $2 AND deleted_at IS NULL`
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, photoURL, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("karyawan tidak ditemukan")
		}
		return nil
	})
}

func GetManagerDashboardStats(ctx context.Context, userID uuid.UUID) (*models.ManagerDashboardResponse, error) {
	resp := &models.ManagerDashboardResponse{}

	// Find the department where this user is the head
	var deptID uuid.UUID
	err := database.Pool.QueryRow(ctx, `SELECT id FROM departments WHERE head_id = $1 AND deleted_at IS NULL LIMIT 1`, userID).Scan(&deptID)
	if err != nil {
		// User is not a department head, return empty stats
		resp.Composition = []models.EmployeeComposition{}
		resp.RecentMembers = []models.EmployeeSummary{}
		return resp, nil
	}

	// Team size (all employees in this department)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE department_id = $1 AND deleted_at IS NULL`, deptID).Scan(&resp.TeamSize)

	// Active team
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE department_id = $1 AND deleted_at IS NULL AND is_active = TRUE`, deptID).Scan(&resp.ActiveTeam)

	// Attendance today for team
	database.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT ar.employee_id) FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		WHERE e.department_id = $1 AND ar.date = CURRENT_DATE AND ar.status != 'tanpa_keterangan'
	`, deptID).Scan(&resp.AttendanceToday)

	// Pending approvals (leave, reimbursement, overtime from this dept)
	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT id FROM leave_requests lr JOIN employees e ON lr.employee_id = e.id WHERE e.department_id = $1 AND lr.status = 'pending' AND lr.deleted_at IS NULL
			UNION ALL
			SELECT id FROM reimbursements r JOIN employees e ON r.employee_id = e.id WHERE e.department_id = $1 AND r.status = 'pending' AND r.deleted_at IS NULL
			UNION ALL
			SELECT id FROM overtime_requests otr JOIN employees e ON otr.employee_id = e.id WHERE e.department_id = $1 AND otr.status = 'pending' AND otr.deleted_at IS NULL
		) p
	`, deptID).Scan(&resp.PendingApprovals)

	// Composition by employment status
	compRows, err := database.Pool.Query(ctx, `
		SELECT employment_status::text, COUNT(*) as cnt FROM employees
		WHERE department_id = $1 AND deleted_at IS NULL AND is_active = TRUE
		GROUP BY employment_status ORDER BY cnt DESC
	`, deptID)
	if err == nil {
		defer compRows.Close()
		for compRows.Next() {
			var comp models.EmployeeComposition
			if err := compRows.Scan(&comp.Status, &comp.Count); err == nil {
				resp.Composition = append(resp.Composition, comp)
			}
		}
	}
	if resp.Composition == nil {
		resp.Composition = []models.EmployeeComposition{}
	}

	// Recent team members
	recentQuery := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			COALESCE(e.gender::text, ''), COALESCE(e.employment_status::text, ''),
			e.is_active, COALESCE(r.name, ''), COALESCE(p.name, ''), COALESCE(d.name, ''),
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.department_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.created_at DESC LIMIT 5
	`
	recentRows, err := database.Pool.Query(ctx, recentQuery, deptID)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var emp models.EmployeeSummary
			if err := recentRows.Scan(
				&emp.ID, &emp.EmployeeID, &emp.FullName, &emp.Email,
				&emp.Gender, &emp.EmploymentStatus, &emp.IsActive,
				&emp.RoleName, &emp.PositionName, &emp.DepartmentName,
				&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
				&emp.BaseSalary,
			); err == nil {
				resp.RecentMembers = append(resp.RecentMembers, emp)
			}
		}
	}
	if resp.RecentMembers == nil {
		resp.RecentMembers = []models.EmployeeSummary{}
	}

	return resp, nil
}

func GetHRDashboardStats(ctx context.Context) (*models.HRDashboardResponse, error) {
	resp := &models.HRDashboardResponse{}

	// Total employees
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL`).Scan(&resp.TotalEmployees)

	// Active employees
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL AND is_active = TRUE`).Scan(&resp.ActiveEmployees)

	// Department count
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM departments WHERE deleted_at IS NULL AND is_active = TRUE`).Scan(&resp.DepartmentCount)

	// Hiring this month
	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM employees
		WHERE deleted_at IS NULL AND join_date >= date_trunc('month', CURRENT_DATE)
	`).Scan(&resp.HiringThisMonth)

	// Composition
	compRows, err := database.Pool.Query(ctx, `
		SELECT employment_status::text, COUNT(*) as cnt FROM employees
		WHERE deleted_at IS NULL AND is_active = TRUE GROUP BY employment_status ORDER BY cnt DESC
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
	if resp.Composition == nil {
		resp.Composition = []models.EmployeeComposition{}
	}

	// Gender breakdown
	genderRows, err := database.Pool.Query(ctx, `
		SELECT gender::text, COUNT(*) as cnt FROM employees
		WHERE deleted_at IS NULL AND is_active = TRUE GROUP BY gender ORDER BY cnt DESC
	`)
	if err == nil {
		defer genderRows.Close()
		for genderRows.Next() {
			var g models.GenderBreakdown
			if err := genderRows.Scan(&g.Gender, &g.Count); err == nil {
				resp.GenderBreakdown = append(resp.GenderBreakdown, g)
			}
		}
	}
	if resp.GenderBreakdown == nil {
		resp.GenderBreakdown = []models.GenderBreakdown{}
	}

	// Department stats
	deptRows, err := database.Pool.Query(ctx, `
		SELECT COALESCE(d.name, 'Tanpa Departemen'), COUNT(*) FROM employees e
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		GROUP BY d.name ORDER BY COUNT(*) DESC LIMIT 10
	`)
	if err == nil {
		defer deptRows.Close()
		for deptRows.Next() {
			var ds models.DepartmentStat
			if err := deptRows.Scan(&ds.DepartmentName, &ds.EmployeeCount); err == nil {
				resp.DepartmentStats = append(resp.DepartmentStats, ds)
			}
		}
	}
	if resp.DepartmentStats == nil {
		resp.DepartmentStats = []models.DepartmentStat{}
	}

	// Birthdays this month (active employees)
	bdayQuery := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			COALESCE(e.gender::text, ''), COALESCE(e.employment_status::text, ''),
			e.is_active, COALESCE(r.name, ''), COALESCE(p.name, ''), COALESCE(d.name, ''),
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		AND EXTRACT(MONTH FROM e.birth_date) = EXTRACT(MONTH FROM CURRENT_DATE)
		ORDER BY EXTRACT(DAY FROM e.birth_date) ASC
	`
	bdayRows, err := database.Pool.Query(ctx, bdayQuery)
	if err == nil {
		defer bdayRows.Close()
		for bdayRows.Next() {
			var emp models.EmployeeSummary
			if err := bdayRows.Scan(
				&emp.ID, &emp.EmployeeID, &emp.FullName, &emp.Email,
				&emp.Gender, &emp.EmploymentStatus, &emp.IsActive,
				&emp.RoleName, &emp.PositionName, &emp.DepartmentName,
				&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
				&emp.BaseSalary,
			); err == nil {
				resp.BirthdaysThisMonth = append(resp.BirthdaysThisMonth, emp)
			}
		}
	}
	if resp.BirthdaysThisMonth == nil {
		resp.BirthdaysThisMonth = []models.EmployeeSummary{}
	}

	// Contracts expiring within 30 days
	expiringQuery := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			COALESCE(e.gender::text, ''), COALESCE(e.employment_status::text, ''),
			e.is_active, COALESCE(r.name, ''), COALESCE(p.name, ''), COALESCE(d.name, ''),
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		AND e.employment_status = 'kontrak'
		AND e.end_date IS NOT NULL
		AND e.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
		ORDER BY e.end_date ASC
	`
	expRows, err := database.Pool.Query(ctx, expiringQuery)
	if err == nil {
		defer expRows.Close()
		for expRows.Next() {
			var emp models.EmployeeSummary
			if err := expRows.Scan(
				&emp.ID, &emp.EmployeeID, &emp.FullName, &emp.Email,
				&emp.Gender, &emp.EmploymentStatus, &emp.IsActive,
				&emp.RoleName, &emp.PositionName, &emp.DepartmentName,
				&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
				&emp.BaseSalary,
			); err == nil {
				resp.ContractExpiring = append(resp.ContractExpiring, emp)
			}
		}
	}
	if resp.ContractExpiring == nil {
		resp.ContractExpiring = []models.EmployeeSummary{}
	}

	// Recent employees
	recentQuery := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			COALESCE(e.gender::text, ''), COALESCE(e.employment_status::text, ''),
			e.is_active, COALESCE(r.name, ''), COALESCE(p.name, ''), COALESCE(d.name, ''),
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.deleted_at IS NULL
		ORDER BY e.created_at DESC LIMIT 10
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
				&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
				&emp.BaseSalary,
			); err == nil {
				resp.RecentEmployees = append(resp.RecentEmployees, emp)
			}
		}
	}
	if resp.RecentEmployees == nil {
		resp.RecentEmployees = []models.EmployeeSummary{}
	}

	return resp, nil
}

// UpdateFaceDescriptor stores a face descriptor (JSON array of 128 floats) for an employee
func UpdateFaceDescriptor(ctx context.Context, employeeID, descriptorJSON, userID string) error {
	query := `UPDATE employees SET face_descriptor = NULLIF($1, '')::jsonb, face_descriptor_updated_at = NOW(), updated_at = NOW() WHERE id::text = $2 AND deleted_at IS NULL`
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, descriptorJSON, employeeID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errors.New("karyawan tidak ditemukan")
		}
		return nil
	})
}

// GetFaceDescriptor retrieves the face descriptor for an employee
func GetFaceDescriptor(ctx context.Context, employeeID string) (*string, error) {
	var descriptor *string
	err := database.Pool.QueryRow(ctx, `SELECT face_descriptor::text FROM employees WHERE id::text = $1 AND deleted_at IS NULL`, employeeID).Scan(&descriptor)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return descriptor, nil
}

func ListEmployeesForExport(ctx context.Context) ([]models.EmployeeSummary, error) {
	query := `
		SELECT e.id, e.employee_id, e.full_name, e.email,
			COALESCE(e.gender::text, '') as gender,
			COALESCE(e.employment_status::text, '') as employment_status,
			e.is_active,
			COALESCE(r.name, '') as role_name,
			COALESCE(p.name, '') as position_name,
			COALESCE(d.name, '') as department_name,
			e.join_date, COALESCE(e.phone, ''), e.deleted_at,
			COALESCE(e.base_salary, 0) as base_salary
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
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
			&emp.JoinDate, &emp.Phone, &emp.DeletedAt,
			&emp.BaseSalary,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}

func ListEmployeesForExportFull(ctx context.Context) ([]models.EmployeeExportFull, error) {
	query := `
		SELECT
			COALESCE(e.employee_id, ''),
			COALESCE(e.full_name, ''),
			'', -- barcode
			COALESCE(d.name, ''), -- organization
			COALESCE(p.name, ''), -- job position
			COALESCE(pg.name, ''), -- job level (grade name)
			COALESCE(e.join_date::text, ''),
			COALESCE(e.resigned_at::text, ''),
			COALESCE(e.employment_status::text, ''),
			COALESCE(e.end_date::text, ''),
			'', -- sign date
			COALESCE(e.email, ''),
			COALESCE(e.birth_date::text, ''),
			'', -- age (computed)
			COALESCE(e.birth_place, ''),
			COALESCE(decrypt_sensitive(e.encrypted_address_ktp), ''),
			COALESCE(e.address_domicile, ''),
			COALESCE(decrypt_sensitive(e.encrypted_npwp), ''),
			COALESCE(e.ptkp_status::text, ''),
			'', -- employee tax status
			COALESCE(e.tax_method::text, ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_account), ''),
			COALESCE(e.full_name, ''), -- bank account holder
			'', -- bpjs tk
			'', -- bpjs kesehatan
			COALESCE(decrypt_sensitive(e.encrypted_nik), ''),
			COALESCE(e.phone, ''),
			'', -- phone
			'', -- branch name
			'', -- parent branch name
			COALESCE(e.religion::text, ''),
			COALESCE(e.gender::text, ''),
			COALESCE(e.marital_status::text, ''),
			COALESCE(e.blood_type, ''),
			'', -- nationality code
			'', -- currency
			'', -- length of service (computed)
			'', -- payment schedule
			COALESCE((SELECT full_name FROM employees WHERE id = e.approval_line_id), ''), -- approval line
			'', -- manager
			COALESCE(pg.name, ''), -- grade
			'', -- class
			CASE WHEN COALESCE(e.photo_url, '') != '' THEN 'true' ELSE 'false' END,
			'', -- cost center
			'', -- cost center category
			'', -- sbu
			COALESCE(decrypt_sensitive(e.encrypted_nik), ''), -- npwp baru
			'', -- passport
			'', -- passport expiration date
			'', -- jenis dok referensi
			'', -- nomor dok referensi
			'', -- tanggal dok referensi
			'', -- tin
			''  -- ukuran baju
		FROM employees e
		LEFT JOIN departments d ON e.department_id = d.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN position_grades pg ON e.position_grade_id = pg.id
		WHERE e.deleted_at IS NULL
		ORDER BY e.full_name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.EmployeeExportFull
	for rows.Next() {
		var emp models.EmployeeExportFull
		err := rows.Scan(
			&emp.EmployeeID, &emp.FullName, &emp.Barcode,
			&emp.Organization, &emp.JobPosition, &emp.JobLevel,
			&emp.JoinDate, &emp.ResignDate, &emp.StatusEmployee,
			&emp.EndDate, &emp.SignDate,
			&emp.Email, &emp.BirthDate, &emp.Age,
			&emp.BirthPlace, &emp.CitizenIDAddress, &emp.ResidentialAddress,
			&emp.NPWP, &emp.PTKPStatus, &emp.EmployeeTaxStatus,
			&emp.TaxConfig,
			&emp.BankName, &emp.BankAccount, &emp.BankAccountHolder,
			&emp.BPJSTK, &emp.BPJSKesehatan, &emp.NIK,
			&emp.MobilePhone, &emp.Phone,
			&emp.BranchName, &emp.ParentBranchName,
			&emp.Religion, &emp.Gender, &emp.MaritalStatus,
			&emp.BloodType, &emp.NationalityCode, &emp.Currency,
			&emp.LengthOfService, &emp.PaymentSchedule,
			&emp.ApprovalLine, &emp.Manager, &emp.Grade,
			&emp.Class, &emp.ProfilePicture,
			&emp.CostCenter, &emp.CostCenterCategory, &emp.SBU,
			&emp.NPWPBaru, &emp.Passport, &emp.PassportExpirationDate,
			&emp.JenisDokReferensi, &emp.NomorDokReferensi, &emp.TanggalDokReferensi,
			&emp.TIN, &emp.UkuranBaju,
		)
		if err != nil {
			return nil, err
		}
		// Compute Age & Length of Service
		if emp.BirthDate != "" {
			if bd, err := time.Parse("2006-01-02", emp.BirthDate[:10]); err == nil {
				age := int(time.Since(bd).Hours() / 24 / 365)
				emp.Age = fmt.Sprintf("%d", age)
			}
		}
		if emp.JoinDate != "" {
			if jd, err := time.Parse("2006-01-02", emp.JoinDate[:10]); err == nil {
				dur := time.Since(jd)
				years := int(dur.Hours() / 24 / 365)
				days := int(dur.Hours()/24) % 365
				months := days / 30
				emp.LengthOfService = fmt.Sprintf("%d Year %d Month %d Day", years, months, days%30)
			}
		}
		employees = append(employees, emp)
	}
	return employees, nil
}
