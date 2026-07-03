package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"hrms-backend/internal/config"
	"hrms-backend/internal/database"

	"golang.org/x/crypto/bcrypt"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func formatCurrency(val float64) string {
	return fmt.Sprintf("Rp %.2f", val)
}

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL(), cfg.EncryptionKey); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	fmt.Println("🌱 Seeding database with employees, salaries, and salary components...")

	// Disable audit triggers temporarily
	database.Pool.Exec(ctx, `ALTER TABLE employees DISABLE TRIGGER audit_employees`)
	database.Pool.Exec(ctx, `ALTER TABLE departments DISABLE TRIGGER audit_departments`)
	database.Pool.Exec(ctx, `ALTER TABLE positions DISABLE TRIGGER audit_positions`)
	database.Pool.Exec(ctx, `ALTER TABLE loans DISABLE TRIGGER audit_loans`)
	database.Pool.Exec(ctx, `ALTER TABLE employee_salary_components DISABLE TRIGGER audit_employee_salary_components`)
	database.Pool.Exec(ctx, `ALTER TABLE employee_salary_components DISABLE TRIGGER log_salary_component_changes`)

	// Hash default passwords
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	userPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Check if admin already exists
	var adminCount int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE employee_id = 'ADMIN-001'`).Scan(&adminCount)

	var adminUUID string
	if adminCount == 0 {
		var superAdminRoleID string
		err := database.Pool.QueryRow(ctx, `SELECT id::text FROM roles WHERE slug = 'super_admin'`).Scan(&superAdminRoleID)
		if err != nil {
			log.Fatalf("No super_admin role found: %v", err)
		}

		err = database.Pool.QueryRow(ctx,
			`INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active, role_id,
			 work_schedule_id)
			 VALUES ('ADMIN-001', 'Super Admin', 'admin@company.com', $1, 'laki_laki', CURRENT_DATE, 'tetap', TRUE, $2::uuid,
			 (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1))
			 RETURNING id::text`,
			string(adminPassword), superAdminRoleID).Scan(&adminUUID)
		if err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}
		fmt.Println("✅ Admin user created: admin@company.com / admin123")
	} else {
		database.Pool.QueryRow(ctx, `SELECT id::text FROM employees WHERE employee_id = 'ADMIN-001'`).Scan(&adminUUID)
		var superAdminRoleID string
		database.Pool.QueryRow(ctx, `SELECT id::text FROM roles WHERE slug = 'super_admin'`).Scan(&superAdminRoleID)

		_, err := database.Pool.Exec(ctx,
			`UPDATE employees SET password_hash = $1, role_id = $2::uuid WHERE employee_id = 'ADMIN-001'`,
			string(adminPassword), superAdminRoleID)
		if err != nil {
			log.Printf("⚠️ Failed to update admin: %v", err)
		} else {
			fmt.Println("✅ Admin password updated: admin@company.com / admin123")
		}
	}

	// Seed Admin Base Salary
	var adminSalaryCount int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_salary_histories WHERE employee_id::text = $1`, adminUUID).Scan(&adminSalaryCount)
	if adminSalaryCount == 0 {
		_, err := database.Pool.Exec(ctx,
			`INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
			 VALUES ($1::uuid, 15000000, 0, CURRENT_DATE, 'Initial Seed')`,
			adminUUID)
		if err != nil {
			log.Printf("⚠️ Failed to seed salary history for Admin: %v", err)
		} else {
			fmt.Println("💰 Seeded base salary for Admin: Rp 15.000.000")
		}
	}

	// Load all role IDs for assignment
	roleIDs := make(map[string]string)
	for _, slug := range []string{"employee", "hr_manager", "hr_staff", "finance", "manager", "director"} {
		var id string
		err := database.Pool.QueryRow(ctx, `SELECT id::text FROM roles WHERE slug = $1`, slug).Scan(&id)
		if err != nil {
			log.Printf("⚠️ No role found for slug '%s': %v", slug, err)
		} else {
			roleIDs[slug] = id
		}
	}

	// Add missing module permissions to roles
	database.Pool.Exec(ctx, `ALTER TABLE roles DISABLE TRIGGER audit_roles`)
	for _, slug := range []string{"hr_manager", "hr_staff", "super_admin"} {
		_, err := database.Pool.Exec(ctx, `
			UPDATE roles SET permissions = permissions || '{
				"attendance_location": {"create": true, "read": true, "update": true, "delete": true},
				"position_grade": {"create": true, "read": true, "update": true, "delete": true},
				"position": {"create": true, "read": true, "update": true, "delete": true},
				"work_schedule": {"create": true, "read": true, "update": true, "delete": true},
				"shift_change": {"create": true, "read": true, "update": true, "delete": true},
				"overtime": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
				"reimbursement": {"create": true, "read": true, "update": true, "delete": true, "approve": true}
			}'::jsonb WHERE slug = $1
		`, slug)
		if err != nil {
			log.Printf("⚠️ Failed to update permissions for '%s': %v", slug, err)
		}
	}
	// Tambah permission overtime & reimbursement untuk finance (approve reimbursement, read overtime)
	for _, slug := range []string{"finance"} {
		_, err := database.Pool.Exec(ctx, `
			UPDATE roles SET permissions = permissions || '{
				"overtime": {"create": true, "read": true, "update": true},
				"reimbursement": {"create": true, "read": true, "update": true, "approve": true}
			}'::jsonb WHERE slug = $1
		`, slug)
		if err != nil {
			log.Printf("⚠️ Failed to update permissions for '%s': %v", slug, err)
		}
	}
	// Tambah permission overtime & reimbursement untuk employee
	for _, slug := range []string{"employee"} {
		_, err := database.Pool.Exec(ctx, `
			UPDATE roles SET permissions = permissions || '{
				"overtime": {"create": true, "read": true},
				"reimbursement": {"create": true, "read": true}
			}'::jsonb WHERE slug = $1
		`, slug)
		if err != nil {
			log.Printf("⚠️ Failed to update permissions for '%s': %v", slug, err)
		}
	}
	// Tambah permission attendance (create & update) untuk employee, finance, manager, director agar bisa absen
	for _, slug := range []string{"employee", "finance", "manager", "director"} {
		_, err := database.Pool.Exec(ctx, `
			UPDATE roles SET permissions = permissions || '{
				"attendance": {"create": true, "read": true, "update": true}
			}'::jsonb WHERE slug = $1
		`, slug)
		if err != nil {
			log.Printf("⚠️ Failed to update attendance permissions for '%s': %v", slug, err)
		}
	}
	database.Pool.Exec(ctx, `ALTER TABLE roles ENABLE TRIGGER audit_roles`)

	// Determine role based on department
	deptRoleMap := map[string]string{
		"Sumber Daya Manusia": "hr_manager",
		"Keuangan":            "finance",
	}

	// Create departments
	depts := []struct {
		Name string
		Code string
	}{
		{"Teknologi Informasi", "IT"},
		{"Keuangan", "FIN"},
		{"Penjualan", "SALES"},
		{"Sumber Daya Manusia", "HR"},
		{"Pemasaran", "MKT"},
	}

	deptIDs := make(map[string]string)
	for _, d := range depts {
		var id string
		err := database.Pool.QueryRow(ctx,
			`SELECT id::text FROM departments WHERE code = $1 LIMIT 1`, d.Code).Scan(&id)
		if err != nil {
			err = database.Pool.QueryRow(ctx,
				`INSERT INTO departments (name, code, work_schedule_id) VALUES ($1, $2, (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)) RETURNING id::text`,
				d.Name, d.Code).Scan(&id)
			if err != nil {
				log.Printf("⚠️ Could not create department %s: %v", d.Name, err)
				continue
			}
		}
		deptIDs[d.Name] = id
	}

	// Create positions
	positions := []struct {
		Name     string
		DeptName string
	}{
		{"Staff IT", "Teknologi Informasi"},
		{"Finance Staff", "Keuangan"},
		{"Sales Executive", "Penjualan"},
		{"HR Staff", "Sumber Daya Manusia"},
		{"Sales", "Penjualan"},
		{"Marketing", "Pemasaran"},
		{"Accounting", "Keuangan"},
	}

	posIDs := make(map[string]string)
	for _, p := range positions {
		var id string
		err := database.Pool.QueryRow(ctx,
			`SELECT id::text FROM positions WHERE name = $1 LIMIT 1`, p.Name).Scan(&id)
		if err != nil {
			deptID := deptIDs[p.DeptName]
			err = database.Pool.QueryRow(ctx,
				`INSERT INTO positions (name, department_id) VALUES ($1, $2::uuid) RETURNING id::text`,
				p.Name, deptID).Scan(&id)
			if err != nil {
				log.Printf("⚠️ Could not create position %s: %v", p.Name, err)
				continue
			}
		}
		posIDs[p.Name] = id
	}

	// Test Employees default salaries and components schema
	type componentSeed struct {
		Name   string
		Amount float64
	}

	testEmployees := []struct {
		ID              string
		Name            string
		Email           string
		Status          string
		Gender          string
		Position        string
		DeptName        string
		ContractEndDate *time.Time
		BaseSalary      float64
		DailyWage       float64
		Allowances      []componentSeed
		Deductions      []componentSeed
	}{
		{
			"EMP-001", "Budi Hartono", "budi@company.com", "tetap", "laki_laki", "Staff IT", "Teknologi Informasi", nil,
			8500000, 0,
			[]componentSeed{{"Tunjangan Makan", 500000}, {"Tunjangan Transport", 300000}},
			[]componentSeed{{"BPJS Tambahan", 100000}},
		},
		{
			"EMP-002", "Siti Rahayu", "siti@company.com", "tetap", "perempuan", "Finance Staff", "Keuangan", nil,
			9000000, 0,
			[]componentSeed{{"Tunjangan Jabatan", 1500000}, {"Tunjangan Komunikasi", 200000}},
			nil,
		},
		{
			"EMP-003", "Andi Wijaya", "andi@company.com", "kontrak", "laki_laki", "Sales Executive", "Penjualan", timePtr(time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)),
			6000000, 0,
			[]componentSeed{{"Tunjangan Sales", 1000000}},
			nil,
		},
		{
			"EMP-004", "Dewi Lestari", "dewi@company.com", "kontrak", "perempuan", "HR Staff", "Sumber Daya Manusia", timePtr(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
			7500000, 0,
			[]componentSeed{{"Tunjangan Makan", 500000}},
			nil,
		},
		{
			"EMP-005", "Rudi Hartono", "rudi@company.com", "percobaan", "laki_laki", "Staff IT", "Teknologi Informasi", nil,
			7000000, 0,
			[]componentSeed{{"Tunjangan Makan", 500000}},
			nil,
		},
		{
			"EMP-006", "Ahmad Fauzi", "ahmad@company.com", "percobaan", "laki_laki", "Sales", "Penjualan", nil,
			0, 200000, // harian
			[]componentSeed{{"Tunjangan Transport Harian", 25000}},
			nil,
		},
		{
			"EMP-007", "Rina Marlina", "rina@company.com", "tetap", "perempuan", "Marketing", "Pemasaran", nil,
			6500000, 0,
			[]componentSeed{{"Tunjangan Makan", 500000}},
			nil,
		},
		{
			"EMP-008", "Hendra Gunawan", "hendra@company.com", "tetap", "laki_laki", "Accounting", "Keuangan", nil,
			8000000, 0,
			[]componentSeed{{"Tunjangan Makan", 500000}, {"Tunjangan Transport", 300000}},
			nil,
		},
	}

	createdCount := 0
	for _, emp := range testEmployees {
		roleSlug := deptRoleMap[emp.DeptName]
		if roleSlug == "" {
			roleSlug = "employee"
		}
		roleID := roleIDs[roleSlug]

		posID := posIDs[emp.Position]
		deptID := deptIDs[emp.DeptName]
		joinDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		if emp.Status == "percobaan" {
			joinDate = time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
		}

		var existingID string
		err := database.Pool.QueryRow(ctx,
			`SELECT id::text FROM employees WHERE employee_id = $1 LIMIT 1`, emp.ID).Scan(&existingID)

		var empUUID string
		if err == nil {
			// Employee already exists
			empUUID = existingID
			// Update daily_wage and role if needed
			_, err = database.Pool.Exec(ctx,
				`UPDATE employees SET role_id = NULLIF($1, '')::uuid, daily_wage = NULLIF($2, 0) WHERE id = $3::uuid`,
				roleID, emp.DailyWage, empUUID)
			if err != nil {
				log.Printf("⚠️ Failed to update existing employee %s: %v", emp.Name, err)
			}
		} else {
			// Insert new employee
			query := `INSERT INTO employees (employee_id, full_name, email, password_hash, gender,
				join_date, employment_status, is_active, role_id, position_id, department_id,
				end_date, daily_wage, work_schedule_id)
				VALUES ($1, $2, $3, $4, $5::gender_type, $6, $7::employment_status, TRUE,
				NULLIF($8, '')::uuid, NULLIF($9, '')::uuid, NULLIF($10, '')::uuid,
				$11::timestamptz, NULLIF($12, 0),
				(SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1))
				RETURNING id::text`

			err = database.Pool.QueryRow(ctx, query,
				emp.ID, emp.Name, emp.Email, string(userPassword),
				emp.Gender, joinDate, emp.Status,
				roleID, posID, deptID, emp.ContractEndDate, emp.DailyWage).Scan(&empUUID)
			if err != nil {
				log.Printf("⚠️ Could not create employee %s: %v", emp.Name, err)
				continue
			}
			createdCount++
			fmt.Printf("✅ Employee created: %s (%s) / password123 — Role: %s\n", emp.Name, emp.Email, roleSlug)
		}

		// Seed salary history if not exists
		var salaryCount int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_salary_histories WHERE employee_id::text = $1`, empUUID).Scan(&salaryCount)
		if salaryCount == 0 {
			_, err = database.Pool.Exec(ctx,
				`INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
				 VALUES ($1::uuid, $2, $3, CURRENT_DATE, 'Initial Seed')`,
				empUUID, emp.BaseSalary, emp.DailyWage)
			if err != nil {
				log.Printf("⚠️ Failed to seed salary history for %s: %v", emp.Name, err)
			} else {
				fmt.Printf("💰 Seeded base salary for %s: %s\n", emp.Name, formatCurrency(emp.BaseSalary))
			}
		}

		// Seed salary components if not exists
		var componentCount int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_salary_components WHERE employee_id::text = $1`, empUUID).Scan(&componentCount)
		if componentCount == 0 {
			// Seed allowances
			for _, allow := range emp.Allowances {
				_, err = database.Pool.Exec(ctx,
					`INSERT INTO employee_salary_components (employee_id, component_name, component_type, amount, is_active, effective_date)
					 VALUES ($1::uuid, $2, 'allowance', $3, TRUE, CURRENT_DATE)`,
					empUUID, allow.Name, allow.Amount)
				if err != nil {
					log.Printf("⚠️ Failed to seed allowance component %s for %s: %v", allow.Name, emp.Name, err)
				}
			}
			// Seed deductions
			for _, deduct := range emp.Deductions {
				_, err = database.Pool.Exec(ctx,
					`INSERT INTO employee_salary_components (employee_id, component_name, component_type, amount, is_active, effective_date)
					 VALUES ($1::uuid, $2, 'deduction', $3, TRUE, CURRENT_DATE)`,
					empUUID, deduct.Name, deduct.Amount)
				if err != nil {
					log.Printf("⚠️ Failed to seed deduction component %s for %s: %v", deduct.Name, emp.Name, err)
				}
			}
			fmt.Printf("📦 Seeded active salary components for %s\n", emp.Name)
		}
	}

	// ─────────────────────────────────────────────────────────
	// 📊 SEED KPI TEMPLATES & INDICATORS
	// ─────────────────────────────────────────────────────────
	fmt.Println("\n📊 Seeding KPI templates...")
	
	// KPI Template 1: Staff IT (yearly 2026)
	var kpiTemplateCount int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM kpi_templates WHERE title = 'KPI Staff IT - 2026'`).Scan(&kpiTemplateCount)
	if kpiTemplateCount == 0 {
		var itDeptID string
		var itPosID string
		database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE code = 'IT' LIMIT 1`).Scan(&itDeptID)
		database.Pool.QueryRow(ctx, `SELECT id::text FROM positions WHERE name = 'Staff IT' LIMIT 1`).Scan(&itPosID)
		
		if itDeptID != "" && itPosID != "" {
			var templateID string
			err := database.Pool.QueryRow(ctx, `
				INSERT INTO kpi_templates (title, position_id, department_id, period_type, year, description, is_active)
				VALUES ('KPI Staff IT - 2026', $1::uuid, $2::uuid, 'yearly', 2026, 'Template KPI untuk Staff IT tahun 2026', TRUE)
				RETURNING id::text
			`, itPosID, itDeptID).Scan(&templateID)
			
			if err != nil {
				log.Printf("⚠️ Failed to create KPI template: %v", err)
			} else {
				fmt.Println("✅ Seeded KPI template: KPI Staff IT - 2026")
				
				// Seed indicators
				indicators := []struct {
					Name   string
					Target float64
					Weight float64
					Unit   string
				}{
					{"Penyelesaian Project IT", 100, 40, "%"},
					{"Server & System Uptime", 99.9, 30, "%"},
					{"Waktu Respon Penyelesaian Tiket", 15, 30, "menit"},
				}
				
				for i, ind := range indicators {
					_, err = database.Pool.Exec(ctx, `
						INSERT INTO kpi_indicators (kpi_template_id, name, target, weight, measurement_unit, sort_order)
						VALUES ($1::uuid, $2, $3, $4, $5, $6)
					`, templateID, ind.Name, ind.Target, ind.Weight, ind.Unit, i+1)
					if err != nil {
						log.Printf("⚠️ Failed to seed indicator %s: %v", ind.Name, err)
					}
				}
				fmt.Println("📊 Seeded indicators for KPI Staff IT")
			}
		}
	}

	// KPI Template 2: Finance Staff (yearly 2026)
	var kpiTemplateCount2 int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM kpi_templates WHERE title = 'KPI Finance Staff - 2026'`).Scan(&kpiTemplateCount2)
	if kpiTemplateCount2 == 0 {
		var finDeptID string
		var finPosID string
		database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE code = 'FIN' LIMIT 1`).Scan(&finDeptID)
		database.Pool.QueryRow(ctx, `SELECT id::text FROM positions WHERE name = 'Finance Staff' LIMIT 1`).Scan(&finPosID)
		
		if finDeptID != "" && finPosID != "" {
			var templateID string
			err := database.Pool.QueryRow(ctx, `
				INSERT INTO kpi_templates (title, position_id, department_id, period_type, year, description, is_active)
				VALUES ('KPI Finance Staff - 2026', $1::uuid, $2::uuid, 'yearly', 2026, 'Template KPI untuk Staff Finance tahun 2026', TRUE)
				RETURNING id::text
			`, finPosID, finDeptID).Scan(&templateID)
			
			if err != nil {
				log.Printf("⚠️ Failed to create KPI template 2: %v", err)
			} else {
				fmt.Println("✅ Seeded KPI template: KPI Finance Staff - 2026")
				
				// Seed indicators
				indicators := []struct {
					Name   string
					Target float64
					Weight float64
					Unit   string
				}{
					{"Ketepatan Laporan Keuangan", 100, 50, "%"},
					{"Kepatuhan Audit Pajak", 100, 30, "%"},
					{"Waktu Pemrosesan Reimbursement", 2, 20, "hari"},
				}
				
				for i, ind := range indicators {
					_, err = database.Pool.Exec(ctx, `
						INSERT INTO kpi_indicators (kpi_template_id, name, target, weight, measurement_unit, sort_order)
						VALUES ($1::uuid, $2, $3, $4, $5, $6)
					`, templateID, ind.Name, ind.Target, ind.Weight, ind.Unit, i+1)
					if err != nil {
						log.Printf("⚠️ Failed to seed indicator %s: %v", ind.Name, err)
					}
				}
				fmt.Println("📊 Seeded indicators for KPI Finance Staff")
			}
		}
	}

	// Disable audit triggers for overtime & reimbursement (seed runs outside app context)
	database.Pool.Exec(ctx, `ALTER TABLE overtime_requests DISABLE TRIGGER audit_overtime_requests`)
	database.Pool.Exec(ctx, `ALTER TABLE reimbursements DISABLE TRIGGER audit_reimbursements`)

	// ─────────────────────────────────────────────────────────
	// 🌙 TAMBAH DATA SEED OVERTIME (LEMBUR)
	// ─────────────────────────────────────────────────────────
	fmt.Println("\n🌙 Seeding overtime requests...")

	type overtimeSeed struct {
		EmployeeID   string
		Date         string
		StartTime    string
		EndTime      string
		TotalHours   float64
		OvertimeType string
		Reason       string
		Status       string
	}

	overtimeData := []overtimeSeed{
		{"EMP-001", "2026-06-20", "2026-06-20T17:00:00Z", "2026-06-20T19:00:00Z", 2.0, "weekday", "Menyelesaikan project migrasi server", "approved"},
		{"EMP-002", "2026-06-21", "2026-06-21T18:00:00Z", "2026-06-21T21:30:00Z", 3.5, "weekday", "Closing laporan keuangan bulanan", "pending"},
		{"EMP-003", "2026-06-22", "2026-06-22T08:00:00Z", "2026-06-22T16:00:00Z", 8.0, "weekend", "Event peluncuran produk baru", "pending"},
		{"EMP-004", "2026-06-23", "2026-06-23T17:00:00Z", "2026-06-23T18:30:00Z", 1.5, "weekday", "Persiapan briefing HR mingguan", "approved"},
		{"EMP-005", "2026-06-24", "2026-06-24T08:00:00Z", "2026-06-24T14:00:00Z", 6.0, "holiday", "Maintenance sistem darurat", "pending"},
	}

	for _, ot := range overtimeData {
		var existingCount int
		database.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM overtime_requests
			 WHERE employee_id = (SELECT id FROM employees WHERE employee_id = $1)
			 AND date = $2`,
			ot.EmployeeID, ot.Date).Scan(&existingCount)

		if existingCount == 0 {
			status := ot.Status
			now := time.Now().UTC().Format(time.RFC3339)
			// Build approval_trail - pake adminUUID langsung (bukan SQL subquery biar valid JSON)
			approvalEntry := "[]"
			if status == "approved" {
				approvalEntry = fmt.Sprintf(`[{"status":"approved","approver_id":"%s","date":"%s"}]`, adminUUID, now)
			}
			_, err := database.Pool.Exec(ctx,
				`INSERT INTO overtime_requests (employee_id, date, start_time, end_time, total_hours, overtime_type, reason, status, approval_trail)
				 VALUES (
					(SELECT id FROM employees WHERE employee_id = $1),
					$2::date, $3::timestamptz, $4::timestamptz, $5, $6::overtime_type,
					$7, $8::leave_status, $9::jsonb
				)`,
				ot.EmployeeID, ot.Date, ot.StartTime, ot.EndTime,
				ot.TotalHours, ot.OvertimeType, ot.Reason, status, approvalEntry)
			if err != nil {
				log.Printf("⚠️ Failed to seed overtime for %s: %v", ot.EmployeeID, err)
			} else {
				fmt.Printf("⏰ Overtime: %s — %s (%.1f jam, %s) [%s]\n", ot.EmployeeID, ot.Reason, ot.TotalHours, ot.OvertimeType, status)
			}
		} else {
			fmt.Printf("⏰ Overtime already exists for %s on %s, skipping\n", ot.EmployeeID, ot.Date)
		}
	}

	// ─────────────────────────────────────────────────────────
	// 💰 TAMBAH DATA SEED REIMBURSEMENT
	// ─────────────────────────────────────────────────────────
	fmt.Println("\n💰 Seeding reimbursement requests...")

	type reimbursementSeed struct {
		EmployeeID string
		Type       string
		Amount     float64
		Desc       string
		ReceiptURL string
		Status     string
	}

	reimbursementData := []reimbursementSeed{
		{"EMP-001", "medical", 500000, "Biaya berobat di Klinik Sehat", "/uploads/receipts/receipt-medical-001.jpg", "approved"},
		{"EMP-002", "travel", 1200000, "Tiket pesawat dinas Jakarta-Surabaya", "", "pending"},
		{"EMP-003", "training", 2500000, "Biaya kursus online Data Analytics", "/uploads/receipts/receipt-training-003.jpg", "approved"},
		{"EMP-004", "supplies", 350000, "Pembelian ATK kantor", "", "pending"},
		{"EMP-008", "other", 750000, "Biaya sertifikasi tahunan akuntansi", "/uploads/receipts/receipt-cert-008.jpg", "approved_paid"}, // approved + paid
	}

	for _, rm := range reimbursementData {
		var existingCount int
		database.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM reimbursements
			 WHERE employee_id = (SELECT id FROM employees WHERE employee_id = $1)
			 AND description = $2`,
			rm.EmployeeID, rm.Desc).Scan(&existingCount)

		if existingCount == 0 {
			isPaid := rm.Status == "approved_paid"
			dbStatus := rm.Status
			if isPaid {
				dbStatus = "approved" // leave_status gak punya 'paid', pake 'approved' + paid_at/paid_by
			}
			now := time.Now().UTC().Format(time.RFC3339)

			// Build receipt_urls array
			var receiptURLs []string
			if rm.ReceiptURL != "" {
				receiptURLs = []string{rm.ReceiptURL}
			}

			// Build approval_trail - pake adminUUID langsung
			var approvalEntry string
			if dbStatus == "approved" || isPaid {
				approvalEntry = fmt.Sprintf(`[{"status":"approved","approver_id":"%s","date":"%s"}]`, adminUUID, now)
			} else {
				approvalEntry = "[]"
			}

			_, err := database.Pool.Exec(ctx, `
				INSERT INTO reimbursements (employee_id, type, amount, description, receipt_urls, status, approval_trail, paid_by, paid_at)
				VALUES (
					(SELECT id FROM employees WHERE employee_id = $1),
					$2::reimbursement_type, $3, $4,
					$5::text[], $6::leave_status,
					$7::jsonb,
					CASE WHEN $8 THEN $9::uuid ELSE NULL END,
					CASE WHEN $8 THEN $10::timestamptz ELSE NULL END
				)`,
				rm.EmployeeID, rm.Type, rm.Amount, rm.Desc,
				receiptURLs, dbStatus, approvalEntry,
				isPaid, adminUUID, now)
			if err != nil {
				log.Printf("⚠️ Failed to seed reimbursement for %s: %v", rm.EmployeeID, err)
			} else {
				fmt.Printf("💰 Reimbursement: %s — %s (%s) [%s]\n", rm.EmployeeID, rm.Desc, formatCurrency(rm.Amount), dbStatus)
			}
		} else {
			fmt.Printf("💰 Reimbursement already exists for %s — %s, skipping\n", rm.EmployeeID, rm.Desc)
		}
	}

	// Re-enable audit triggers
	database.Pool.Exec(ctx, `ALTER TABLE overtime_requests ENABLE TRIGGER audit_overtime_requests`)
	database.Pool.Exec(ctx, `ALTER TABLE reimbursements ENABLE TRIGGER audit_reimbursements`)
	database.Pool.Exec(ctx, `ALTER TABLE employees ENABLE TRIGGER audit_employees`)

	// ──────────────────────────────────────────────────────────────
	// Seed Loans (Pinjaman)
	// ──────────────────────────────────────────────────────────────
	fmt.Println("\n💳 Seeding loan data...")
	type loanSeed struct {
		EmployeeID string
		LoanType   string
		Amount     float64
		InstCount  int
		Rate       float64
		Method     string
		Purpose    string
		Status     string // pending | approved | active | rejected | settled
	}
	loanSeeds := []loanSeed{
		{"EMP-001", "regular", 5000000, 10, 2.0, "payroll_deduction", "Kebutuhan renovasi rumah", "active"},
		{"EMP-002", "emergency", 3000000, 6, 0.0, "payroll_deduction", "Biaya pengobatan darurat", "approved"},
		{"EMP-003", "education", 8000000, 12, 1.5, "manual_transfer", "Biaya kuliah S2", "pending"},
		{"EMP-004", "regular", 2000000, 5, 2.0, "payroll_deduction", "Beli laptop kerja", "rejected"},
		{"EMP-005", "regular", 10000000, 24, 1.5, "manual_transfer", "Modal usaha sampingan", "pending"},
	}

	for _, ls := range loanSeeds {
		var loanCount int
		database.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM loans l
			 JOIN employees e ON e.id = l.employee_id
			 WHERE e.employee_id = $1 AND l.deleted_at IS NULL`,
			ls.EmployeeID).Scan(&loanCount)

		if loanCount == 0 {
			totalInterest := ls.Amount * ls.Rate / 100.0
			totalAmount := ls.Amount + totalInterest
			installmentAmount := totalAmount / float64(ls.InstCount)

			approvalTrail := "[]"
			if ls.Status == "approved" || ls.Status == "active" {
				approvalTrail = fmt.Sprintf(`[{"status":"approved","approver_id":"%s","date":"%s"}]`, adminUUID, now)
			} else if ls.Status == "rejected" {
				approvalTrail = fmt.Sprintf(`[{"status":"rejected","approver_id":"%s","date":"%s","reason":"Tidak memenuhi syarat"}]`, adminUUID, now)
			}

			disbursedAt := "NULL"
			disbursedBy := "NULL"
			if ls.Status == "active" {
				disbursedAt = fmt.Sprintf("'%s'::date", time.Now().AddDate(0, 0, -15).Format("2006-01-02"))
				disbursedBy = fmt.Sprintf("'%s'::uuid", adminUUID)
			}

			query := fmt.Sprintf(`
				INSERT INTO loans (employee_id, loan_type, amount, interest_rate,
					total_interest, total_amount, installment_count, installment_amount,
					payment_method, remaining_balance, purpose, status, approval_trail,
					disbursed_at, disbursed_by)
				SELECT e.id, $1::loan_type, $2, $3, $4, $5, $6, $7,
					$8::loan_payment_method, $5, $9, $10::loan_status, $11::jsonb,
					%s, %s
				FROM employees e WHERE e.employee_id = $12`,
				disbursedAt, disbursedBy)

			_, err := database.Pool.Exec(ctx, query,
				ls.LoanType, ls.Amount, ls.Rate, totalInterest, totalAmount,
				ls.InstCount, installmentAmount, ls.Method, ls.Purpose, ls.Status,
				approvalTrail, ls.EmployeeID)
			if err != nil {
				log.Printf("⚠️ Failed to seed loan for %s: %v", ls.EmployeeID, err)
			} else {
				fmt.Printf("💳 Loan: %s — %s %s [%s]\n", ls.EmployeeID, ls.LoanType, formatCurrency(ls.Amount), ls.Status)
			}
		} else {
			fmt.Printf("💳 Loan already exists for %s, skipping\n", ls.EmployeeID)
		}
	}

	// ──────────────────────────────────────────────────────────────
	// Seed Leave Requests
	// ──────────────────────────────────────────────────────────────
	fmt.Println("\n📅 Seeding leave request data...")
	type leaveSeed struct {
		EmployeeID  string
		LeaveType   string // slug: annual, sick, etc.
		StartDate   string
		EndDate     string
		Days        int
		Reason      string
		Status      string // pending | approved | rejected | cancelled
	}
	leaveSeeds := []leaveSeed{
		{"EMP-001", "tahunan", "2026-07-07", "2026-07-09", 3, "Liburan keluarga", "approved"},
		{"EMP-002", "sakit", "2026-06-23", "2026-06-24", 2, "Demam dan flu", "approved"},
		{"EMP-003", "tahunan", "2026-07-14", "2026-07-15", 2, "Mengurus keperluan pribadi", "pending"},
		{"EMP-004", "tahunan", "2026-07-21", "2026-07-22", 2, "Acara pernikahan saudara", "pending"},
		{"EMP-005", "sakit", "2026-06-30", "2026-06-30", 1, "Sakit kepala berat", "approved"},
	}

	// Disable leave triggers to avoid side effects during seed
	database.Pool.Exec(ctx, `ALTER TABLE leave_requests DISABLE TRIGGER ALL`)

	for _, ls := range leaveSeeds {
		var leaveCount int
		database.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM leave_requests lr
			 JOIN employees e ON e.id = lr.employee_id
			 WHERE e.employee_id = $1 AND lr.start_date = $2::date AND lr.deleted_at IS NULL`,
			ls.EmployeeID, ls.StartDate).Scan(&leaveCount)

		if leaveCount == 0 {
			approvalTrail := "[]"
			if ls.Status == "approved" {
				approvalTrail = fmt.Sprintf(`[{"action":"approved","by":"%s","at":"%s"}]`, adminUUID, now)
			} else if ls.Status == "rejected" {
				approvalTrail = fmt.Sprintf(`[{"action":"rejected","by":"%s","at":"%s","reason":"Tidak memenuhi syarat"}]`, adminUUID, now)
			}

			_, err := database.Pool.Exec(ctx, `
				INSERT INTO leave_requests
					(employee_id, leave_type_id, start_date, end_date, total_days, reason, status, approval_trail)
				SELECT
					e.id,
					lt.id,
					$3::date, $4::date, $5,
					$6, $7::leave_status,
					$8::jsonb
				FROM employees e
				JOIN leave_types lt ON lt.code = $2
				WHERE e.employee_id = $1`,
				ls.EmployeeID, ls.LeaveType, ls.StartDate, ls.EndDate,
				ls.Days, ls.Reason, ls.Status, approvalTrail)
			if err != nil {
				log.Printf("⚠️ Failed to seed leave for %s: %v", ls.EmployeeID, err)
			} else {
				fmt.Printf("📅 Leave: %s — %s %s to %s [%s]\n", ls.EmployeeID, ls.LeaveType, ls.StartDate, ls.EndDate, ls.Status)
			}
		} else {
			fmt.Printf("📅 Leave already exists for %s on %s, skipping\n", ls.EmployeeID, ls.StartDate)
		}
	}

	database.Pool.Exec(ctx, `ALTER TABLE leave_requests ENABLE TRIGGER ALL`)
	database.Pool.Exec(ctx, `ALTER TABLE departments ENABLE TRIGGER audit_departments`)
	database.Pool.Exec(ctx, `ALTER TABLE positions ENABLE TRIGGER audit_positions`)
	database.Pool.Exec(ctx, `ALTER TABLE loans ENABLE TRIGGER audit_loans`)
	database.Pool.Exec(ctx, `ALTER TABLE employee_salary_components ENABLE TRIGGER audit_employee_salary_components`)
	database.Pool.Exec(ctx, `ALTER TABLE employee_salary_components ENABLE TRIGGER log_salary_component_changes`)

	// Seed sample KPI Review for Budi
	var kpiReviewCount int
	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM kpi_reviews kr
		JOIN employees e ON e.id = kr.employee_id
		WHERE e.employee_id = 'EMP-001'
	`).Scan(&kpiReviewCount)
	
	if kpiReviewCount == 0 {
		var budiUUID string
		var templateID string
		database.Pool.QueryRow(ctx, `SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`).Scan(&budiUUID)
		database.Pool.QueryRow(ctx, `SELECT id::text FROM kpi_templates WHERE title = 'KPI Staff IT - 2026' LIMIT 1`).Scan(&templateID)
		
		if budiUUID != "" && templateID != "" {
			_, err := database.Pool.Exec(ctx, `
				INSERT INTO kpi_reviews (employee_id, kpi_template_id, period, year, status, self_score, manager_score, final_score, final_category)
				VALUES ($1::uuid, $2::uuid, 'yearly', 2026, 'completed', 85.00, 88.00, 87.50, 'meets')
			`, budiUUID, templateID)
			if err != nil {
				log.Printf("⚠️ Failed to seed KPI review for Budi: %v", err)
			} else {
				fmt.Println("📊 Seeded completed KPI review for Budi Hartono")
			}
		}
	}

	// ============================================================
	// 📍 SEED ATTENDANCE LOCATIONS
	// ============================================================
	fmt.Println("\n📍 Seeding attendance locations...")
	attendanceLocations := []struct {
		Name    string
		Address string
		Lat     float64
		Lng     float64
		Radius  int
	}{
		{"Kantor Pusat", "Jl. Jenderal Sudirman No. 1, Jakarta Pusat", -6.2088, 106.8456, 100},
		{"Cabang Surabaya", "Jl. Basuki Rahmat No. 10, Surabaya", -7.2575, 112.7521, 100},
	}
	for _, loc := range attendanceLocations {
		var count int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM attendance_locations WHERE name = $1`, loc.Name).Scan(&count)
		if count == 0 {
			_, err := database.Pool.Exec(ctx, `
				INSERT INTO attendance_locations (name, address, latitude, longitude, radius_meters, is_active)
				VALUES ($1, $2, $3, $4, $5, TRUE)
			`, loc.Name, loc.Address, loc.Lat, loc.Lng, loc.Radius)
			if err != nil {
				log.Printf("⚠️ Failed to seed attendance location %s: %v", loc.Name, err)
			} else {
				fmt.Printf("📍 Attendance location created: %s\n", loc.Name)
			}
		}
	}

	// ============================================================
	// 📄 SEED EMPLOYEE DOCUMENTS
	// ============================================================
	fmt.Println("\n📄 Seeding employee documents...")
	type docSeed struct {
		EmployeeID string
		DocType    string
		FileName   string
		Title      string
		Status     string
	}
	docSeeds := []docSeed{
		{"EMP-001", "ktp", "ktp_budi.pdf", "KTP Budi Hartono", "verified"},
		{"EMP-001", "ijazah", "ijazah_budi.pdf", "Ijazah S1 Budi Hartono", "verified"},
		{"EMP-002", "ktp", "ktp_siti.pdf", "KTP Siti Rahayu", "verified"},
		{"EMP-003", "ktp", "ktp_andi.pdf", "KTP Andi Wijaya", "pending"},
		{"EMP-004", "ktp", "ktp_dewi.pdf", "KTP Dewi Lestari", "pending"},
		{"EMP-006", "ktp", "ktp_ahmad.pdf", "KTP Ahmad Fauzi", "pending"},
	}
	for _, d := range docSeeds {
		var count int
		database.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM employee_documents ed
			JOIN employees e ON e.id = ed.employee_id
			WHERE e.employee_id = $1 AND ed.title = $2
		`, d.EmployeeID, d.Title).Scan(&count)
		if count == 0 {
			status := d.Status
			verifiedBy := "NULL"
			verifiedAt := "NULL"
			if status == "verified" {
				verifiedBy = fmt.Sprintf("'%s'::uuid", adminUUID)
				verifiedAt = fmt.Sprintf("'%s'", now)
			}
			fileURL := "/uploads/documents/" + d.FileName
			query := fmt.Sprintf(`
				INSERT INTO employee_documents (employee_id, doc_type, file_name, file_url, file_size, mime_type, title, status, verified_by, verified_at)
				SELECT e.id, $1::doc_type, $2, $3, 102400, 'application/pdf', $4, $5::doc_status, %s, %s
				FROM employees e WHERE e.employee_id = $6
			`, verifiedBy, verifiedAt)
			_, err := database.Pool.Exec(ctx, query, d.DocType, d.FileName, fileURL, d.Title, status, d.EmployeeID)
			if err != nil {
				log.Printf("⚠️ Failed to seed doc %s: %v", d.Title, err)
			} else {
				fmt.Printf("📄 Document created: %s\n", d.Title)
			}
		}
	}

	// ============================================================
	// 🆘 SEED EMPLOYEE EMERGENCY CONTACTS
	// ============================================================
	fmt.Println("\n🆘 Seeding emergency contacts...")
	type emergencySeed struct {
		EmployeeID   string
		Name         string
		Relationship string
		Phone        string
	}
	emergencySeeds := []emergencySeed{
		{"EMP-001", "Sari Dewi", "Istri", "081234567890"},
		{"EMP-002", "Ahmad Rahayu", "Suami", "081234567891"},
		{"EMP-003", "Maya Wijaya", "Istri", "081234567892"},
		{"EMP-008", "Dian Gunawan", "Istri", "081234567893"},
	}
	for _, ec := range emergencySeeds {
		var count int
		database.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM employee_emergency_contacts eec
			JOIN employees e ON e.id = eec.employee_id
			WHERE e.employee_id = $1 AND eec.name = $2
		`, ec.EmployeeID, ec.Name).Scan(&count)
		if count == 0 {
			_, err := database.Pool.Exec(ctx, `
				INSERT INTO employee_emergency_contacts (employee_id, name, relationship, phone, is_primary)
				SELECT e.id, $2, $3, $4, TRUE
				FROM employees e WHERE e.employee_id = $1
			`, ec.EmployeeID, ec.Name, ec.Relationship, ec.Phone)
			if err != nil {
				log.Printf("⚠️ Failed to seed emergency contact for %s: %v", ec.EmployeeID, err)
			} else {
				fmt.Printf("🆘 Emergency contact: %s for %s\n", ec.Name, ec.EmployeeID)
			}
		}
	}

	// ============================================================
	// 📅 SEED EMPLOYEE SCHEDULES (template-based, priority 1)
	// ============================================================
	fmt.Println("\n📅 Seeding employee schedules...")
	// Get first schedule template ID
	var schedTemplateID string
	database.Pool.QueryRow(ctx, `SELECT id::text FROM schedule_templates LIMIT 1`).Scan(&schedTemplateID)
	if schedTemplateID != "" {
		empScheduleEmployees := []string{"EMP-001", "EMP-002", "EMP-003", "EMP-004", "EMP-005", "EMP-006", "EMP-007", "EMP-008"}
		for _, eid := range empScheduleEmployees {
			var count int
			database.Pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM employee_schedules es
				JOIN employees e ON e.id = es.employee_id
				WHERE e.employee_id = $1
			`, eid).Scan(&count)
			if count == 0 {
				// Assign template for workdays (Mon-Fri, day_of_week 1-5)
				for dow := 1; dow <= 5; dow++ {
					_, err := database.Pool.Exec(ctx, `
						INSERT INTO employee_schedules (employee_id, template_id, day_of_week, priority, is_active, effective_from)
						SELECT e.id, $2::uuid, $3, 1, TRUE, CURRENT_DATE
						FROM employees e WHERE e.employee_id = $1
					`, eid, schedTemplateID, dow)
					if err != nil {
						log.Printf("⚠️ Failed to seed schedule for %s dow=%d: %v", eid, dow, err)
					}
				}
				fmt.Printf("📅 Employee schedule assigned: %s (Mon-Fri)\n", eid)
			}
		}
	}

	// ============================================================
	// 💰 SEED PAYROLL PERIODS + ITEMS
	// ============================================================
	fmt.Println("\n💰 Seeding payroll periods...")
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())
	// Create last 3 months of payroll periods
	for offset := 3; offset >= 1; offset-- {
		month := currentMonth - offset
		year := currentYear
		if month <= 0 {
			month += 12
			year--
		}
		periodName := fmt.Sprintf("%s %d", time.Month(month).String(), year)
		startDate := fmt.Sprintf("%04d-%02d-01", year, month)
		endDate := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

		var periodCount int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM payroll_periods WHERE period_name = $1`, periodName).Scan(&periodCount)
		if periodCount == 0 {
			var periodID string
			status := "completed" // Mark older periods as completed
			if offset == 1 {
				status = "approved" // Most recent is approved but maybe not paid
			}
			paidBy := "NULL"
			paidAt := "NULL"
			approvedBy := "NULL"
			approvedAt := "NULL"
			if status == "completed" {
				approvedBy = fmt.Sprintf("'%s'::uuid", adminUUID)
				approvedAt = fmt.Sprintf("'%s'", now)
				paidBy = fmt.Sprintf("'%s'::uuid", adminUUID)
				paidAt = fmt.Sprintf("'%s'", now)
			} else if status == "approved" {
				approvedBy = fmt.Sprintf("'%s'::uuid", adminUUID)
				approvedAt = fmt.Sprintf("'%s'", now)
			}
			query := fmt.Sprintf(`
				INSERT INTO payroll_periods (month, year, period_name, start_date, end_date, status,
					approved_by, approved_at, paid_by, paid_at, created_by)
				VALUES ($1, $2, $3, $4::date, $5::date, $6::payroll_status,
					%s, %s, %s, %s, $7::uuid)
				RETURNING id::text
			`, approvedBy, approvedAt, paidBy, paidAt)
			err := database.Pool.QueryRow(ctx, query,
				month, year, periodName, startDate, endDate, status, adminUUID).Scan(&periodID)
			if err != nil {
				log.Printf("⚠️ Failed to create payroll period %s: %v", periodName, err)
				continue
			}
			fmt.Printf("💰 Payroll period created: %s [%s]\n", periodName, status)

			// Seed payroll items for each employee
			empIDs := []string{"EMP-001", "EMP-002", "EMP-003", "EMP-004", "EMP-005", "EMP-006", "EMP-007", "EMP-008"}
			for _, eid := range empIDs {
				var itemCount int
				database.Pool.QueryRow(ctx, `
					SELECT COUNT(*) FROM payroll_items pi
					JOIN employees e ON e.id = pi.employee_id
					WHERE e.employee_id = $1 AND pi.payroll_period_id = $2::uuid
				`, eid, periodID).Scan(&itemCount)
				if itemCount == 0 {
					// Get base_salary for employee
					var baseSalary float64
					var dailyWage float64
					database.Pool.QueryRow(ctx, `
						SELECT COALESCE(base_salary, 0), COALESCE(daily_wage, 0)
						FROM employee_salary_histories
						WHERE employee_id = (SELECT id FROM employees WHERE employee_id = $1)
						ORDER BY effective_date DESC LIMIT 1
					`, eid).Scan(&baseSalary, &dailyWage)

					grossSalary := baseSalary
					if dailyWage > 0 {
						grossSalary = dailyWage * 22 // 22 working days
					}
					overtimePay := 0.0
					if eid == "EMP-001" || eid == "EMP-004" {
						overtimePay = 150000.0
					}
					thrAmount := 0.0
					if month == 6 { // THR in June
						thrAmount = grossSalary
					}
					netSalary := grossSalary + overtimePay + thrAmount
					if grossSalary > 0 {
						netSalary = grossSalary*0.85 + overtimePay + thrAmount // ~15% deductions
					}

					_, err := database.Pool.Exec(ctx, `
						INSERT INTO payroll_items (payroll_period_id, employee_id, base_salary, daily_wage,
							total_days_worked, overtime_pay, overtime_hours, thr_amount, gross_salary,
						pph21_amount, bpjs_kesehatan, bpjs_jht, total_deductions, net_salary, status)
					SELECT $1::uuid, e.id, $3, $4, 22, $5, 2, $6, $7, $8, $9, $10, $11, $12, 'completed'::payroll_status
						FROM employees e WHERE e.employee_id = $2
					`, periodID, eid, baseSalary, dailyWage, overtimePay, thrAmount, grossSalary,
						grossSalary*0.05, grossSalary*0.01, grossSalary*0.02, grossSalary*0.15, netSalary)
					if err != nil {
						log.Printf("⚠️ Failed to create payroll item for %s: %v", eid, err)
					}
				}
			}
			fmt.Printf("   📊 Payroll items created for period %s\n", periodName)
		}
	}

	// ============================================================
	// 📢 SEED ADDITIONAL ANNOUNCEMENTS
	// ============================================================
	fmt.Println("\n📢 Seeding announcements...")
	announcementSeeds := []struct {
		Title   string
		Content string
		Type    string
		Pinned  bool
	}{
		{"Selamat Hari Raya Idul Fitri 1447 H", "Kami mengucapkan Selamat Hari Raya Idul Fitri 1447 H, mohon maaf lahir dan batin. Libur cuti bersama akan dilaksanakan pada tanggal 1-7 April 2026.", "general", true},
		{"Pemberitahuan Jadwal Penggajian Juni 2026", "Penggajian bulan Juni 2026 akan diproses pada tanggal 30 Juni 2026. Pastikan seluruh data kehadiran dan lembur sudah dilaporkan sebelum tanggal 28 Juni 2026.", "important", false},
		{"Pelatihan Sertifikasi ISO 27001", "Diberitahukan kepada seluruh karyawan bahwa akan diadakan pelatihan ISO 27001 pada tanggal 15-17 Juli 2026. Harap mendaftar melalui HRD maksimal 10 Juli 2026.", "general", false},
		{"Perubahan Kebijakan Work From Home", "Mulai bulan Juli 2026, kebijakan WFH diperpanjang dengan ketentuan baru: maksimal 2 hari per minggu, harus mendapat persetujuan atasan langsung.", "important", false},
	}
	for _, a := range announcementSeeds {
		var count int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM announcements WHERE title = $1`, a.Title).Scan(&count)
		if count == 0 {
			_, err := database.Pool.Exec(ctx, `
				INSERT INTO announcements (title, content, announcement_type, is_pinned, published_at, created_by)
				VALUES ($1, $2, $3::announcement_type, $4, $5, $6::uuid)
			`, a.Title, a.Content, a.Type, a.Pinned, now, adminUUID)
			if err != nil {
				log.Printf("⚠️ Failed to create announcement %s: %v", a.Title, err)
			} else {
				fmt.Printf("📢 Announcement created: %s\n", a.Title)
			}
		}
	}

	// ============================================================
	// ⚠️ SEED REPRIMANDS
	// ============================================================
	fmt.Println("\n⚠️ Seeding reprimands...")
	type reprimandSeed struct {
		EmployeeID  string
		Type        string
		Title       string
		Description string
		Status      string
	}

	reprimandSeeds := []reprimandSeed{
		{"EMP-003", "sp1", "Surat Peringatan 1 — Keterlambatan Berulang", "Terlambat masuk kerja sebanyak 5 kali dalam sebulan tanpa keterangan yang sah.", "issued"},
		{"EMP-005", "verbal", "Teguran Lisan — Pelanggaran Aturan Seragam", "Tidak menggunakan seragam kantor sesuai ketentuan yang berlaku.", "acknowledged"},
		{"EMP-007", "sp1", "Surat Peringatan 1 — Penyalahgunaan Fasilitas Kantor", "Menggunakan kendaraan operasional untuk keperluan pribadi tanpa izin.", "issued"},
	}
	for _, r := range reprimandSeeds {
		var count int
		database.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM reprimands rp
			JOIN employees e ON e.id = rp.employee_id
			WHERE e.employee_id = $1 AND rp.title = $2
		`, r.EmployeeID, r.Title).Scan(&count)
		if count == 0 {
			ackDate := "NULL"
			ackNote := "NULL"
			if r.Status == "acknowledged" {
				ackDate = fmt.Sprintf("'%s'", now)
				ackNote = "'Saya mengakui kesalahan dan berjanji tidak akan mengulangi lagi'"
			}
			expDate := time.Now().AddDate(0, 3, 0).Format("2006-01-02") // 3 months from now
			query := fmt.Sprintf(`
				INSERT INTO reprimands (employee_id, reprimand_type, title, description, violation_date,
					issued_by, issued_date, effective_period_months, status, expired_at, acknowledgment_date, acknowledgment_note)
				SELECT e.id, $2::reprimand_type, $3, $4, CURRENT_DATE - interval '7 days',
					$5::uuid, CURRENT_DATE, 3, $6::reprimand_status,
					'%s'::date, %s::timestamptz, %s
				FROM employees e WHERE e.employee_id = $1
			`, expDate, ackDate, ackNote)
			_, err := database.Pool.Exec(ctx, query,
				r.EmployeeID, r.Type, r.Title, r.Description, adminUUID, r.Status)
			if err != nil {
				log.Printf("⚠️ Failed to seed reprimand %s: %v", r.Title, err)
			} else {
				fmt.Printf("⚠️ Reprimand created: %s for %s\n", r.Title, r.EmployeeID)
			}
		}
	}

	// ============================================================
	// 🔄 SEED SHIFT CHANGE REQUESTS
	// ============================================================
	fmt.Println("\n🔄 Seeding shift change requests...")
	type shiftSeed struct {
		EmployeeID      string
		RequestType     string
		TargetDate      string
		Reason          string
		SwapPartnerID   string
		Status          string
	}
	// Get a work schedule ID for shift change requests (FK references work_schedules.id)
	var shiftWorkScheduleID string
	database.Pool.QueryRow(ctx, `SELECT id::text FROM work_schedules WHERE deleted_at IS NULL LIMIT 1`).Scan(&shiftWorkScheduleID)

	if shiftWorkScheduleID != "" {
		shiftSeeds := []shiftSeed{
			{"EMP-005", "individual", "2026-07-10", "Ada keperluan keluarga mendadak", "", "pending"},
			{"EMP-007", "individual", "2026-07-15", "Janjian dengan dokter gigi", "", "pending"},
			{"EMP-006", "swap", "2026-07-20", "Ingin tukar shift dengan Rina", "EMP-007", "pending"},
		}
		for _, s := range shiftSeeds {
			var count int
			database.Pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM shift_change_requests scr
				JOIN employees e ON e.id = scr.employee_id
				WHERE e.employee_id = $1 AND scr.target_date = $2::date
			`, s.EmployeeID, s.TargetDate).Scan(&count)
			if count == 0 {
				var swapPartnerUUID string
				if s.SwapPartnerID != "" {
					database.Pool.QueryRow(ctx, `SELECT id::text FROM employees WHERE employee_id = $1`, s.SwapPartnerID).Scan(&swapPartnerUUID)
				}
				_, err := database.Pool.Exec(ctx, `
					INSERT INTO shift_change_requests (request_type, employee_id, target_date,
						requested_schedule_id, reason, swap_partner_id, status)
					SELECT $2::shift_change_type, e.id, $3::date, $4::uuid, $5,
						CASE WHEN $6 <> '' THEN $6::uuid ELSE NULL END,
						$7::shift_change_status
					FROM employees e WHERE e.employee_id = $1
				`, s.EmployeeID, s.RequestType, s.TargetDate, shiftWorkScheduleID, s.Reason, swapPartnerUUID, s.Status)
				if err != nil {
					log.Printf("⚠️ Failed to seed shift change for %s: %v", s.EmployeeID, err)
				} else {
					fmt.Printf("🔄 Shift change created: %s on %s [%s]\n", s.EmployeeID, s.TargetDate, s.RequestType)
				}
			}
		}
	}

	// ============================================================
	// 📋 SEED LEAVE BALANCES (if not auto-generated)
	// ============================================================
	fmt.Println("\n📋 Seeding leave balances...")
	var lbCount int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM leave_balances WHERE year = 2026`).Scan(&lbCount)
	if lbCount == 0 {
		empList := []string{"EMP-001", "EMP-002", "EMP-003", "EMP-004", "EMP-005", "EMP-006", "EMP-007", "EMP-008"}
		for _, eid := range empList {
			// Get leave types and create balances
			rows, err := database.Pool.Query(ctx, `SELECT id::text, code FROM leave_types WHERE is_active = TRUE`)
			if err != nil {
				log.Printf("⚠️ Failed to query leave types: %v", err)
				continue
			}
			for rows.Next() {
				var ltID, ltCode string
				rows.Scan(&ltID, &ltCode)
				quota := 12
				if ltCode == "sakit" {
					quota = 365
				} else if ltCode == "menikah" || ltCode == "melahirkan" {
					quota = 3
				}
				used := 0
				if eid == "EMP-001" && ltCode == "tahunan" {
					used = 3
				} else if eid == "EMP-002" && ltCode == "sakit" {
					used = 2
				} else if eid == "EMP-005" && ltCode == "sakit" {
					used = 1
				}
				remaining := quota - used
				_, err := database.Pool.Exec(ctx, `
					INSERT INTO leave_balances (employee_id, leave_type_id, year, total_quota, used, remaining)
					SELECT e.id, $2::uuid, 2026, $3, $4, $5
					FROM employees e WHERE e.employee_id = $1
				`, eid, ltID, quota, used, remaining)
				if err != nil {
					log.Printf("⚠️ Failed to seed leave balance for %s: %v", eid, err)
				}
			}
			rows.Close()
			fmt.Printf("📋 Leave balances seeded for %s\n", eid)
		}
	}

	// ============================================================
	// 🔔 SEED NOTIFICATIONS
	// ============================================================
	fmt.Println("\n🔔 Seeding notifications...")
	notifSeeds := []struct {
		EmployeeID string
		Type       string
		Title      string
		Body       string
	}{
		{"EMP-001", "approved", "Cuti Disetujui", "Pengajuan cuti tahunan Anda tanggal 7-9 Juli 2026 telah disetujui."},
		{"EMP-002", "approved", "Cuti Disetujui", "Pengajuan cuti sakit Anda tanggal 23-24 Juni 2026 telah disetujui."},
		{"EMP-001", "approved", "Lembur Disetujui", "Pengajuan lembur Anda tanggal 20 Juni 2026 telah disetujui."},
		{"EMP-004", "approved", "Lembur Disetujui", "Pengajuan lembur Anda tanggal 23 Juni 2026 telah disetujui."},
		{"EMP-002", "approved", "Pinjaman Disetujui", "Pengajuan pinjaman darurat Anda sebesar Rp3.000.000 telah disetujui."},
	}
	for _, n := range notifSeeds {
		var count int
		database.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM notifications n2
			WHERE n2.title = $1 AND n2.user_id = (SELECT id FROM employees WHERE employee_id = $2)
		`, n.Title, n.EmployeeID).Scan(&count)
		if count == 0 {
			_, err := database.Pool.Exec(ctx, `
				INSERT INTO notifications (user_id, notification_type, title, body, is_read)
				SELECT e.id, $2::notification_type, $3, $4, FALSE
				FROM employees e WHERE e.employee_id = $1
			`, n.EmployeeID, n.Type, n.Title, n.Body)
			if err != nil {
				log.Printf("⚠️ Failed to create notification for %s: %v", n.EmployeeID, err)
			} else {
				fmt.Printf("🔔 Notification: %s — %s\n", n.EmployeeID, n.Title)
			}
		}
	}

	// ============================================================
	// 📊 SEED ADDITIONAL POSITION GRADES (sudah ada seed dari migration 00003)
	// ============================================================
	fmt.Println("\n📊 Checking position grades...")
	var gradeCount int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM position_grades`).Scan(&gradeCount)
	if gradeCount == 0 {
		// Migration 00003 seharusnya sudah seed, ini fallback
		type gradeSeed struct {
			Name        string
			Level       int
			Description string
		}
		gradeSeeds := []gradeSeed{
			{"Staff", 1, "Karyawan pelaksana / staf"},
			{"Senior Staff", 2, "Karyawan senior dengan pengalaman"},
			{"Supervisor", 3, "Pengawas tim kecil"},
			{"Senior Supervisor", 4, "Pengawas tim menengah"},
			{"Assistant Manager", 5, "Asisten manajer departemen"},
			{"Manager", 6, "Kepala departemen"},
			{"Senior Manager", 7, "Kepala departemen senior"},
			{"General Manager", 8, "Kepala divisi"},
			{"Director", 9, "Direktur"},
			{"President Director", 10, "Direktur Utama"},
		}
		for _, g := range gradeSeeds {
			_, err := database.Pool.Exec(ctx, `
				INSERT INTO position_grades (name, level, description)
				VALUES ($1, $2, $3)
			`, g.Name, g.Level, g.Description)
			if err != nil {
				log.Printf("⚠️ Failed to seed position grade %s: %v", g.Name, err)
			}
		}
		fmt.Printf("📊 Seeded %d position grades from fallback\n", len(gradeSeeds))
	} else {
		fmt.Printf("📊 Position grades already exist (%d grades), skipping\n", gradeCount)
	}

	// ============================================================
	// ✅ SEED ATTENDANCE RECORDS (check-in/out history — untuk dashboard)
	// ============================================================
	fmt.Println("\n✅ Seeding attendance records...")
	today := time.Now()
	empIDs := []string{"EMP-001", "EMP-002", "EMP-003", "EMP-004", "EMP-005", "EMP-006", "EMP-007", "EMP-008"}
	totalAttend := 0
	for i := 1; i <= 14; i++ { // 14 hari ke belakang
		date := today.AddDate(0, 0, -i)
		weekday := date.Weekday()
		if weekday == time.Sunday || weekday == time.Saturday {
			continue // Skip weekends
		}
		dateStr := date.Format("2006-01-02")
		for _, eid := range empIDs {
			var count int
			database.Pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM attendance_records ar
				JOIN employees e ON e.id = ar.employee_id
				WHERE e.employee_id = $1 AND ar.date = $2::date
			`, eid, dateStr).Scan(&count)
			if count == 0 {
				hour := 7
				min := 30 + (i*3)%30
				if min >= 60 {
					hour = 8
					min = min - 60
				}
				checkIn := fmt.Sprintf("%s %02d:%02d:00", dateStr, hour, min)
				checkOut := fmt.Sprintf("%s 17:%02d:00", dateStr, (i*2)%30)
				status := "hadir"
				if min > 15 {
					status = "terlambat"
				}
				// Some employees 'sakit' or 'izin' on random days
				if i == 5 && (eid == "EMP-005" || eid == "EMP-007") {
					status = "sakit"
					checkIn = fmt.Sprintf("%s 00:00:00", dateStr)
					checkOut = fmt.Sprintf("%s 00:00:00", dateStr)
				}
				if i == 8 && eid == "EMP-003" {
					status = "izin"
					checkIn = fmt.Sprintf("%s 00:00:00", dateStr)
					checkOut = fmt.Sprintf("%s 00:00:00", dateStr)
				}
				_, err := database.Pool.Exec(ctx, `
					INSERT INTO attendance_records (employee_id, date, check_in_time, check_out_time, status)
					SELECT e.id, $1::date, $2::timestamptz, $3::timestamptz, $4::attendance_status
					FROM employees e WHERE e.employee_id = $5
				`, dateStr, checkIn, checkOut, status, eid)
				if err != nil {
					log.Printf("⚠️ Failed to seed attendance for %s on %s: %v", eid, dateStr, err)
				} else {
					totalAttend++
				}
			}
		}
	}
	fmt.Printf("✅ Attendance records seeded: %d total records (past 14 working days)\n", totalAttend)

	// ============================================================
	// 📝 SEED DAILY JOURNALS (jurnal harian)
	// ============================================================
	fmt.Println("\n📝 Seeding daily journals...")
	type journalSeed struct {
		EmployeeID      string
		JournalDate     string
		WorkDescription string
		Achievements    string
		Challenges      string
		PlanTomorrow    string
		Status          string // draft | submitted | acknowledged
	}

	// Get dept IDs for department_id
	var itDeptID, hrDeptID, finDeptID string
	database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE code = 'IT' LIMIT 1`).Scan(&itDeptID)
	database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE code = 'HR' LIMIT 1`).Scan(&hrDeptID)
	database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE code = 'FIN' LIMIT 1`).Scan(&finDeptID)

	journalSeeds := []journalSeed{
		{"EMP-001", time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			"Menyelesaikan migrasi server database ke versi terbaru. Melakukan testing dan validasi data.",
			"Berhasil migrasi 5 database tanpa downtime. Testing all clear.",
			"Kendala koneksi jaringan saat migrasi, namun bisa diatasi.",
			"Lanjutkan optimasi query database yang lambat.",
			"submitted"},
		{"EMP-001", time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
			"Melakukan maintenance rutin server dan backup database harian.",
			"Backup semua database selesai tepat waktu. Server monitoring all green.",
			"",
			"Persiapan environment untuk development fitur baru.",
			"acknowledged"},
		{"EMP-001", time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
			"Meeting dengan tim developer untuk membahas arsitektur fitur baru.",
			"",
			"",
			"Buat draft technical specification.",
			"acknowledged"},
		{"EMP-002", time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			"Menyusun laporan keuangan bulan Juni 2026. Rekonsiliasi bank dan penutupan buku.",
			"Laporan keuangan selesai tepat waktu. Tidak ada selisih.",
			"Beberapa transaksi bulan lalu belum tercatat.",
			"Finalisasi laporan keuangan dan siap untuk audit.",
			"submitted"},
		{"EMP-002", time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
			"Memproses pembayaran reimbursement dan gaji karyawan.",
			"Semua pembayaran berhasil diproses.",
			"",
			"Lanjutkan persiapan data untuk pelaporan pajak.",
			"acknowledged"},
		{"EMP-004", time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			"Melakukan briefing HR dengan seluruh staf. Review kebijakan baru perusahaan.",
			"Briefing berjalan lancar. Semua staf memahami kebijakan baru.",
			"",
			"Persiapan dokumen untuk karyawan baru yang akan masuk.",
			"submitted"},
		{"EMP-005", time.Now().Format("2006-01-02"),
			"Menyelesaikan bug fixing pada modul attendance. Testing dan debugging.",
			"Berhasil fix 3 bug pada modul attendance.",
			"Bug ke-4 masih dalam investigasi.",
			"Lanjutkan debugging bug terakhir dan code review.",
			"draft"},
		{"EMP-007", time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			"Membuat konten marketing untuk campaign produk baru Q3 2026.",
			"Konten brochure dan landing page selesai.",
			"Menunggu approval dari tim produk untuk finalisasi.",
			"Follow up approval dan siapkan materi social media.",
			"submitted"},
	}

	for _, js := range journalSeeds {
		var count int
		database.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM daily_journals dj
			JOIN employees e ON e.id = dj.employee_id
			WHERE e.employee_id = $1 AND dj.journal_date = $2::date
		`, js.EmployeeID, js.JournalDate).Scan(&count)

		if count == 0 {
			// Determine department_id based on employee
			var deptID string
			database.Pool.QueryRow(ctx, `
				SELECT COALESCE(d.id::text, '') FROM employees e
				LEFT JOIN departments d ON d.id = e.department_id
				WHERE e.employee_id = $1
			`, js.EmployeeID).Scan(&deptID)

			submittedAt := "NULL"
			ackBy := "NULL"
			ackAt := "NULL"
			status := js.Status

			if status == "submitted" {
				submittedAt = fmt.Sprintf("'%s'", now)
			} else if status == "acknowledged" {
				submittedAt = fmt.Sprintf("'%s'", now)
				ackBy = fmt.Sprintf("'%s'::uuid", adminUUID)
				ackAt = fmt.Sprintf("'%s'", now)
			}

			query := fmt.Sprintf(`
				INSERT INTO daily_journals (employee_id, journal_date, work_description,
					achievements, challenges, plan_tomorrow, status, department_id,
					submitted_at, acknowledged_by, acknowledged_at)
				SELECT e.id, $2::date, $3, $4, $5, $6, $7::journal_status,
					NULLIF($8, '')::uuid,
					%s, %s, %s
				FROM employees e WHERE e.employee_id = $1
			`, submittedAt, ackBy, ackAt)

			_, err := database.Pool.Exec(ctx, query,
				js.EmployeeID, js.JournalDate, js.WorkDescription,
				js.Achievements, js.Challenges, js.PlanTomorrow,
				status, deptID)
			if err != nil {
				log.Printf("⚠️ Failed to seed daily journal for %s: %v", js.EmployeeID, err)
			} else {
				fmt.Printf("📝 Daily journal: %s — %s [%s]\n", js.EmployeeID, js.JournalDate, status)
			}
		} else {
			fmt.Printf("📝 Daily journal already exists for %s on %s, skipping\n", js.EmployeeID, js.JournalDate)
		}
	}

	// ============================================================
	// 💳 SEED LOAN INSTALLMENTS (for active loans)
	// ============================================================
	fmt.Println("\n💳 Seeding loan installments...")
	// For EMP-001 active loan
	var loanID string
	database.Pool.QueryRow(ctx, `
		SELECT l.id::text FROM loans l
		JOIN employees e ON e.id = l.employee_id
		WHERE e.employee_id = 'EMP-001' AND l.status = 'active' AND l.deleted_at IS NULL
		LIMIT 1
	`).Scan(&loanID)
	if loanID != "" {
		var instCount int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM loan_installments WHERE loan_id = $1::uuid`, loanID).Scan(&instCount)
		if instCount == 0 {
			var installmentAmount float64
			var totalInstallments int
			database.Pool.QueryRow(ctx, `
				SELECT installment_amount, installment_count FROM loans WHERE id = $1::uuid
			`, loanID).Scan(&installmentAmount, &totalInstallments)
			startDate := time.Now().AddDate(0, -2, 0) // Started 2 months ago
			for i := 0; i < totalInstallments; i++ {
				dueDate := startDate.AddDate(0, i, 0)
				paidDate := "NULL"
				paidAmt := "NULL"
				status := "pending"
				if i < 2 { // First 2 installments are paid
					paidDate = fmt.Sprintf("'%s'::date", dueDate.Format("2006-01-02"))
					paidAmt = fmt.Sprintf("%.2f", installmentAmount)
					status = "paid"
				}
				query := fmt.Sprintf(`
					INSERT INTO loan_installments (loan_id, installment_number, amount, due_date,
						paid_date, paid_amount, status)
					VALUES ($1::uuid, $2, $3, $4::date, %s, %s, $5)
				`, paidDate, paidAmt)
				_, err := database.Pool.Exec(ctx, query,
					loanID, i+1, installmentAmount, dueDate.Format("2006-01-02"), status)
				if err != nil {
					log.Printf("⚠️ Failed to seed installment %d: %v", i+1, err)
				}
			}
			fmt.Printf("💳 Loan installments seeded: %d installments for EMP-001\n", totalInstallments)
		}
	}

	// ============================================================
	// 📊 FINAL TOTALS
	// ============================================================
	var total, overtimeTotal, reimbursementTotal, loanTotal, leaveTotal, kpiReviewTotal,
		attLocationTotal, docTotal, scheduleTotal, payrollPeriodTotal, reprimandTotal,
		shiftTotal, notifTotal, emergencyTotal, instTotal, annTotal, dailyJournalTotal int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL`).Scan(&total)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM overtime_requests WHERE deleted_at IS NULL`).Scan(&overtimeTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM reimbursements WHERE deleted_at IS NULL`).Scan(&reimbursementTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM loans WHERE deleted_at IS NULL`).Scan(&loanTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM leave_requests WHERE deleted_at IS NULL`).Scan(&leaveTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM kpi_reviews WHERE deleted_at IS NULL`).Scan(&kpiReviewTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM attendance_locations`).Scan(&attLocationTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_documents`).Scan(&docTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_schedules`).Scan(&scheduleTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM payroll_periods`).Scan(&payrollPeriodTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM reprimands`).Scan(&reprimandTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM shift_change_requests`).Scan(&shiftTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&notifTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_emergency_contacts`).Scan(&emergencyTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM loan_installments`).Scan(&instTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM announcements`).Scan(&annTotal)
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM daily_journals WHERE deleted_at IS NULL`).Scan(&dailyJournalTotal)

	fmt.Printf("\n📊 DATA TOTALS\n")
	fmt.Printf("   👥 Employees: %d\n", total)
	fmt.Printf("   📍 Attendance Locations: %d\n", attLocationTotal)
	fmt.Printf("   📄 Employee Documents: %d\n", docTotal)
	fmt.Printf("   📅 Employee Schedules: %d\n", scheduleTotal)
	fmt.Printf("   🔄 Shift Change Requests: %d\n", shiftTotal)
	fmt.Printf("   ⏰ Overtime Requests: %d\n", overtimeTotal)
	fmt.Printf("   💰 Reimbursements: %d\n", reimbursementTotal)
	fmt.Printf("   💳 Loans: %d | Installments: %d\n", loanTotal, instTotal)
	fmt.Printf("   📅 Leave Requests: %d\n", leaveTotal)
	fmt.Printf("   📊 KPI Reviews: %d\n", kpiReviewTotal)
	fmt.Printf("   💰 Payroll Periods: %d\n", payrollPeriodTotal)
	fmt.Printf("   ⚠️ Reprimands: %d\n", reprimandTotal)
	fmt.Printf("   🆘 Emergency Contacts: %d\n", emergencyTotal)
	fmt.Printf("   🔔 Notifications: %d\n", notifTotal)
	fmt.Printf("   📢 Announcements: %d\n", annTotal)
	fmt.Printf("   📝 Daily Journals: %d\n", dailyJournalTotal)
	fmt.Println("🌱 Seed completed!")
}
