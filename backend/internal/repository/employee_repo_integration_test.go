//go:build integration

package repository

import (
	"context"
	"os"
	"testing"
)

func TestEmployee_ListEmployees_StatusFilter(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test — set RUN_INTEGRATION_TESTS=true to run")
	}

	ctx := context.Background()

	// 1. Test filtering by status = "active"
	empsActive, totalActive, err := ListEmployees(ctx, 1, 100, "", "", "active", false)
	if err != nil {
		t.Fatalf("ListEmployees (active) failed: %v", err)
	}
	t.Logf("Active employees count: %d (total: %d)", len(empsActive), totalActive)
	for _, emp := range empsActive {
		if !emp.IsActive {
			t.Errorf("Expected employee %s to be active, but got is_active = false", emp.ID)
		}
	}

	// 2. Test filtering by status = "inactive"
	empsInactive, totalInactive, err := ListEmployees(ctx, 1, 100, "", "", "inactive", false)
	if err != nil {
		t.Fatalf("ListEmployees (inactive) failed: %v", err)
	}
	t.Logf("Inactive employees count: %d (total: %d)", len(empsInactive), totalInactive)
	for _, emp := range empsInactive {
		if emp.IsActive {
			t.Errorf("Expected employee %s to be inactive, but got is_active = true", emp.ID)
		}
	}

	// 3. Test filtering by employment_status enum (e.g., "tetap")
	empsTetap, totalTetap, err := ListEmployees(ctx, 1, 100, "", "", "tetap", false)
	if err != nil {
		t.Fatalf("ListEmployees (tetap) failed: %v", err)
	}
	t.Logf("Tetap employees count: %d (total: %d)", len(empsTetap), totalTetap)
	for _, emp := range empsTetap {
		if emp.EmploymentStatus != "tetap" {
			t.Errorf("Expected employee %s to have status 'tetap', but got '%s'", emp.ID, emp.EmploymentStatus)
		}
	}
}
