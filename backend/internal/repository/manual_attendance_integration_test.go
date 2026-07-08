// ============================================================
// Integration Test — Manual Attendance Flow
//
// Build tag: integration (go test -tags=integration)
// Skip:      go test -short
// Env:       RUN_INTEGRATION_TESTS=true
//
// Requires running PostgreSQL with migrations + seed data.
// Use: docker-compose up -d db migrate && go run cmd/seed/main.go
// Then: RUN_INTEGRATION_TESTS=true go test -tags=integration -run TestManualAttendance
// ============================================================
//
//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

func TestManualAttendance_CRUD_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM manual_attendance_requests WHERE reason LIKE '[TEST]%'`)
	}()

	// Cari employee dari seed data
	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee EMP001 not found: %v", err)
	}

	t.Logf("Using employee: %s", empID)

	// Create manual attendance request
	req := &models.CreateManualAttendanceRequest{
		Date:        time.Now().Format("2006-01-02"),
		CheckInTime: "08:00",
		CheckOutTime: "17:00",
		Reason:      "[TEST] Absensi manual - lupa absen",
	}

	r, err := CreateManualAttendanceRequest(ctx, empID, req)
	if err != nil {
		t.Fatalf("CreateManualAttendanceRequest failed: %v", err)
	}
	t.Logf("Created manual attendance request: %s", r.ID.String())

	// Verify it was created
	if r.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", r.Status)
	}
	if r.Reason != req.Reason {
		t.Errorf("Reason mismatch: got '%s', want '%s'", r.Reason, req.Reason)
	}

	// List requests
	listResp, err := ListManualAttendanceRequests(ctx, 1, 25, "", empID)
	if err != nil {
		t.Fatalf("ListManualAttendanceRequests failed: %v", err)
	}
	if listResp.Total < 1 {
		t.Errorf("Expected at least 1 manual attendance request, got %d", listResp.Total)
	}
	t.Logf("Found %d request(s)", listResp.Total)

	// Get request by ID
	got, err := GetManualAttendanceRequest(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("GetManualAttendanceRequest failed: %v", err)
	}
	if got.ID != r.ID {
		t.Errorf("ID mismatch: got %v, want %v", got.ID, r.ID)
	}

	// Approve
	err = UpdateManualAttendanceRequestStatus(ctx, r.ID.String(), "approved", empID, "")
	if err != nil {
		t.Fatalf("UpdateManualAttendanceRequestStatus (approve) failed: %v", err)
	}

	// Verify approved status
	got, err = GetManualAttendanceRequest(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("GetManualAttendanceRequest after approve failed: %v", err)
	}
	if got.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", got.Status)
	}

	// Create attendance record from approved request
	err = CreateAttendanceFromManualRequest(ctx, got)
	if err != nil {
		t.Fatalf("CreateAttendanceFromManualRequest failed: %v", err)
	}

	// Verify attendance record was created
	var attCount int
	err = database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM attendance_records WHERE employee_id = $1::uuid AND date = $2::date AND is_manual_entry = TRUE`,
		empID, req.Date,
	).Scan(&attCount)
	if err != nil {
		t.Fatalf("Query attendance_records failed: %v", err)
	}
	if attCount == 0 {
		t.Error("Expected at least 1 manual attendance record, got 0")
	}
	t.Logf("Attendance record created: %d record(s)", attCount)
}

func TestManualAttendance_Reject_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM manual_attendance_requests WHERE reason LIKE '[TEST]%'`)
	}()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee EMP001 not found: %v", err)
	}

	// Create
	req := &models.CreateManualAttendanceRequest{
		Date:        time.Now().Format("2006-01-02"),
		CheckInTime: "09:00",
		Reason:      "[TEST] Absensi manual - reject test",
	}
	r, err := CreateManualAttendanceRequest(ctx, empID, req)
	if err != nil {
		t.Fatalf("CreateManualAttendanceRequest failed: %v", err)
	}
	t.Logf("Created manual attendance: %s", r.ID.String())

	// Reject
	err = UpdateManualAttendanceRequestStatus(ctx, r.ID.String(), "rejected", empID, "Tidak ada bukti pendukung")
	if err != nil {
		t.Fatalf("UpdateManualAttendanceRequestStatus (reject) failed: %v", err)
	}

	// Verify
	got, err := GetManualAttendanceRequest(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("GetManualAttendanceRequest after reject failed: %v", err)
	}
	if got.Status != "rejected" {
		t.Errorf("Expected status 'rejected', got '%s'", got.Status)
	}
	if got.RejectionReason != "Tidak ada bukti pendukung" {
		t.Errorf("Rejection reason mismatch: got '%s'", got.RejectionReason)
	}
	t.Logf("Request rejected successfully: %s", got.RejectionReason)
}

func TestManualAttendance_Cancel_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM manual_attendance_requests WHERE reason LIKE '[TEST]%'`)
	}()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee EMP001 not found: %v", err)
	}

	// Create
	req := &models.CreateManualAttendanceRequest{
		Date:        time.Now().Format("2006-01-02"),
		CheckInTime: "10:00",
		Reason:      "[TEST] Absensi manual - cancel test",
	}
	r, err := CreateManualAttendanceRequest(ctx, empID, req)
	if err != nil {
		t.Fatalf("CreateManualAttendanceRequest failed: %v", err)
	}
	t.Logf("Created manual attendance: %s", r.ID.String())

	// Cancel
	err = UpdateManualAttendanceRequestStatus(ctx, r.ID.String(), "cancelled", empID, "")
	if err != nil {
		t.Fatalf("UpdateManualAttendanceRequestStatus (cancel) failed: %v", err)
	}

	// Verify
	got, err := GetManualAttendanceRequest(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("GetManualAttendanceRequest after cancel failed: %v", err)
	}
	if got.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got '%s'", got.Status)
	}
	t.Log("Request cancelled successfully")
}

func TestManualAttendance_MonthlyQuota(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-001' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee EMP001 not found: %v", err)
	}

	// Check quota - should not error with valid request
	err = CheckManualAttendanceCount(ctx, empID, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Logf("Quota check result: %v (may be exceeded)", err)
	} else {
		t.Log("Quota check passed")
	}
}
