package models

import (
	"time"

	"github.com/google/uuid"
)

type Loan struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	EmployeeName      string     `json:"employee_name"`
	LoanType          string     `json:"loan_type"`
	Amount            float64    `json:"amount"`
	InterestRate      float64    `json:"interest_rate"`
	TotalInterest     float64    `json:"total_interest"`
	TotalAmount       float64    `json:"total_amount"`
	InstallmentCount  int        `json:"installment_count"`
	InstallmentAmount float64    `json:"installment_amount"`
	PaymentMethod     string     `json:"payment_method"`
	RemainingBalance  float64    `json:"remaining_balance"`
	Purpose           string     `json:"purpose"`
	ApprovalTrail     string     `json:"approval_trail"`
	Status            string     `json:"status"`
	DisbursedAt       *time.Time `json:"disbursed_at"`
	DisbursedBy       *uuid.UUID `json:"disbursed_by"`
	SettledAt         *time.Time `json:"settled_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

type LoanSummary struct {
	ID                uuid.UUID `json:"id"`
	EmployeeID        uuid.UUID `json:"employee_id"`
	EmployeeName      string    `json:"employee_name"`
	LoanType          string    `json:"loan_type"`
	Amount            float64   `json:"amount"`
	InstallmentCount  int       `json:"installment_count"`
	InstallmentAmount float64   `json:"installment_amount"`
	RemainingBalance  float64   `json:"remaining_balance"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateLoanRequest struct {
	EmployeeID       string  `json:"employee_id"`
	LoanType         string  `json:"loan_type"`
	Amount           float64 `json:"amount"`
	InterestRate     float64 `json:"interest_rate"`
	InstallmentCount int     `json:"installment_count"`
	PaymentMethod    string  `json:"payment_method"`
	Purpose          string  `json:"purpose"`
}

type UpdateLoanStatusRequest struct {
	RejectionReason string `json:"rejection_reason"`
	DisburseDate    string `json:"disburse_date"`
}

type LoanListResponse struct {
	Loans   []LoanSummary `json:"loans"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
}

type PaymentMethodStat struct {
	Method string  `json:"method"`
	Count  int     `json:"count"`
	Total  float64 `json:"total"`
}

type LoanStatsResponse struct {
	TotalLoans       int                 `json:"total_loans"`
	ActiveLoans      int                 `json:"active_loans"`
	TotalOutstanding float64             `json:"total_outstanding"`
	TotalDisbursed   float64             `json:"total_disbursed"`
	ByStatus         map[string]int      `json:"by_status"`
	ByMethod         []PaymentMethodStat `json:"by_method"`
}
