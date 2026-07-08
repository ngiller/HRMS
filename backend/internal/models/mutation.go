package models

import "time"

// EmployeeMutation tracks a mutation/promotion/transfer of an employee
type EmployeeMutation struct {
	ID              string     `json:"id"`
	EmployeeID      string     `json:"employee_id"`
	EmployeeName    string     `json:"employee_name,omitempty"`
	MutationType    string     `json:"mutation_type"`

	OldDepartmentID     *string `json:"old_department_id,omitempty"`
	OldDepartmentName   string  `json:"old_department_name,omitempty"`
	OldPositionID       *string `json:"old_position_id,omitempty"`
	OldPositionName     string  `json:"old_position_name,omitempty"`
	OldPositionGradeID  *string `json:"old_position_grade_id,omitempty"`
	OldPositionGradeName string `json:"old_position_grade_name,omitempty"`
	OldEmploymentStatus *string `json:"old_employment_status,omitempty"`
	OldBaseSalary       *float64 `json:"old_base_salary,omitempty"`

	NewDepartmentID     *string `json:"new_department_id,omitempty"`
	NewDepartmentName   string  `json:"new_department_name,omitempty"`
	NewPositionID       *string `json:"new_position_id,omitempty"`
	NewPositionName     string  `json:"new_position_name,omitempty"`
	NewPositionGradeID  *string `json:"new_position_grade_id,omitempty"`
	NewPositionGradeName string `json:"new_position_grade_name,omitempty"`
	NewEmploymentStatus *string `json:"new_employment_status,omitempty"`
	NewBaseSalary       *float64 `json:"new_base_salary,omitempty"`

	Reason          string     `json:"reason"`
	DocumentURL     string     `json:"document_url,omitempty"`
	EffectiveDate   string     `json:"effective_date"`
	Notes           string     `json:"notes,omitempty"`
	Status          string     `json:"status"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedByName  string     `json:"approved_by_name,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`

	CreatedBy      string    `json:"created_by"`
	CreatedByName  string    `json:"created_by_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateMutationRequest is the request to create a mutation request
type CreateMutationRequest struct {
	EmployeeID      string  `json:"employee_id"`
	MutationType    string  `json:"mutation_type"`
	NewDepartmentID string  `json:"new_department_id,omitempty"`
	NewPositionID   string  `json:"new_position_id,omitempty"`
	NewPositionGradeID string `json:"new_position_grade_id,omitempty"`
	NewEmploymentStatus string `json:"new_employment_status,omitempty"`
	NewBaseSalary   *float64 `json:"new_base_salary,omitempty"`
	Reason          string  `json:"reason"`
	DocumentURL     string  `json:"document_url,omitempty"`
	EffectiveDate   string  `json:"effective_date"`
	Notes           string  `json:"notes,omitempty"`
}

// MutationListResponse wraps paginated mutations
type MutationListResponse struct {
	Mutations []EmployeeMutation `json:"mutations"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PerPage   int                `json:"per_page"`
}
