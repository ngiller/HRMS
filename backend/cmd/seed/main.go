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

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL()); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	fmt.Println("🌱 Seeding database...")

	// Disable audit triggers temporarily
	database.Pool.Exec(ctx, `ALTER TABLE employees DISABLE TRIGGER audit_employees`)

	// Hash default passwords
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	userPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Check if admin already exists
	var adminCount int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE employee_id = 'ADMIN-001'`).Scan(&adminCount)

	if adminCount == 0 {
		var superAdminRoleID string
		err := database.Pool.QueryRow(ctx, `SELECT id::text FROM roles WHERE slug = 'super_admin'`).Scan(&superAdminRoleID)
		if err != nil {
			log.Fatalf("No super_admin role found: %v", err)
		}

		_, err = database.Pool.Exec(ctx,
			`INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active, role_id)
			 VALUES ('ADMIN-001', 'Super Admin', 'admin@company.com', $1, 'laki_laki', CURRENT_DATE, 'tetap', TRUE, $2::uuid)`,
			string(adminPassword), superAdminRoleID)
		if err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}
		fmt.Println("✅ Admin user created: admin@company.com / admin123")
	} else {
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

	// Get employee role UUID
	var employeeRoleID string
	err := database.Pool.QueryRow(ctx, `SELECT id::text FROM roles WHERE slug = 'employee'`).Scan(&employeeRoleID)
	if err != nil {
		log.Printf("⚠️ No employee role found: %v", err)
		employeeRoleID = ""
	}

	// Create departments first using direct SQL (no parameterized types)
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
				`INSERT INTO departments (name, code) VALUES ($1, $2) RETURNING id::text`,
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

	// Create test employees
	testEmployees := []struct {
		ID         string
		Name       string
		Email      string
		Status     string
		Gender     string
		Position   string
		DeptName   string
	}{
		{"EMP-001", "Budi Hartono", "budi@company.com", "tetap", "laki_laki", "Staff IT", "Teknologi Informasi"},
		{"EMP-002", "Siti Rahayu", "siti@company.com", "tetap", "perempuan", "Finance Staff", "Keuangan"},
		{"EMP-003", "Andi Wijaya", "andi@company.com", "kontrak", "laki_laki", "Sales Executive", "Penjualan"},
		{"EMP-004", "Dewi Lestari", "dewi@company.com", "kontrak", "perempuan", "HR Staff", "Sumber Daya Manusia"},
		{"EMP-005", "Rudi Hartono", "rudi@company.com", "percobaan", "laki_laki", "Staff IT", "Teknologi Informasi"},
		{"EMP-006", "Ahmad Fauzi", "ahmad@company.com", "percobaan", "laki_laki", "Sales", "Penjualan"},
		{"EMP-007", "Rina Marlina", "rina@company.com", "tetap", "perempuan", "Marketing", "Pemasaran"},
		{"EMP-008", "Hendra Gunawan", "hendra@company.com", "tetap", "laki_laki", "Accounting", "Keuangan"},
	}

	createdCount := 0
	for _, emp := range testEmployees {
		var count int
		database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE employee_id = $1`, emp.ID).Scan(&count)

		if count > 0 {
			continue
		}

		posID := posIDs[emp.Position]
		deptID := deptIDs[emp.DeptName]
		joinDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		if emp.Status == "percobaan" {
			joinDate = time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
		}

		query := `INSERT INTO employees (employee_id, full_name, email, password_hash, gender,
			join_date, employment_status, is_active, role_id, position_id, department_id)
			VALUES ($1, $2, $3, $4, $5::gender_type, $6, $7::employment_status, TRUE,
			NULLIF($8, '')::uuid, NULLIF($9, '')::uuid, NULLIF($10, '')::uuid)`

		_, err := database.Pool.Exec(ctx, query,
			emp.ID, emp.Name, emp.Email, string(userPassword),
			emp.Gender, joinDate, emp.Status,
			employeeRoleID, posID, deptID)
		if err != nil {
			log.Printf("⚠️ Could not create employee %s: %v", emp.Name, err)
		} else {
			createdCount++
			fmt.Printf("✅ Employee created: %s (%s) / password123\n", emp.Name, emp.Email)
		}
	}

	// Re-enable audit trigger
	database.Pool.Exec(ctx, `ALTER TABLE employees ENABLE TRIGGER audit_employees`)

	// Check total employees
	var total int
	database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE deleted_at IS NULL`).Scan(&total)

	fmt.Printf("\n📊 Total employees in database: %d\n", total)
	fmt.Println("🌱 Seed completed!")
}
