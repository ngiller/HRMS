package models

import (
	"time"

	"github.com/google/uuid"
)

// ─── Leave Type ──────────────────────────────────────────────

type LeaveType struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Code                string     `json:"code"`
	DefaultQuota        int        `json:"default_quota"`
	IsPaid              bool       `json:"is_paid"`
	RequiresDocument    bool       `json:"requires_document"`
	MaxConsecutiveDays  *int       `json:"max_consecutive_days"`
	CanRollover         bool       `json:"can_rollover"`
	RolloverMaxDays     int        `json:"rollover_max_days"`
	CanCashout          bool       `json:"can_cashout"`
	IsActive            bool       `json:"is_active"`
	SortOrder           int        `json:"sort_order"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
}

type LeaveTypeSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	DefaultQuota int       `json:"default_quota"`
	IsPaid       bool      `json:"is_paid"`
	IsActive     bool      `json:"is_active"`
}

// ─── Leave Balance ────────────────────────────────────────────

type LeaveBalance struct {
	ID             uuid.UUID `json:"id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	LeaveTypeID    uuid.UUID `json:"leave_type_id"`
	Year           int       `json:"year"`
	TotalQuota     int       `json:"total_quota"`
	Used           int       `json:"used"`
	Remaining      int       `json:"remaining"`
	RolledOverFrom int       `json:"rolled_over_from"`
	LeaveTypeName  string    `json:"leave_type_name,omitempty"`
	LeaveTypeCode  string    `json:"leave_type_code,omitempty"`
}

type LeaveBalanceResponse struct {
	Balances []LeaveBalance `json:"balances"`
	Total    int            `json:"total"`
}

// ─── Leave Request ────────────────────────────────────────────

type LeaveRequest struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	EmployeeName        string     `json:"employee_name,omitempty"`
	LeaveTypeID         uuid.UUID  `json:"leave_type_id"`
	LeaveTypeName       string     `json:"leave_type_name,omitempty"`
	StartDate           string     `json:"start_date"`
	EndDate             string     `json:"end_date"`
	TotalDays           int        `json:"total_days"`
	IsHalfDay           bool       `json:"is_half_day"`
	Reason              string     `json:"reason"`
	DocumentURL         string     `json:"document_url,omitempty"`
	ContactDuringLeave  string     `json:"contact_during_leave,omitempty"`
	ApprovalTrail       string     `json:"approval_trail"`
	Status              string     `json:"status"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CancelReason        string     `json:"cancel_reason,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
}

type LeaveRequestSummary struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	LeaveTypeName string    `json:"leave_type_name"`
	StartDate     string    `json:"start_date"`
	EndDate       string    `json:"end_date"`
	TotalDays     int       `json:"total_days"`
	IsHalfDay     bool      `json:"is_half_day"`
	Reason        string    `json:"reason"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateLeaveRequestReq struct {
	LeaveTypeID        string `json:"leave_type_id"`
	StartDate          string `json:"start_date"`
	EndDate            string `json:"end_date"`
	TotalDays          int    `json:"total_days"`
	IsHalfDay          bool   `json:"is_half_day"`
	Reason             string `json:"reason"`
	DocumentURL        string `json:"document_url"`
	ContactDuringLeave string `json:"contact_during_leave"`
}

type UpdateLeaveStatusReq struct {
	RejectionReason  string `json:"rejection_reason"`
	CancelReason     string `json:"cancel_reason"`
}

type LeaveRequestListResponse struct {
	LeaveRequests []LeaveRequestSummary `json:"leave_requests"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
}
