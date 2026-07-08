package service

import (
	"context"
	"testing"

	"hrms-backend/internal/models"
)

func TestGetEntityLabel(t *testing.T) {
	svc := NewApprovalWorkflowService()

	tests := []struct {
		entityType string
		want       string
	}{
		{"leave", "Cuti"},
		{"overtime", "Lembur"},
		{"reimbursement", "Reimbursement"},
		{"shift_change", "Perubahan Shift"},
		{"loan", "Pinjaman"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.entityType, func(t *testing.T) {
			got := svc.getEntityLabel(tt.entityType)
			if got != tt.want {
				t.Errorf("getEntityLabel(%q) = %q, want %q", tt.entityType, got, tt.want)
			}
		})
	}
}

func TestCreateWorkflowValidation(t *testing.T) {
	svc := NewApprovalWorkflowService()

	t.Run("empty entity type returns error", func(t *testing.T) {
		req := &models.CreateApprovalWorkflowReq{
			EntityType: "",
			Name:       "Test Workflow",
		}
		_, err := svc.CreateWorkflow(context.Background(), req)
		if err == nil {
			t.Error("expected error for empty entity type, got nil")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		req := &models.CreateApprovalWorkflowReq{
			EntityType: "leave",
			Name:       "",
		}
		_, err := svc.CreateWorkflow(context.Background(), req)
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})
}

func TestAddWorkflowStepValidation(t *testing.T) {
	svc := NewApprovalWorkflowService()

	t.Run("empty workflow ID returns error", func(t *testing.T) {
		req := &models.CreateApprovalWorkflowStepReq{
			WorkflowID:   "",
			ApproverType: "approval_line",
			StepOrder:    1,
		}
		_, err := svc.AddWorkflowStep(context.Background(), req)
		if err == nil {
			t.Error("expected error for empty workflow ID, got nil")
		}
	})

	t.Run("empty approver type returns error", func(t *testing.T) {
		req := &models.CreateApprovalWorkflowStepReq{
			WorkflowID:   "some-id",
			ApproverType: "",
			StepOrder:    1,
		}
		_, err := svc.AddWorkflowStep(context.Background(), req)
		if err == nil {
			t.Error("expected error for empty approver type, got nil")
		}
	})

	t.Run("invalid step order returns error", func(t *testing.T) {
		req := &models.CreateApprovalWorkflowStepReq{
			WorkflowID:   "some-id",
			ApproverType: "approval_line",
			StepOrder:    0,
		}
		_, err := svc.AddWorkflowStep(context.Background(), req)
		if err == nil {
			t.Error("expected error for step order < 1, got nil")
		}
	})
}

func TestDeleteWorkflowStepValidation(t *testing.T) {
	svc := NewApprovalWorkflowService()

	t.Run("empty step ID returns error", func(t *testing.T) {
		err := svc.DeleteWorkflowStep(context.Background(), "")
		if err == nil {
			t.Error("expected error for empty step ID, got nil")
		}
	})
}

func TestDeleteWorkflowValidation(t *testing.T) {
	svc := NewApprovalWorkflowService()

	t.Run("empty workflow ID returns error", func(t *testing.T) {
		err := svc.DeleteWorkflow(context.Background(), "")
		if err == nil {
			t.Error("expected error for empty workflow ID, got nil")
		}
	})
}
