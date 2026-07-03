package models

import (
	"time"

	"github.com/google/uuid"
)

type SalaryComponent struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id"`
	ComponentName string     `json:"component_name"`
	ComponentType string     `json:"component_type"` // allowance / deduction
	Amount        float64    `json:"amount"`
	IsActive      bool       `json:"is_active"`
	EffectiveDate time.Time  `json:"effective_date"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type SalaryComponentSummary struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	ComponentName string    `json:"component_name"`
	ComponentType string    `json:"component_type"`
	Amount        float64   `json:"amount"`
	IsActive      bool      `json:"is_active"`
	EffectiveDate time.Time `json:"effective_date"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateSalaryComponentRequest struct {
	ComponentName string  `json:"component_name"`
	ComponentType string  `json:"component_type"`
	Amount        float64 `json:"amount"`
	EffectiveDate string  `json:"effective_date"`
}

type UpdateSalaryComponentRequest struct {
	ComponentName *string  `json:"component_name"`
	ComponentType *string  `json:"component_type"`
	Amount        *float64 `json:"amount"`
	IsActive      *bool    `json:"is_active"`
	EffectiveDate *string  `json:"effective_date"`
}

type SalaryComponentListResponse struct {
	Components []SalaryComponentSummary `json:"components"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PerPage    int                      `json:"per_page"`
}
