//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

func TestRoster_SoftDelete_Recreate_UniqueConstraint(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test — set RUN_INTEGRATION_TESTS=true to run")
	}

	ctx := context.Background()

	// Clean up any test rosters first
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM department_rosters WHERE name LIKE '[TEST]%'`)
	}()

	// Get a test department
	var deptID string
	err := database.Pool.QueryRow(ctx, `SELECT id::text FROM departments LIMIT 1`).Scan(&deptID)
	if err != nil {
		t.Skip("Skip - no departments found in database")
	}

	// Get a test user/employee to act as creator
	var userID string
	err = database.Pool.QueryRow(ctx, `SELECT id::text FROM employees LIMIT 1`).Scan(&userID)
	if err != nil {
		t.Skip("Skip - no employees found in database")
	}

	rosterName := "[TEST] Roster Jan 2026"

	// 1. Create first roster
	req1 := &models.CreateDepartmentRosterRequest{
		DepartmentID: deptID,
		Name:         rosterName,
		Month:        1,
		Year:         2026,
		Notes:        "First test roster",
	}
	r1, err := CreateDepartmentRoster(ctx, req1, userID)
	if err != nil {
		t.Fatalf("Failed to create first roster: %v", err)
	}
	t.Logf("Created first roster ID: %s", r1.ID)

	// 2. Soft-delete the first roster
	err = DeleteDepartmentRoster(ctx, r1.ID.String(), userID)
	if err != nil {
		t.Fatalf("Failed to soft-delete first roster: %v", err)
	}
	t.Log("Soft-deleted first roster")

	// 3. Re-create the roster with the same name and department (should not trigger unique constraint error)
	req2 := &models.CreateDepartmentRosterRequest{
		DepartmentID: deptID,
		Name:         rosterName,
		Month:        1,
		Year:         2026,
		Notes:        "Recreated test roster",
	}
	r2, err := CreateDepartmentRoster(ctx, req2, userID)
	if err != nil {
		t.Fatalf("Failed to recreate roster after soft-delete: %v", err)
	}
	t.Logf("Recreated roster ID: %s", r2.ID)

	if r1.ID == r2.ID {
		t.Errorf("Expected different IDs for r1 and r2, got identical ID: %s", r1.ID)
	}
}
