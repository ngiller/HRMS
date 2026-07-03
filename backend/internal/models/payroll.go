package models

import (
	"time"

	"github.com/google/uuid"
)

type PayrollPeriodStatus string

const (
	PayrollStatusDraft      PayrollPeriodStatus = "draft"
	PayrollStatusCalculated PayrollPeriodStatus = "calculated"
	PayrollStatusApproved   PayrollPeriodStatus = "approved"
	PayrollStatusPaid       PayrollPeriodStatus = "paid"
)

// PayrollPeriod represents a payroll period (e.g., Jan 2026)
type PayrollPeriod struct {
	ID          uuid.UUID            `json:"id"`
	Month       int                  `json:"month"`
	Year        int                  `json:"year"`
	PeriodName  string               `json:"period_name"`
	StartDate   time.Time            `json:"start_date"`
	EndDate     time.Time            `json:"end_date"`
	Status      PayrollPeriodStatus  `json:"status"`
	ApprovedBy  *uuid.UUID           `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time           `json:"approved_at,omitempty"`
	PaidBy      *uuid.UUID           `json:"paid_by,omitempty"`
	PaidAt      *time.Time           `json:"paid_at,omitempty"`

	// Summary
	TotalEmployee    int     `json:"total_employee"`
	TotalGross       float64 `json:"total_gross"`
	TotalDeductions  float64 `json:"total_deductions"`
	TotalNet         float64 `json:"total_net"`
	TotalCompanyCost float64 `json:"total_company_cost"`

	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PayrollPeriodSummary is a lighter version for list views
type PayrollPeriodSummary struct {
	ID              uuid.UUID           `json:"id"`
	Month           int                 `json:"month"`
	Year            int                 `json:"year"`
	PeriodName      string              `json:"period_name"`
	StartDate       time.Time           `json:"start_date"`
	EndDate         time.Time           `json:"end_date"`
	Status          PayrollPeriodStatus `json:"status"`
	TotalEmployee   int                 `json:"total_employee"`
	TotalGross      float64             `json:"total_gross"`
	TotalNet        float64             `json:"total_net"`
	ApprovedByName  string              `json:"approved_by_name,omitempty"`
	PaidByName      string              `json:"paid_by_name,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
}

// PayrollItem represents a single employee's payroll for a period
type PayrollItem struct {
	ID                uuid.UUID `json:"id"`
	PayrollPeriodID   uuid.UUID `json:"payroll_period_id"`
	EmployeeID        uuid.UUID `json:"employee_id"`

	// Income
	BaseSalary       float64 `json:"base_salary"`
	DailyWage        float64 `json:"daily_wage"`
	TotalDaysWorked  int     `json:"total_days_worked"`
	RawAllowances    []byte  `json:"-"` // raw JSONB
	Allowances       []AllowanceItem `json:"allowances"`
	OvertimePay      float64 `json:"overtime_pay"`
	OvertimeHours    float64 `json:"overtime_hours"`
	THRAmount        float64 `json:"thr_amount"`
	BonusAmount      float64 `json:"bonus_amount"`
	GrossSalary      float64 `json:"gross_salary"`

	// Deductions
	RawDeductions    []byte  `json:"-"`
	Deductions       []AllowanceItem `json:"deductions"`
	PPh21Amount      float64 `json:"pph21_amount"`
	BPJSKesehatan    float64 `json:"bpjs_kesehatan"`
	BPJSJHT          float64 `json:"bpjs_jht"`
	BPJSJP           float64 `json:"bpjs_jp"`
	LoanDeduction    float64 `json:"loan_deduction"`
	OtherDeductions  float64 `json:"other_deductions"`
	TotalDeductions  float64 `json:"total_deductions"`

	// Net
	NetSalary        float64 `json:"net_salary"`

	// Company costs
	CompanyCost             []AllowanceItem `json:"company_cost"`
	BPJSKesehatanCompany    float64 `json:"bpjs_kesehatan_company"`
	BPJSJHTCompany          float64 `json:"bpjs_jht_company"`
	BPJSJPCompany           float64 `json:"bpjs_jp_company"`
	BPJSJKK                 float64 `json:"bpjs_jkk"`
	BPJSJKM                 float64 `json:"bpjs_jkm"`

	// Employee info (joined)
	EmployeeName     string  `json:"employee_name,omitempty"`
	EmployeeIDCode   string  `json:"employee_id_code,omitempty"`
	DepartmentName   string  `json:"department_name,omitempty"`
	PositionName     string  `json:"position_name,omitempty"`
	EmploymentStatus string  `json:"employment_status,omitempty"`

	Status    PayrollPeriodStatus `json:"status"`
	Notes     string              `json:"notes,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type AllowanceItem struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

// Request types
type CreatePayrollPeriodRequest struct {
	PeriodName string `json:"period_name"`
	Month      int    `json:"month"`
	Year       int    `json:"year"`
	StartDate  string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate    string `json:"end_date,omitempty"`   // YYYY-MM-DD
}

type CalculatePayrollRequest struct {
	EmployeeID      string  `json:"employee_id,omitempty"`       // empty = all employees
	BaseSalary      float64 `json:"base_salary,omitempty"`
	DailyWage       float64 `json:"daily_wage,omitempty"`
	TotalDaysWorked int     `json:"total_days_worked,omitempty"`
	OvertimePay     float64 `json:"overtime_pay,omitempty"`
	THRAmount       float64 `json:"thr_amount,omitempty"`
	BonusAmount     float64 `json:"bonus_amount,omitempty"`
	LoanDeduction   float64 `json:"loan_deduction,omitempty"`
	OtherDeductions float64 `json:"other_deductions,omitempty"`
}

type PayrollPeriodListResponse struct {
	Periods []PayrollPeriodSummary `json:"periods"`
	Total   int                    `json:"total"`
	Page    int                    `json:"page"`
	PerPage int                    `json:"per_page"`
}

type PayrollItemsListResponse struct {
	Items   []PayrollItem `json:"items"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
}

type PayslipResponse struct {
	PayrollItem
	PeriodName    string `json:"period_name"`
	PeriodStatus  string `json:"period_status"`
}
