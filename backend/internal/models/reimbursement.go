package models

import (
	"time"

	"github.com/google/uuid"
)

type Reimbursement struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id"`
	EmployeeName  string     `json:"employee_name"`
	Type          string     `json:"type"`
	Amount        float64    `json:"amount"`
	Description   string     `json:"description"`
	ReceiptUrls   []string   `json:"receipt_urls"`
	ApprovalTrail string     `json:"approval_trail"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	PaidAt        *time.Time `json:"paid_at"`
	PaidBy        *uuid.UUID `json:"paid_by"`
	PaidByName    string     `json:"paid_by_name"`
	RejectionReason string   `json:"rejection_reason"`
	CancelledBy   *uuid.UUID `json:"cancelled_by"`
	CancelledAt   *time.Time `json:"cancelled_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type ReimbursementSummary struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Type         string    `json:"type"`
	Amount       float64   `json:"amount"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateReimbursementReq struct {
	Type        string   `json:"type"`
	Amount      float64  `json:"amount"`
	Description string   `json:"description"`
	ReceiptUrls []string `json:"receipt_urls"`
}

type UpdateReimbursementStatusReq struct {
	RejectionReason string `json:"rejection_reason"`
}

type PayReimbursementReq struct {
	PaymentMethod string `json:"payment_method"`
}

type ReimbursementListResponse struct {
	Reimbursements []ReimbursementSummary `json:"reimbursements"`
	Total          int                    `json:"total"`
	Page           int                    `json:"page"`
	PerPage        int                    `json:"per_page"`
}
