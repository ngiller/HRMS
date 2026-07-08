package models

import (
	"time"

	"github.com/google/uuid"
)

type PositionGrade struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	MinSalary   *float64  `json:"min_salary"`
	MaxSalary   *float64  `json:"max_salary"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PositionGradeSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	MinSalary   *float64  `json:"min_salary"`
	MaxSalary   *float64  `json:"max_salary"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreatePositionGradeRequest struct {
	Name        string   `json:"name" validate:"required"`
	Level       int      `json:"level" validate:"required"`
	MinSalary   *float64 `json:"min_salary"`
	MaxSalary   *float64 `json:"max_salary"`
	Description string   `json:"description"`
}

type UpdatePositionGradeRequest struct {
	Name        *string  `json:"name"`
	Level       *int     `json:"level"`
	MinSalary   *float64 `json:"min_salary"`
	MaxSalary   *float64 `json:"max_salary"`
	Description *string  `json:"description"`
	IsActive    *bool    `json:"is_active"`
}

type PositionGradeListResponse struct {
	PositionGrades []PositionGradeSummary `json:"position_grades"`
	Total          int                    `json:"total"`
	Page           int                    `json:"page"`
	PerPage        int                    `json:"per_page"`
}
