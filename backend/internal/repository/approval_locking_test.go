// ============================================================
// Integration Test — Optimistic Locking di Approval Flows
//
// Build tag: integration (go test -tags=integration)
// Skip:      go test -short (go test -short ./...)
// Env:       RUN_INTEGRATION_TESTS=true
//
// Requires a running PostgreSQL instance with migrations applied.
// Use docker-compose up -d db migrate untuk setup test database.
// ============================================================
//
//go:build integration

package repository

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"hrms-backend/internal/config"
	"hrms-backend/internal/database"
)

// TestMain initializes database connection for integration tests.
// Only runs when RUN_INTEGRATION_TESTS=true.
func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		os.Exit(0) // Skip entirely if not integration mode
	}
	cfg := config.Load()
	if err := database.Connect(cfg.DatabaseURL(), cfg.EncryptionKey); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	os.Exit(m.Run())
}

// skipIfNotIntegration skips the test if not in integration test mode.
func skipIfNotIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test — set RUN_INTEGRATION_TESTS=true to run")
	}
}

// setupTestData creates test records needed for optimistic locking tests.
// Returns cleanup function that deletes all test data.
func setupTestApprovals(t *testing.T, ctx context.Context) func() {
	t.Helper()

	// Pastikan database terhubung
	if database.Pool == nil {
		t.Fatal("database.Pool is nil — pastikan database terhubung")
	}

	// Cek koneksi
	if err := database.Pool.Ping(ctx); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}

	// Pastikan employee test ada (dari seed data)
	// Seed data seharusnya sudah ada dari migration + cmd/seed
	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Fatalf("seed employee EMP001 not found — jalankan seed dulu: %v", err)
	}

	// Pastikan leave_types ada
	var leaveTypeID string
	err = database.Pool.QueryRow(ctx,
		`SELECT id::text FROM leave_types WHERE code = 'tahunan' LIMIT 1`,
	).Scan(&leaveTypeID)
	if err != nil {
		t.Fatalf("leave_type 'tahunan' not found — jalankan migration dulu: %v", err)
	}

	// Pastikan leave_balance ada, buat jika belum
	_, err = database.Pool.Exec(ctx, `
		INSERT INTO leave_balances (employee_id, leave_type_id, year, total_quota, used, remaining)
		VALUES ($1::uuid, $2::uuid, EXTRACT(YEAR FROM NOW())::int, 12, 0, 12)
		ON CONFLICT (employee_id, leave_type_id, year) DO NOTHING
	`, empID, leaveTypeID)
	if err != nil {
		t.Fatalf("create leave_balance: %v", err)
	}

	return func() {
		// Cleanup: hapus leave requests test
		database.Pool.Exec(ctx,
			`DELETE FROM leave_requests WHERE reason LIKE '[TEST]%'`,
		)
	}
}

// TestOptimisticLocking_Leave_UpdateStatus_DoubleApprove menguji bahwa
// jika 2 goroutine mencoba approve pengajuan cuti yang sama secara bersamaan,
// hanya 1 yang berhasil. Yang kedua harus gagal dengan error "sudah diproses".
func TestOptimisticLocking_Leave_UpdateStatus_DoubleApprove(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()
	cleanup := setupTestApprovals(t, ctx)
	defer cleanup()

	// Cari employee & leave type dari seed
	var empID, leaveTypeID string
	err := database.Pool.QueryRow(ctx,
		`SELECT e.id::text, lt.id::text FROM employees e
		 CROSS JOIN leave_types lt
		 WHERE e.employee_id = 'EMP-001' AND lt.code = 'tahunan'
		 LIMIT 1`,
	).Scan(&empID, &leaveTypeID)
	if err != nil {
		t.Fatalf("get test data: %v", err)
	}

	// Buat 1 pengajuan cuti test
	var leaveID string
	err = database.Pool.QueryRow(ctx, `
		INSERT INTO leave_requests (employee_id, leave_type_id, start_date, end_date,
			total_days, is_half_day, reason, status)
		VALUES ($1::uuid, $2::uuid, CURRENT_DATE, (CURRENT_DATE + INTERVAL '1 day')::date,
			1, false, '[TEST] Optimistic Locking - concurrent approve', 'pending'::leave_status)
		RETURNING id::text
	`, empID, leaveTypeID).Scan(&leaveID)
	if err != nil {
		t.Fatalf("create test leave request: %v", err)
	}
	t.Logf("Created test leave request: %s", leaveID)

	// Simulasikan 2 approve request secara concurrent (race condition)
	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(approverSuffix string) {
			defer wg.Done()
			_, err := UpdateLeaveStatus(ctx, leaveID, "approved", "", "00000000-0000-0000-0000-00000000000"+approverSuffix)
			mu.Lock()
			if err != nil {
				t.Logf("Approve attempt failed (expected): %v", err)
				errorCount++
			} else {
				successCount++
			}
			mu.Unlock()
		}(string(rune('1' + i))) // "1" atau "2"
	}
	wg.Wait()

	// Assert: hanya 1 yang berhasil approve
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful approve, got %d", successCount)
	}
	if errorCount != 1 {
		t.Errorf("Expected exactly 1 failed approve, got %d", errorCount)
	}

	// Verifikasi: status leave request jadi 'approved'
	var status string
	err = database.Pool.QueryRow(ctx,
		`SELECT status::text FROM leave_requests WHERE id::text = $1`, leaveID,
	).Scan(&status)
	if err != nil {
		t.Fatalf("get leave status: %v", err)
	}
	if status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", status)
	}
}

// TestOptimisticLocking_Leave_UpdateStatus_AlreadyProcessed menguji bahwa
// approve setelah reject (atau sebaliknya) gagal karena status sudah berubah.
func TestOptimisticLocking_Leave_UpdateStatus_AlreadyProcessed(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()
	cleanup := setupTestApprovals(t, ctx)
	defer cleanup()

	var empID, leaveTypeID string
	err := database.Pool.QueryRow(ctx,
		`SELECT e.id::text, lt.id::text FROM employees e
		 CROSS JOIN leave_types lt
		 WHERE e.employee_id = 'EMP-001' AND lt.code = 'tahunan'
		 LIMIT 1`,
	).Scan(&empID, &leaveTypeID)
	if err != nil {
		t.Fatalf("get test data: %v", err)
	}

	// Buat 1 pengajuan cuti test
	var leaveID string
	err = database.Pool.QueryRow(ctx, `
		INSERT INTO leave_requests (employee_id, leave_type_id, start_date, end_date,
			total_days, is_half_day, reason, status)
		VALUES ($1::uuid, $2::uuid, CURRENT_DATE, (CURRENT_DATE + INTERVAL '1 day')::date,
			1, false, '[TEST] Already processed', 'pending'::leave_status)
		RETURNING id::text
	`, empID, leaveTypeID).Scan(&leaveID)
	if err != nil {
		t.Fatalf("create test leave request: %v", err)
	}
	t.Logf("Created test leave request: %s", leaveID)

	// Step 1: Approve pertama — harus sukses
	_, err = UpdateLeaveStatus(ctx, leaveID, "approved", "", "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("First approve should succeed: %v", err)
	}

	// Step 2: Approve kedua — harus gagal karena status sudah 'approved'
	_, err = UpdateLeaveStatus(ctx, leaveID, "approved", "", "00000000-0000-0000-0000-000000000002")
	if err == nil {
		t.Error("Second approve should fail with 'sudah diproses' error")
	} else {
		t.Logf("Second approve correctly rejected: %v", err)
	}

	// Step 3: Reject setelah approve — harus gagal juga
	_, err = UpdateLeaveStatus(ctx, leaveID, "rejected", "Alasan reject", "00000000-0000-0000-0000-000000000003")
	if err == nil {
		t.Error("Reject after approve should fail")
	} else {
		t.Logf("Reject after approve correctly rejected: %v", err)
	}
}

// TestOptimisticLocking_ShiftChange_UpdateStatus menguji optimistic locking
// di shift change request — approve/reject hanya bisa pada status pending/partner_pending.
func TestOptimisticLocking_ShiftChange_UpdateStatus(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cek apakah ada employee dan work_schedule
	var empID, wsID string
	err := database.Pool.QueryRow(ctx,
		`SELECT e.id::text, ws.id::text FROM employees e
		 CROSS JOIN work_schedules ws
		 WHERE e.employee_id = 'EMP-001' AND ws.name = '5 Hari Kerja'
		 LIMIT 1`,
	).Scan(&empID, &wsID)
	if err != nil {
		t.Skipf("Skip — test data tidak lengkap: %v", err)
	}

	// Buat 1 shift change request (tipe individual)
	req := &struct {
		RequestType         string
		TargetDate          string
		CurrentScheduleID   string
		RequestedScheduleID string
		Reason              string
	}{
		RequestType:         "individual",
		TargetDate:          "2026-07-15",
		CurrentScheduleID:   "",
		RequestedScheduleID: wsID,
		Reason:              "[TEST] Shift change locking",
	}

	var scID string
	err = database.Pool.QueryRow(ctx, `
		INSERT INTO shift_change_requests (request_type, employee_id, target_date,
			current_schedule_id, requested_schedule_id, reason, status)
		VALUES ($1::shift_change_type, $2::uuid, $3::date,
			NULL, $4::uuid, $5, 'pending'::shift_change_status)
		RETURNING id::text
	`, req.RequestType, empID, req.TargetDate, req.RequestedScheduleID, req.Reason).Scan(&scID)
	if err != nil {
		t.Fatalf("create shift change request: %v", err)
	}
	t.Logf("Created shift change request: %s", scID)

	// Double-approve test (sama seperti leave)
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(suffix string) {
			defer wg.Done()
			_, err := UpdateShiftChangeStatus(ctx, scID, "approved", "", "00000000-0000-0000-0000-00000000000"+suffix)
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				t.Logf("Shift change approve attempt failed: %v", err)
			}
			mu.Unlock()
		}(string(rune('1' + i)))
	}
	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful shift change approve, got %d", successCount)
	}

	// Cleanup
	database.Pool.Exec(ctx,
		`DELETE FROM shift_change_requests WHERE reason LIKE '[TEST]%'`,
	)
}

// TestOptimisticLocking_Reimbursement menguji flow approve → pay optimistic locking.
func TestOptimisticLocking_Reimbursement(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee not found: %v", err)
	}

	// Buat reimbursement request
	var reimbID string
	err = database.Pool.QueryRow(ctx, `
		INSERT INTO reimbursements (employee_id, type, amount, description, status)
		VALUES ($1::uuid, 'medical'::reimbursement_type, 500000, '[TEST] Reimbursement locking', 'pending'::reimbursement_status)
		RETURNING id::text
	`, empID).Scan(&reimbID)
	if err != nil {
		t.Fatalf("create reimbursement: %v", err)
	}
	t.Logf("Created reimbursement: %s", reimbID)

	// Approve — harus sukses
	_, err = UpdateReimbursementStatus(ctx, reimbID, "approved", "", "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("Approve reimbursement should succeed: %v", err)
	}

	// Double approve — harus gagal
	_, err = UpdateReimbursementStatus(ctx, reimbID, "approved", "", "00000000-0000-0000-0000-000000000002")
	if err == nil {
		t.Error("Second approve should fail")
	} else {
		t.Logf("Second approve correctly rejected: %v", err)
	}

	// Cancel setelah approve — harus sukses (cancel diizinkan dari status approved)
	err = CancelReimbursement(ctx, reimbID, empID)
	if err != nil {
		t.Errorf("Cancel after approve should succeed: %v", err)
	}

	// Cleanup
	database.Pool.Exec(ctx,
		`DELETE FROM reimbursements WHERE description LIKE '[TEST]%'`,
	)
}

// TestOptimisticLocking_Loan menguji flow loan approval locking.
func TestOptimisticLocking_Loan(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee not found: %v", err)
	}

	// Buat loan request
	var loanID string
	err = database.Pool.QueryRow(ctx, `
		INSERT INTO loans (employee_id, loan_type, amount, interest_rate, installment_count, installment_amount, total_amount, remaining_balance, purpose, status)
		VALUES ($1::uuid, 'regular'::loan_type, 1000000, 0, 12, 83333, 1000000, 1000000, '[TEST] Loan locking', 'pending'::loan_status)
		RETURNING id::text
	`, empID).Scan(&loanID)
	if err != nil {
		t.Fatalf("create loan: %v", err)
	}
	t.Logf("Created loan: %s", loanID)

	// Approve
	_, err = UpdateLoanStatus(ctx, loanID, "approved", "", "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("Approve loan should succeed: %v", err)
	}

	// Double approve
	_, err = UpdateLoanStatus(ctx, loanID, "approved", "", "00000000-0000-0000-0000-000000000002")
	if err == nil {
		t.Error("Second loan approve should fail")
	} else {
		t.Logf("Second loan approve correctly rejected: %v", err)
	}

	// Cleanup
	database.Pool.Exec(ctx,
		`DELETE FROM loans WHERE description LIKE '[TEST]%'`,
	)
}

// TestOptimisticLocking_Overtime menguji flow overtime approval locking.
func TestOptimisticLocking_Overtime(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	var empID, wsID string
	err := database.Pool.QueryRow(ctx,
		`SELECT e.id::text, ws.id::text FROM employees e
		 CROSS JOIN work_schedules ws
		 WHERE e.employee_id = 'EMP-001' AND ws.name = '5 Hari Kerja'
		 LIMIT 1`,
	).Scan(&empID, &wsID)
	if err != nil {
		t.Skipf("Skip — test data tidak lengkap: %v", err)
	}

	// Buat overtime request
	var ovtID string
	err = database.Pool.QueryRow(ctx, `
		INSERT INTO overtime_requests (employee_id, overtime_date, start_time, end_time,
			overtime_type, reason, total_hours, status)
		VALUES ($1::uuid, CURRENT_DATE, '08:00', '10:00',
			'weekday'::overtime_type, '[TEST] Overtime locking', 2, 'pending'::overtime_status)
		RETURNING id::text
	`, empID).Scan(&ovtID)
	if err != nil {
		t.Fatalf("create overtime: %v", err)
	}
	t.Logf("Created overtime: %s", ovtID)

	// Simulasikan 2 concurrent approve
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(suffix string) {
			defer wg.Done()
			_, err := UpdateOvertimeStatus(ctx, ovtID, "approved", "", "00000000-0000-0000-0000-00000000000"+suffix)
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				t.Logf("Overtime approve attempt failed: %v", err)
			}
			mu.Unlock()
		}(string(rune('1' + i)))
	}
	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful overtime approve, got %d", successCount)
	}

	// Cleanup
	database.Pool.Exec(ctx,
		`DELETE FROM overtime_requests WHERE reason LIKE '[TEST]%'`,
	)
}
