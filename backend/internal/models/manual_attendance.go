package models

import (
	"time"

	"github.com/google/uuid"
)

// ManualAttendanceRequest represents a request to manually set attendance
type ManualAttendanceRequest struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	EmployeeName    string     `json:"employee_name,omitempty"`
	Date            time.Time  `json:"date"`
	CheckInTime     *time.Time `json:"check_in_time"`
	CheckOutTime    *time.Time `json:"check_out_time"`
	Reason          string     `json:"reason"`
	Status          string     `json:"status"`
	ApprovedBy      *uuid.UUID `json:"approved_by"`
	ApprovedByName  string     `json:"approved_by_name"`
	ApprovedAt      *time.Time `json:"approved_at"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateManualAttendanceRequest is the request body
type CreateManualAttendanceRequest struct {
	Date         string  `json:"date"`
	CheckInTime  string  `json:"check_in_time"`
	CheckOutTime string  `json:"check_out_time"`
	Reason       string  `json:"reason"`
}

// ManualAttendanceListResponse wraps paginated manual attendance requests
type ManualAttendanceListResponse struct {
	Requests []ManualAttendanceRequest `json:"requests"`
	Total    int                       `json:"total"`
	Page     int                       `json:"page"`
	PerPage  int                       `json:"per_page"`
}
