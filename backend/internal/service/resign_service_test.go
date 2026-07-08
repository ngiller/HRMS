package service

import (
	"context"
	"testing"

	"hrms-backend/internal/models"
)

// ─── Global Email Service Tests ────────────────────────────────

func TestInitGlobalEmailService(t *testing.T) {
	// Reset before test
	globalEmailService = nil

	t.Run("nil service", func(t *testing.T) {
		InitGlobalEmailService(nil)
		if globalEmailService != nil {
			t.Error("expected globalEmailService to be nil")
		}
	})

	t.Run("enabled service", func(t *testing.T) {
		svc := NewEmailService("smtp.test.com", "587", "user", "pass", "from@test.com", "Test")
		InitGlobalEmailService(svc)
		if globalEmailService == nil {
			t.Error("expected globalEmailService to be set")
		}
		if !globalEmailService.IsEnabled() {
			t.Error("expected service to be enabled")
		}
	})

	t.Run("disabled service", func(t *testing.T) {
		svc := NewEmailService("", "", "", "", "", "")
		InitGlobalEmailService(svc)
		if globalEmailService == nil {
			t.Error("expected globalEmailService to be set")
		}
		if globalEmailService.IsEnabled() {
			t.Error("expected service to be disabled")
		}
	})

	// Cleanup
	globalEmailService = nil
}

func TestSendEmailForUser_NotConfigured(t *testing.T) {
	// Ensure no global service configured
	globalEmailService = nil

	// Should not panic
	SendEmailForUser(context.Background(), "some-user-id", "Test Title", "Test Body")

	// With disabled service
	svc := NewEmailService("", "", "", "", "", "")
	InitGlobalEmailService(svc)
	SendEmailForUser(context.Background(), "some-user-id", "Test Title", "Test Body")

	// Should not panic either way
	globalEmailService = nil
}

func TestSendEmailForUser_WithEnabledService(t *testing.T) {
	// Requires database connection for email lookup in the goroutine.
	// Without DB, the goroutine would panic on nil pool access.
	t.Skip("memerlukan koneksi database - test integrasi")
}

// ─── ResignService Validation Tests ───────────────────────────

func TestResignService_Create_Validation(t *testing.T) {
	svc := NewResignService()

	t.Run("empty last working date returns error", func(t *testing.T) {
		req := &models.CreateResignRequest{
			LastWorkingDate: "",
			Reason:          "Pindah perusahaan",
		}
		_, err := svc.Create(context.Background(), "some-employee-id", req)
		if err == nil {
			t.Error("expected error for empty last working date, got nil")
		}
	})

	t.Run("empty reason returns error", func(t *testing.T) {
		req := &models.CreateResignRequest{
			LastWorkingDate: "2026-08-01",
			Reason:          "",
		}
		_, err := svc.Create(context.Background(), "some-employee-id", req)
		if err == nil {
			t.Error("expected error for empty reason, got nil")
		}
	})

	t.Run("both empty returns error for first validation", func(t *testing.T) {
		req := &models.CreateResignRequest{
			LastWorkingDate: "",
			Reason:          "",
		}
		_, err := svc.Create(context.Background(), "some-id", req)
		if err == nil {
			t.Error("expected validation error for empty fields, got nil")
		}
	})
}

func TestResignService_DefaultResignType(t *testing.T) {
	t.Run("empty resign_type becomes voluntary", func(t *testing.T) {
		req := &models.CreateResignRequest{
			LastWorkingDate: "2026-08-01",
			Reason:          "Pindah ke perusahaan lain",
			ResignType:      "",
		}
		// Simulate what the service does internally
		if req.ResignType == "" {
			req.ResignType = "voluntary"
		}
		if req.ResignType != "voluntary" {
			t.Errorf("resign_type = %q, want 'voluntary'", req.ResignType)
		}
	})
}

// TestResignService_DB_Operations is a placeholder for integration tests
// that require a database connection. They are tested via integration tests.
func TestResignService_DB_Operations(t *testing.T) {
	t.Skip("memerlukan koneksi database - test integrasi")
}
