package models

import (
	"time"

	"github.com/google/uuid"
)

type Position struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	DepartmentID uuid.UUID  `json:"department_id"`
	GradeID      *uuid.UUID `json:"grade_id"`
	Description  string     `json:"description"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type PositionSummary struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	DepartmentName string    `json:"department_name"`
	DepartmentID   string    `json:"department_id"`
	GradeName      string    `json:"grade_name"`
	GradeID        string    `json:"grade_id"`
	Description    string    `json:"description"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreatePositionRequest struct {
	Name         string `json:"name" validate:"required"`
	DepartmentID string `json:"department_id" validate:"required"`
	GradeID      string `json:"grade_id"`
	Description  string `json:"description"`
}

type UpdatePositionRequest struct {
	Name         *string `json:"name"`
	DepartmentID *string `json:"department_id"`
	GradeID      *string `json:"grade_id"`
	Description  *string `json:"description"`
	IsActive     *bool   `json:"is_active"`
}

type PositionListResponse struct {
	Positions []PositionSummary `json:"positions"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PerPage   int               `json:"per_page"`
}
