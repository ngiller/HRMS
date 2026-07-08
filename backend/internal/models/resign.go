package models

import (
	"time"

	"github.com/google/uuid"
)

// ResignRequest represents an employee's resignation
type ResignRequest struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       uuid.UUID  `json:"employee_id"`
	EmployeeName     string     `json:"employee_name,omitempty"`
	ResignDate       time.Time  `json:"resign_date"`
	LastWorkingDate  time.Time  `json:"last_working_date"`
	Reason           string     `json:"reason"`
	ResignType       string     `json:"resign_type"` // voluntary, termination, retirement, etc.
	Status           string     `json:"status"`      // pending, approved, rejected, processed
	ApprovedBy       *uuid.UUID `json:"approved_by"`
	ApprovedByName   string     `json:"approved_by_name,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at"`
	RejectionReason  string     `json:"rejection_reason,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ExitClearanceItem represents a checklist item for exit clearance
type ExitClearanceItem struct {
	ID          uuid.UUID  `json:"id"`
	ResignID    uuid.UUID  `json:"resign_id"`
	ItemName    string     `json:"item_name"`
	Description string     `json:"description,omitempty"`
	IsChecked   bool       `json:"is_checked"`
	CheckedBy   *uuid.UUID `json:"checked_by"`
	CheckedByName string   `json:"checked_by_name,omitempty"`
	CheckedAt   *time.Time `json:"checked_at"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateResignRequest is the request to create a resignation
type CreateResignRequest struct {
	LastWorkingDate string `json:"last_working_date"`
	Reason          string `json:"reason"`
	ResignType      string `json:"resign_type"`
}

// ResignListResponse wraps paginated resignation requests
type ResignListResponse struct {
	Resigns []ResignRequest `json:"resigns"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
}

// ExitClearanceListResponse wraps clearance items
type ExitClearanceListResponse struct {
	Items []ExitClearanceItem `json:"items"`
}
