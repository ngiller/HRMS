package models

import (
	"time"
)

// ActivityLog represents an audit trail entry
type ActivityLog struct {
	ID         string          `json:"id"`
	UserID     *string         `json:"user_id,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	EntityName string          `json:"entity_name,omitempty"`
	OldValues  *map[string]any `json:"old_values,omitempty"`
	NewValues  *map[string]any `json:"new_values,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	// Joined fields
	EmployeeName *string `json:"employee_name,omitempty"`
	EmployeeEmail *string `json:"employee_email,omitempty"`
}

// ActivityLogSummary is a lightweight log entry for listing
type ActivityLogSummary struct {
	ID           string    `json:"id"`
	UserID       *string   `json:"user_id,omitempty"`
	Action       string    `json:"action"`
	EntityType   string    `json:"entity_type"`
	EntityName   string    `json:"entity_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	EmployeeName *string   `json:"employee_name,omitempty"`
}

// ActivityLogListResponse wraps paginated activity logs
type ActivityLogListResponse struct {
	Logs  []ActivityLogSummary `json:"logs"`
	Total int                  `json:"total"`
	Page  int                  `json:"page"`
	PerPage int                `json:"per_page"`
}

// ActivityLogFilter holds filter parameters for activity logs
type ActivityLogFilter struct {
	Action     string `json:"action,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
}

// AuditAction constants
const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionApprove = "approve"
	AuditActionReject  = "reject"
	AuditActionPay     = "pay"
	AuditActionCancel  = "cancel"
	AuditActionLogin   = "login"
	AuditActionLogout  = "logout"
)
