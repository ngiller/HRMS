package models

import (
	"time"

	"github.com/google/uuid"
)

type OvertimeRequest struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	EmployeeName    string     `json:"employee_name"`
	Date            string     `json:"date"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	TotalHours      float64    `json:"total_hours"`
	OvertimeType    string     `json:"overtime_type"`
	Reason          string     `json:"reason"`
	IsMandatory     bool       `json:"is_mandatory"`
	ApprovalTrail   string     `json:"approval_trail"`
	Status          string     `json:"status"`
	RejectionReason string     `json:"rejection_reason"`
	CancelledBy     *uuid.UUID `json:"cancelled_by"`
	CancelledAt     *time.Time `json:"cancelled_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`

	// Calculation result (from overtime_calculation view)
	HourlyRate  float64 `json:"hourly_rate,omitempty"`
	OvertimePay float64 `json:"overtime_pay,omitempty"`
}

type OvertimeRequestSummary struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Date         string    `json:"date"`
	TotalHours   float64   `json:"total_hours"`
	OvertimeType string    `json:"overtime_type"`
	Reason       string    `json:"reason"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateOvertimeRequestReq struct {
	Date         string  `json:"date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	TotalHours   float64 `json:"total_hours"`
	OvertimeType string  `json:"overtime_type"`
	Reason       string  `json:"reason"`
	IsMandatory  bool    `json:"is_mandatory"`
}

type UpdateOvertimeStatusReq struct {
	RejectionReason string `json:"rejection_reason"`
}

type OvertimeRequestListResponse struct {
	OvertimeRequests []OvertimeRequestSummary `json:"overtime_requests"`
	Total            int                      `json:"total"`
	Page             int                      `json:"page"`
	PerPage          int                      `json:"per_page"`
}

type OvertimeCalculationResponse struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	Date         string    `json:"date"`
	TotalHours   float64   `json:"total_hours"`
	OvertimeType string    `json:"overtime_type"`
	BaseSalary   float64   `json:"base_salary"`
	HourlyRate   float64   `json:"hourly_rate"`
	RateSegments string    `json:"rate_segments"`
	OvertimePay  float64   `json:"overtime_pay"`
}
