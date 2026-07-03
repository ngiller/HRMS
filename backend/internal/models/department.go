package models

import (
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Code            string     `json:"code"`
	ParentID        *uuid.UUID `json:"parent_id"`
	HeadID          *uuid.UUID `json:"head_id"`
	WorkScheduleID  *uuid.UUID `json:"work_schedule_id"`
	Description     string     `json:"description"`
	IsActive        bool       `json:"is_active"`
	SortOrder       int        `json:"sort_order"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// DepartmentSummary — digunakan di list response (tanpa nested fields berat)
type DepartmentSummary struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	ParentName        string    `json:"parent_name,omitempty"`
	HeadName          string    `json:"head_name,omitempty"`
	WorkScheduleName  string    `json:"work_schedule_name,omitempty"`
	Description       string    `json:"description"`
	IsActive          bool      `json:"is_active"`
	SortOrder         int       `json:"sort_order"`
	EmployeeCnt       int       `json:"employee_count"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateDepartmentRequest — payload untuk create department
type CreateDepartmentRequest struct {
	Name           string  `json:"name" validate:"required"`
	Code           string  `json:"code" validate:"required"`
	ParentID       *string `json:"parent_id"`
	HeadID         *string `json:"head_id"`
	WorkScheduleID *string `json:"work_schedule_id"`
	Description    string  `json:"description"`
	SortOrder      int     `json:"sort_order"`
}

// UpdateDepartmentRequest — payload untuk update department
type UpdateDepartmentRequest struct {
	Name           *string `json:"name"`
	Code           *string `json:"code"`
	ParentID       *string `json:"parent_id"`
	HeadID         *string `json:"head_id"`
	WorkScheduleID *string `json:"work_schedule_id"`
	Description    *string `json:"description"`
	IsActive       *bool   `json:"is_active"`
	SortOrder      *int    `json:"sort_order"`
}

// DepartmentListResponse — paginated response
type DepartmentListResponse struct {
	Departments []DepartmentSummary `json:"departments"`
	Total       int                 `json:"total"`
	Page        int                 `json:"page"`
	PerPage     int                 `json:"per_page"`
}

// AssignScheduleRequest — payload untuk assign jadwal kerja ke departemen
type AssignScheduleRequest struct {
	WorkScheduleID string `json:"work_schedule_id"`
}
