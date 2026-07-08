// ============================================================
// Integration Test — Resign & Exit Clearance Flow
//
// Build tag: integration (go test -tags=integration)
// Skip:      go test -short
// Env:       RUN_INTEGRATION_TESTS=true
//
// Requires running PostgreSQL with migrations + seed data.
// Use: docker-compose up -d db migrate && go run cmd/seed/main.go
// Then: RUN_INTEGRATION_TESTS=true go test -tags=integration -run TestResign
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

func TestResign_CRUD_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test resign & clearance
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM resign_requests WHERE reason LIKE '[TEST]%'`)
	}()

	// Cari employee dari seed data
	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-005' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee not found: %v", err)
	}
	t.Logf("Using employee: %s", empID)

	// Create resign request
	req := &models.CreateResignRequest{
		LastWorkingDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		Reason:          "[TEST] Pindah ke perusahaan lain",
		ResignType:      "voluntary",
	}

	r, err := CreateResignRequest(ctx, empID, req)
	if err != nil {
		t.Fatalf("CreateResignRequest failed: %v", err)
	}
	t.Logf("Created resign request: %s", r.ID.String())

	// Verify default status
	if r.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", r.Status)
	}
	if r.ResignType != "voluntary" {
		t.Errorf("Expected resign_type 'voluntary', got '%s'", r.ResignType)
	}

	// Create exit clearance items
	err = CreateExitClearanceItems(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("CreateExitClearanceItems failed: %v", err)
	}

	// List clearance items
	items, err := ListExitClearanceItems(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("ListExitClearanceItems failed: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("Expected at least 1 clearance item, got 0")
	}
	t.Logf("Created %d clearance items", len(items))

	// Check default items names
	defaultNames := DefaultExitClearanceItems()
	if len(items) != len(defaultNames) {
		t.Errorf("Expected %d clearance items, got %d", len(defaultNames), len(items))
	}

	// Check all items are unchecked initially
	for _, item := range items {
		if item.IsChecked {
			t.Errorf("Expected item '%s' to be unchecked initially", item.ItemName)
		}
	}

	// Update a clearance item (check it)
	itemID := items[0].ID.String()
	err = UpdateExitClearanceItem(ctx, itemID, empID, true)
	if err != nil {
		t.Fatalf("UpdateExitClearanceItem failed: %v", err)
	}

	// Verify item is now checked
	items, err = ListExitClearanceItems(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("ListExitClearanceItems after update failed: %v", err)
	}
	found := false
	for _, item := range items {
		if item.ID.String() == itemID {
			found = true
			if !item.IsChecked {
				t.Error("Expected item to be checked after update")
			}
			break
		}
	}
	if !found {
		t.Error("Updated item not found in list")
	}

	// Test CheckAllClearanceItemsChecked - should be false since not all items checked
	allDone, err := CheckAllClearanceItemsChecked(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("CheckAllClearanceItemsChecked failed: %v", err)
	}
	if allDone {
		t.Error("Expected allDone=false since not all items are checked")
	}
}

func TestResign_Approve_Reject_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test resign
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM resign_requests WHERE reason LIKE '[TEST]%'`)
	}()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-005' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee not found: %v", err)
	}

	// Create resign request
	req := &models.CreateResignRequest{
		LastWorkingDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		Reason:          "[TEST] Mengundurkan diri - approve test",
		ResignType:      "voluntary",
	}
	r, err := CreateResignRequest(ctx, empID, req)
	if err != nil {
		t.Fatalf("CreateResignRequest failed: %v", err)
	}
	t.Logf("Created resign: %s", r.ID.String())

	// Reject (with reason)
	err = UpdateResignRequestStatus(ctx, r.ID.String(), "rejected", empID, "Masa kontrak belum selesai")
	if err != nil {
		t.Fatalf("UpdateResignRequestStatus (reject) failed: %v", err)
	}

	// Verify rejected status
	got, err := GetResignRequest(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("GetResignRequest after reject failed: %v", err)
	}
	if got.Status != "rejected" {
		t.Errorf("Expected status 'rejected', got '%s'", got.Status)
	}
	// Verify rejection reason
	gotRejectReason := got.RejectionReason
	if gotRejectReason != "Masa kontrak belum selesai" {
		t.Errorf("Expected rejection reason 'Masa kontrak belum selesai', got '%s'", gotRejectReason)
	}
	t.Logf("Resign rejected: %s", gotRejectReason)

	// Create another resign for approve flow
	r2, err := CreateResignRequest(ctx, empID, &models.CreateResignRequest{
		LastWorkingDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		Reason:          "[TEST] Mengundurkan diri - approve with clearance test",
		ResignType:      "voluntary",
	})
	if err != nil {
		t.Fatalf("CreateResignRequest 2 failed: %v", err)
	}

	// Create clearance and check all items
	err = CreateExitClearanceItems(ctx, r2.ID.String())
	if err != nil {
		t.Fatalf("CreateExitClearanceItems failed: %v", err)
	}

	items, err := ListExitClearanceItems(ctx, r2.ID.String())
	if err != nil {
		t.Fatalf("ListExitClearanceItems failed: %v", err)
	}
	for _, item := range items {
		err = UpdateExitClearanceItem(ctx, item.ID.String(), empID, true)
		if err != nil {
			t.Fatalf("UpdateExitClearanceItem failed: %v", err)
		}
	}

	// Now all items should be checked
	allDone, err := CheckAllClearanceItemsChecked(ctx, r2.ID.String())
	if err != nil {
		t.Fatalf("CheckAllClearanceItemsChecked failed: %v", err)
	}
	if !allDone {
		t.Error("Expected allDone=true after checking all items")
	}

	// Approve
	err = UpdateResignRequestStatus(ctx, r2.ID.String(), "approved", empID, "")
	if err != nil {
		t.Fatalf("UpdateResignRequestStatus (approve) failed: %v", err)
	}

	// Process resignation (deactivate employee)
	err = ProcessEmployeeResignation(ctx, r2.ID.String())
	if err != nil {
		t.Fatalf("ProcessEmployeeResignation failed: %v", err)
	}

	// Verify employee is deactivated
	var isActive bool
	var deletedAt interface{}
	err = database.Pool.QueryRow(ctx,
		`SELECT is_active, deleted_at FROM employees WHERE id::text = $1`,
		empID,
	).Scan(&isActive, &deletedAt)
	if err != nil {
		t.Fatalf("Query employee after resign failed: %v", err)
	}
	if isActive {
		t.Error("Expected employee to be inactive after resignation")
	}
	if deletedAt == nil {
		t.Error("Expected deleted_at to be set after resignation")
	}
	t.Logf("Employee deactivated successfully (is_active=%v, deleted_at=%v)", isActive, deletedAt)

	// Reactivate employee for other tests
	_, err = database.Pool.Exec(ctx,
		`UPDATE employees SET deleted_at = NULL, is_active = TRUE, updated_at = NOW() WHERE id::text = $1`,
		empID,
	)
	if err != nil {
		t.Logf("Warning: Could not reactivate employee: %v", err)
	}
}

func TestResign_GetResignIDByItemID(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test resign
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM resign_requests WHERE reason LIKE '[TEST]%'`)
	}()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-005' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee not found: %v", err)
	}

	// Create resign with clearance
	r, err := CreateResignRequest(ctx, empID, &models.CreateResignRequest{
		LastWorkingDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		Reason:          "[TEST] GetResignID test",
		ResignType:      "voluntary",
	})
	if err != nil {
		t.Fatalf("CreateResignRequest failed: %v", err)
	}

	err = CreateExitClearanceItems(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("CreateExitClearanceItems failed: %v", err)
	}

	items, err := ListExitClearanceItems(ctx, r.ID.String())
	if err != nil {
		t.Fatalf("ListExitClearanceItems failed: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("No clearance items created")
	}

	// Test GetResignIDByItemID
	var resignID string
	err = GetResignIDByItemID(ctx, items[0].ID.String(), &resignID)
	if err != nil {
		t.Fatalf("GetResignIDByItemID failed: %v", err)
	}
	if resignID != r.ID.String() {
		t.Errorf("ResignID mismatch: got '%s', want '%s'", resignID, r.ID.String())
	}
	t.Logf("GetResignIDByItemID works: item -> resign %s", resignID)
}

func TestResign_DefaultExitClearanceItems(t *testing.T) {
	items := DefaultExitClearanceItems()
	if len(items) == 0 {
		t.Error("Expected at least 1 default clearance item")
	}
	t.Logf("Default clearance items: %d items", len(items))

	expectedNames := []string{
		"Pengembalian Aset Perusahaan",
		"Pengembalian Seragam/ID Card",
		"Clearance Keuangan",
		"Clearance BPJS",
		"Exit Interview",
		"Pengembalian Akses",
		"Serah Terima Pekerjaan",
		"Penyelesaian Cuti",
		"Penyelesaian Hak",
		"Surat Keterangan Kerja",
	}
	if len(items) != len(expectedNames) {
		t.Errorf("Expected %d items, got %d", len(expectedNames), len(items))
	}

	for i, name := range expectedNames {
		if i < len(items) && items[i].Name != name {
			t.Errorf("Item %d: expected name '%s', got '%s'", i+1, name, items[i].Name)
		}
	}
}
