package models

import (
	"time"

	"github.com/google/uuid"
)

// ─── Approval Workflow ─────────────────────────────────────────

type ApprovalWorkflow struct {
	ID          uuid.UUID  `json:"id"`
	EntityType  string     `json:"entity_type"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	Steps       []ApprovalWorkflowStep `json:"steps,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type ApprovalWorkflowSummary struct {
	ID          uuid.UUID `json:"id"`
	EntityType  string    `json:"entity_type"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	StepCount   int       `json:"step_count"`
}

type CreateApprovalWorkflowReq struct {
	EntityType  string `json:"entity_type"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateApprovalWorkflowReq struct {
	EntityType  string `json:"entity_type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// ─── Approval Workflow Step ────────────────────────────────────

type ApprovalWorkflowStep struct {
	ID                uuid.UUID  `json:"id"`
	WorkflowID        uuid.UUID  `json:"workflow_id"`
	StepOrder         int        `json:"step_order"`
	ApproverType      string     `json:"approver_type"`    // 'approval_line', 'hr_manager', 'finance', 'director', 'department_head', 'specific_role', 'manager', 'specific_employee'
	ApproverRoleID    *uuid.UUID `json:"approver_role_id,omitempty"`
	ApproverEmployeeID *uuid.UUID `json:"approver_employee_id,omitempty"` // for specific_employee type
	ApproverEmployeeName string  `json:"approver_employee_name,omitempty"`
	StepMode          string     `json:"step_mode"`        // 'single' (default) or 'any' (parallel)
	ConditionField    *string    `json:"condition_field,omitempty"`
	ConditionOp       *string    `json:"condition_operator,omitempty"`
	ConditionValue    *float64   `json:"condition_value,omitempty"`
	EscalationHours   int        `json:"escalation_hours"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateApprovalWorkflowStepReq struct {
	WorkflowID        string   `json:"workflow_id"`
	StepOrder         int      `json:"step_order"`
	ApproverType      string   `json:"approver_type"`
	ApproverRoleID    *string  `json:"approver_role_id,omitempty"`
	ApproverEmployeeID *string `json:"approver_employee_id,omitempty"`
	StepMode          string   `json:"step_mode,omitempty"`
	ConditionField    *string  `json:"condition_field,omitempty"`
	ConditionOp       *string  `json:"condition_operator,omitempty"`
	ConditionValue    *float64 `json:"condition_value,omitempty"`
	EscalationHours   int      `json:"escalation_hours"`
}

type UpdateApprovalWorkflowStepReq struct {
	StepOrder         int      `json:"step_order"`
	ApproverType      string   `json:"approver_type"`
	ApproverRoleID    *string  `json:"approver_role_id,omitempty"`
	ApproverEmployeeID *string `json:"approver_employee_id,omitempty"`
	StepMode          string   `json:"step_mode,omitempty"`
	ConditionField    *string  `json:"condition_field,omitempty"`
	ConditionOp       *string  `json:"condition_operator,omitempty"`
	ConditionValue    *float64 `json:"condition_value,omitempty"`
	EscalationHours   int      `json:"escalation_hours"`
}

// ─── Approval Request Tracking ─────────────────────────────────

type ApprovalRequestTracking struct {
	ID           uuid.UUID `json:"id"`
	EntityType   string    `json:"entity_type"`
	EntityID     uuid.UUID `json:"entity_id"`
	WorkflowID   uuid.UUID `json:"workflow_id"`
	CurrentStep  int       `json:"current_step"`
	TotalSteps   int       `json:"total_steps"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ─── Approval Processing ───────────────────────────────────────

type ApprovalActionReq struct {
	Action string `json:"action"` // 'approve' or 'reject'
	Note   string `json:"note,omitempty"`
}

type ApprovalResult struct {
	Action         string `json:"action"`          // 'approved', 'rejected', 'pending_next_level'
	CurrentStep    int    `json:"current_step"`
	TotalSteps     int    `json:"total_steps"`
	NextApprover   string `json:"next_approver,omitempty"`
	Message        string `json:"message"`
	FinalStatus    string `json:"final_status,omitempty"` // 'approved' or 'rejected' if final
}

// ─── Pending Approvals ─────────────────────────────────────────

type PendingApprovalItem struct {
	TrackingID   uuid.UUID `json:"tracking_id"`
	EntityType   string    `json:"entity_type"`
	EntityID     uuid.UUID `json:"entity_id"`
	RequestorName string   `json:"requestor_name"`
	RequestorID  uuid.UUID `json:"requestor_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Amount       float64   `json:"amount,omitempty"`
	CurrentStep  int       `json:"current_step"`
	TotalSteps   int       `json:"total_steps"`
	ApproverType string    `json:"approver_type"`
	CreatedAt    time.Time `json:"created_at"`
	ElapsedHours float64   `json:"elapsed_hours"`
}

type PendingApprovalListResponse struct {
	Items []PendingApprovalItem `json:"items"`
	Total int                   `json:"total"`
}

// ─── Workflow Detail (with steps) ──────────────────────────────

type WorkflowDetailResponse struct {
	Workflow ApprovalWorkflow      `json:"workflow"`
	Steps    []ApprovalWorkflowStep `json:"steps"`
}
