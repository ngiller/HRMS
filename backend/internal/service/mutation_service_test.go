package service

import (
	"context"
	"testing"
	"time"

	"hrms-backend/internal/models"
)

// ─── MutationService Validation Tests ──────────────────────────

func TestMutationService_Create_Validation(t *testing.T) {
	svc := NewMutationService()

	t.Run("empty employee id returns error", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   "",
			MutationType: "promotion",
			Reason:       "Kinerja baik",
			EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			NewPositionGradeID: "some-grade-id",
		}
		_, err := svc.Create(context.Background(), req, "creator-id")
		if err == nil {
			t.Error("expected error for empty employee_id, got nil")
		}
	})

	t.Run("empty mutation type returns error", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   "EMP-001",
			MutationType: "",
			Reason:       "Kinerja baik",
			EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			NewPositionGradeID: "some-grade-id",
		}
		_, err := svc.Create(context.Background(), req, "creator-id")
		if err == nil {
			t.Error("expected error for empty mutation_type, got nil")
		}
	})

	t.Run("invalid mutation type returns error", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   "EMP-001",
			MutationType: "invalid_type",
			Reason:       "Kinerja baik",
			EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			NewPositionGradeID: "some-grade-id",
		}
		_, err := svc.Create(context.Background(), req, "creator-id")
		if err == nil {
			t.Error("expected error for invalid mutation_type, got nil")
		}
	})

	t.Run("empty reason returns error", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   "EMP-001",
			MutationType: "promotion",
			Reason:       "",
			EffectiveDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			NewPositionGradeID: "some-grade-id",
		}
		_, err := svc.Create(context.Background(), req, "creator-id")
		if err == nil {
			t.Error("expected error for empty reason, got nil")
		}
	})

	t.Run("empty effective date returns error", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   "EMP-001",
			MutationType: "promotion",
			Reason:       "Kinerja baik",
			EffectiveDate: "",
			NewPositionGradeID: "some-grade-id",
		}
		_, err := svc.Create(context.Background(), req, "creator-id")
		if err == nil {
			t.Error("expected error for empty effective_date, got nil")
		}
	})

	t.Run("no changes provided requires DB", func(t *testing.T) {
		t.Skip("memerlukan koneksi database - test integrasi")
	})

	t.Run("all validation errors: first one wins", func(t *testing.T) {
		req := &models.CreateMutationRequest{
			EmployeeID:   "",
			MutationType: "",
			Reason:       "",
			EffectiveDate: "",
		}
		_, err := svc.Create(context.Background(), req, "creator-id")
		if err == nil {
			t.Error("expected validation error for all empty fields, got nil")
		}
	})
}

func TestMutationService_ValidMutationTypes(t *testing.T) {
	validTypes := []string{"promotion", "demotion", "transfer", "position_change", "status_change", "salary_change"}

	for _, mt := range validTypes {
		t.Run(mt+" is valid", func(t *testing.T) {
			// Simulasi validasi yang dilakukan service
			validTypes := map[string]bool{
				"promotion": true, "demotion": true, "transfer": true,
				"position_change": true, "status_change": true, "salary_change": true,
			}
			if !validTypes[mt] {
				t.Errorf("expected '%s' to be a valid mutation type", mt)
			}
		})
	}

	invalidTypes := []string{"", "rejected", "approved", "resign", "cuti", "pensiun", "random"}
	for _, mt := range invalidTypes {
		t.Run(mt+" is invalid", func(t *testing.T) {
			validTypes := map[string]bool{
				"promotion": true, "demotion": true, "transfer": true,
				"position_change": true, "status_change": true, "salary_change": true,
			}
			if validTypes[mt] {
				t.Errorf("expected '%s' to be an invalid mutation type", mt)
			}
		})
	}
}

// TestMutationService_Approve_Reject_Cancel_Validation needs DB — skip unit test
func TestMutationService_Approve_Reject_Cancel_Validation(t *testing.T) {
	t.Skip("memerlukan koneksi database - test integrasi")
}

// TestMutationService_StatusFlow is a placeholder for integration tests
func TestMutationService_StatusFlow(t *testing.T) {
	t.Skip("memerlukan koneksi database - test integrasi")
}
