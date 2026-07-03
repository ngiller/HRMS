package models

import (
	"time"

	"github.com/google/uuid"
)

type ShiftChangeRequest struct {
	ID                    uuid.UUID  `json:"id"`
	RequestType           string     `json:"request_type"`
	EmployeeID            uuid.UUID  `json:"employee_id"`
	EmployeeName          string     `json:"employee_name"`
	TargetDate            string     `json:"target_date"`
	CurrentScheduleID     *uuid.UUID `json:"current_schedule_id"`
	CurrentScheduleName   string     `json:"current_schedule_name"`
	RequestedScheduleID   uuid.UUID  `json:"requested_schedule_id"`
	RequestedScheduleName string     `json:"requested_schedule_name"`
	SwapPartnerID         *uuid.UUID `json:"swap_partner_id"`
	SwapPartnerName       string     `json:"swap_partner_name"`
	SwapPartnerDate       *string    `json:"swap_partner_date"`
	SwapPartnerScheduleID *uuid.UUID `json:"swap_partner_schedule_id"`
	Reason                string     `json:"reason"`
	SwapPartnerConfirmed  bool       `json:"swap_partner_confirmed"`
	SwapPartnerConfirmedAt *time.Time `json:"swap_partner_confirmed_at"`
	Status                string     `json:"status"`
	ApprovalTrail         string     `json:"approval_trail"`
	RejectionReason       string     `json:"rejection_reason"`
	CancelledBy           *uuid.UUID `json:"cancelled_by"`
	CancelledAt           *time.Time `json:"cancelled_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty"`
}

type ShiftChangeRequestSummary struct {
	ID                    uuid.UUID `json:"id"`
	RequestType           string    `json:"request_type"`
	EmployeeID            uuid.UUID `json:"employee_id"`
	EmployeeName          string    `json:"employee_name"`
	TargetDate            string    `json:"target_date"`
	CurrentScheduleName   string    `json:"current_schedule_name"`
	RequestedScheduleName string    `json:"requested_schedule_name"`
	SwapPartnerName       string    `json:"swap_partner_name"`
	Reason                string    `json:"reason"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
}

type CreateShiftChangeRequestReq struct {
	RequestType           string `json:"request_type"`
	TargetDate            string `json:"target_date"`
	CurrentScheduleID     string `json:"current_schedule_id"`
	RequestedScheduleID   string `json:"requested_schedule_id"`
	SwapPartnerID         string `json:"swap_partner_id"`
	SwapPartnerDate       string `json:"swap_partner_date"`
	SwapPartnerScheduleID string `json:"swap_partner_schedule_id"`
	Reason                string `json:"reason"`
}

type UpdateShiftChangeStatusReq struct {
	RejectionReason string `json:"rejection_reason"`
}

type ShiftChangeRequestListResponse struct {
	ShiftChangeRequests []ShiftChangeRequestSummary `json:"shift_change_requests"`
	Total              int                          `json:"total"`
	Page               int                          `json:"page"`
	PerPage            int                          `json:"per_page"`
}
