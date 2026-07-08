// ============================================================
// Integration Test — Employee Mutation CRUD & Approval Flow
//
// Build tag: integration (go test -tags=integration)
// Skip:      go test -short
// Env:       RUN_INTEGRATION_TESTS=true
//
// Requires running PostgreSQL with migrations + seed data.
// Use: docker-compose up -d db migrate && go run cmd/seed/main.go
// Then: RUN_INTEGRATION_TESTS=true go test -tags=integration -run TestMutation
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

func TestMutation_CRUD_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	// Cleanup: hapus data test mutation
	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM employee_mutations WHERE reason LIKE '[TEST]%'`)
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

	// Cari department & position untuk new values
	var newDeptID, newPosID string
	database.Pool.QueryRow(ctx, `SELECT id::text FROM departments WHERE code = 'HR' LIMIT 1`).Scan(&newDeptID)
	database.Pool.QueryRow(ctx, `SELECT id::text FROM positions WHERE name = 'HR Staff' LIMIT 1`).Scan(&newPosID)

	repo := NewMutationRepo()

	// Create mutation request
	t.Run("Create mutation", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:      empID,
			MutationType:    "transfer",
			NewDepartmentID: newDeptID,
			NewPositionID:   newPosID,
			Reason:          "[TEST] Mutasi ke departemen HR",
			EffectiveDate:   time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			Notes:           "Test integration",
		}

		m, err := repo.Create(ctx, req, empID)
		if err != nil {
			t.Fatalf("Create mutation failed: %v", err)
		}
		t.Logf("Created mutation: %s (type: %s)", m.ID, m.MutationType)

		if m.Status != "pending" {
			t.Errorf("Expected status 'pending', got '%s'", m.Status)
		}
		if m.MutationType != "transfer" {
			t.Errorf("Expected mutation_type 'transfer', got '%s'", m.MutationType)
		}
		if m.Reason != "[TEST] Mutasi ke departemen HR" {
			t.Errorf("Expected reason '[TEST] Mutasi ke departemen HR', got '%s'", m.Reason)
		}

		// Update old values (like the service does)
		oldDeptID, oldPosID, _, oldStatus, oldSalary, err := repo.GetEmployeeData(ctx, empID)
		if err != nil {
			t.Fatalf("GetEmployeeData failed: %v", err)
		}
		err = repo.UpdateOldValues(ctx, m.ID, oldDeptID, oldPosID, "", oldStatus, oldSalary)
		if err != nil {
			t.Fatalf("UpdateOldValues failed: %v", err)
		}

		// Verify old values are set
		got, err := repo.GetByID(ctx, m.ID)
		if err != nil {
			t.Fatalf("GetByID after update old values failed: %v", err)
		}
		if got.OldDepartmentName == "" {
			t.Error("Expected old_department_name to be set after UpdateOldValues")
		}
	})

	// Create another mutation for approve/reject test
	var mutationID string
	t.Run("Create mutation for approval test", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:      empID,
			MutationType:    "status_change",
			NewEmploymentStatus: "tetap",
			Reason:          "[TEST] Perubahan status ke tetap",
			EffectiveDate:   time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		}

		m, err := repo.Create(ctx, req, empID)
		if err != nil {
			t.Fatalf("Create mutation for approval failed: %v", err)
		}
		mutationID = m.ID
		t.Logf("Created mutation for approval: %s", mutationID)
	})

	t.Run("Approve mutation", func(t *testing.T) {
		if mutationID == "" {
			t.Fatal("No mutation ID from previous step")
		}

		// Update status to approved
		err := repo.UpdateStatus(ctx, mutationID, "approved", empID, "")
		if err != nil {
			t.Fatalf("UpdateStatus (approve) failed: %v", err)
		}

		// Apply mutation changes to employee
		err = repo.ApplyMutation(ctx, mutationID)
		if err != nil {
			t.Fatalf("ApplyMutation failed: %v", err)
		}

		// Verify approved status
		got, err := repo.GetByID(ctx, mutationID)
		if err != nil {
			t.Fatalf("GetByID after approve failed: %v", err)
		}
		if got.Status != "approved" {
			t.Errorf("Expected status 'approved', got '%s'", got.Status)
		}
		if got.ApprovedByName == "" {
			t.Error("Expected approved_by_name to be set")
		}
		t.Logf("Mutation approved successfully")
	})

	// Create mutation for reject test
	var rejectMutID string
	t.Run("Create mutation for reject test", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   empID,
			MutationType: "salary_change",
			NewBaseSalary: func() *float64 { f := 10000000.0; return &f }(),
			Reason:       "[TEST] Reject test - kenaikan gaji",
			EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		}

		m, err := repo.Create(ctx, req, empID)
		if err != nil {
			t.Fatalf("Create mutation for reject failed: %v", err)
		}
		rejectMutID = m.ID
		t.Logf("Created mutation for reject: %s", rejectMutID)
	})

	t.Run("Reject mutation", func(t *testing.T) {
		if rejectMutID == "" {
			t.Fatal("No mutation ID for reject test")
		}

		rejectionReason := "Tidak memenuhi syarat masa kerja minimal"
		err := repo.UpdateStatus(ctx, rejectMutID, "rejected", empID, rejectionReason)
		if err != nil {
			t.Fatalf("UpdateStatus (reject) failed: %v", err)
		}

		got, err := repo.GetByID(ctx, rejectMutID)
		if err != nil {
			t.Fatalf("GetByID after reject failed: %v", err)
		}
		if got.Status != "rejected" {
			t.Errorf("Expected status 'rejected', got '%s'", got.Status)
		}
		if got.RejectionReason != rejectionReason {
			t.Errorf("Expected rejection_reason '%s', got '%s'", rejectionReason, got.RejectionReason)
		}
		t.Logf("Mutation rejected with reason: %s", got.RejectionReason)
	})

	// Create mutation for cancel test
	var cancelMutID string
	t.Run("Create mutation for cancel test", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   empID,
			MutationType: "position_change",
			NewPositionID: newPosID,
			Reason:       "[TEST] Cancel test - pindah posisi",
			EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		}

		m, err := repo.Create(ctx, req, empID)
		if err != nil {
			t.Fatalf("Create mutation for cancel failed: %v", err)
		}
		cancelMutID = m.ID
		t.Logf("Created mutation for cancel: %s", cancelMutID)
	})

	t.Run("Cancel mutation", func(t *testing.T) {
		if cancelMutID == "" {
			t.Fatal("No mutation ID for cancel test")
		}

		err := repo.UpdateStatus(ctx, cancelMutID, "cancelled", empID, "")
		if err != nil {
			t.Fatalf("UpdateStatus (cancel) failed: %v", err)
		}

		got, err := repo.GetByID(ctx, cancelMutID)
		if err != nil {
			t.Fatalf("GetByID after cancel failed: %v", err)
		}
		if got.Status != "cancelled" {
			t.Errorf("Expected status 'cancelled', got '%s'", got.Status)
		}
		t.Logf("Mutation cancelled successfully")
	})

	t.Run("List mutations with filter", func(t *testing.T) {
		mutations, total, err := repo.List(ctx, 1, 25, "", empID)
		if err != nil {
			t.Fatalf("List mutations failed: %v", err)
		}
		if total == 0 {
			t.Error("Expected at least 1 mutation, got 0")
		}
		t.Logf("List mutations: %d total (filtered by employee)", total)

		// Test status filter
		approvedMuts, approvedTotal, err := repo.List(ctx, 1, 25, "approved", "")
		if err != nil {
			t.Fatalf("List mutations with status filter failed: %v", err)
		}
		t.Logf("Approved mutations: %d total", approvedTotal)
		for _, m := range approvedMuts {
			if m.Status != "approved" {
				t.Errorf("Expected all mutations to have status 'approved', got '%s'", m.Status)
			}
		}
	})

	t.Run("ListAll for export", func(t *testing.T) {
		mutations, err := repo.ListAll(ctx, "", empID)
		if err != nil {
			t.Fatalf("ListAll mutations failed: %v", err)
		}
		if len(mutations) == 0 {
			t.Error("Expected at least 1 mutation from ListAll, got 0")
		}
		t.Logf("ListAll mutations: %d records", len(mutations))
	})

	t.Run("GetByID non-existent returns error", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
		if err == nil {
			t.Error("Expected error for non-existent mutation, got nil")
		}
	})
}

func TestMutation_StatusChange_Flow(t *testing.T) {
	skipIfNotIntegration(t)

	ctx := context.Background()

	defer func() {
		database.Pool.Exec(ctx, `DELETE FROM employee_mutations WHERE reason LIKE '[TEST]%'`)
	}()

	var empID string
	err := database.Pool.QueryRow(ctx,
		`SELECT id::text FROM employees WHERE employee_id = 'EMP-005' LIMIT 1`,
	).Scan(&empID)
	if err != nil {
		t.Skipf("Skip — seed employee not found: %v", err)
	}

	repo := NewMutationRepo()

	req := &models.CreateMutationRequest{
		EmployeeID:   empID,
		MutationType: "salary_change",
		NewBaseSalary: func() *float64 { f := 12000000.0; return &f }(),
		Reason:       "[TEST] Double approve test",
		EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}
	m, err := repo.Create(ctx, req, empID)
	if err != nil {
		t.Fatalf("Create mutation failed: %v", err)
	}
	t.Logf("Created mutation: %s", m.ID)

	// First approve
	err = repo.UpdateStatus(ctx, m.ID, "approved", empID, "")
	if err != nil {
		t.Fatalf("First approve failed: %v", err)
	}

	// Apply mutation
	err = repo.ApplyMutation(ctx, m.ID)
	if err != nil {
		t.Fatalf("ApplyMutation failed: %v", err)
	}

	// Second approve should still work at DB level (service handles the check)
	// Just verify status doesn't change from 'rejected' -> 'approved' after being approved
	got, err := repo.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", got.Status)
	}

	// Verify status update works for rejected mutation (can change approved -> rejected via DB)
	// In practice, service prevents this, but DB allows it
	err = repo.UpdateStatus(ctx, m.ID, "rejected", empID, "Test double reject")
	if err != nil {
		t.Fatalf("Second status update (rejected) failed: %v", err)
	}
	got2, err := repo.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetByID after second update failed: %v", err)
	}
	t.Logf("Double status change: approved -> rejected -> current status: %s", got2.Status)
}
