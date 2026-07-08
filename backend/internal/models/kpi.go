package models

import (
	"time"

	"github.com/google/uuid"
)

type KPITemplate struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	PositionID    *uuid.UUID `json:"position_id"`
	PositionName  string     `json:"position_name"`
	DepartmentID  *uuid.UUID `json:"department_id"`
	DeptName      string     `json:"dept_name"`
	PeriodType    string     `json:"period_type"`
	Year          int        `json:"year"`
	Description   string     `json:"description"`
	IsActive      bool       `json:"is_active"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedByName string     `json:"created_by_name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type KPIIndicator struct {
	ID              uuid.UUID `json:"id"`
	KPITemplateID   uuid.UUID `json:"kpi_template_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Target          float64   `json:"target"`
	Weight          float64   `json:"weight"`
	MeasurementUnit string    `json:"measurement_unit"`
	SortOrder       int       `json:"sort_order"`
}

type KPITemplateDetail struct {
	KPITemplate
	Indicators []KPIIndicator `json:"indicators"`
}

type KPIReview struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	EmployeeName      string     `json:"employee_name"`
	KPITemplateID     uuid.UUID  `json:"kpi_template_id"`
	TemplateTitle     string     `json:"template_title"`
	Period            string     `json:"period"`
	Year              int        `json:"year"`
	SelfRating        string     `json:"self_rating"`
	SelfScore         *float64   `json:"self_score"`
	SelfNote          string     `json:"self_note"`
	SelfSubmittedAt   *time.Time `json:"self_submitted_at"`
	ManagerRating     string     `json:"manager_rating"`
	ManagerScore      *float64   `json:"manager_score"`
	ManagerNote       string     `json:"manager_note"`
	ManagerID         *uuid.UUID `json:"manager_id"`
	ManagerName       string     `json:"manager_name"`
	ManagerReviewedAt *time.Time `json:"manager_reviewed_at"`
	HRRating          string     `json:"hr_rating"`
	FinalScore        *float64   `json:"final_score"`
	FinalCategory     string     `json:"final_category"`
	HRNote            string     `json:"hr_note"`
	HRID              *uuid.UUID `json:"hr_id"`
	HRName            string     `json:"hr_name"`
	HRReviewedAt      *time.Time `json:"hr_reviewed_at"`
	Status            string     `json:"status"`
	SalaryIncrease    *float64   `json:"salary_increase"`
	BonusAmount       *float64   `json:"bonus_amount"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type KPIReviewSummary struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	TemplateTitle string    `json:"template_title"`
	Period        string    `json:"period"`
	Year          int       `json:"year"`
	FinalScore    *float64  `json:"final_score"`
	FinalCategory string    `json:"final_category"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateKPIReviewRequest struct {
	EmployeeID    string `json:"employee_id"`
	KPITemplateID string `json:"kpi_template_id"`
	Period        string `json:"period"`
	Year          int    `json:"year"`
}

type KPIReviewListResponse struct {
	Reviews []KPIReviewSummary `json:"reviews"`
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
}

type KPITemplateListResponse struct {
	Templates []KPITemplate `json:"templates"`
	Total     int           `json:"total"`
	Page      int           `json:"page"`
	PerPage   int           `json:"per_page"`
}

type CreateKPITemplateRequest struct {
	Title        string               `json:"title"`
	PositionID   *string              `json:"position_id"`
	DepartmentID *string              `json:"department_id"`
	PeriodType   string               `json:"period_type"`
	Year         int                  `json:"year"`
	Description  string               `json:"description"`
	Indicators   []CreateKPIIndicator `json:"indicators"`
}

type UpdateKPITemplateRequest struct {
	Title        string               `json:"title"`
	PositionID   *string              `json:"position_id"`
	DepartmentID *string              `json:"department_id"`
	PeriodType   string               `json:"period_type"`
	Year         int                  `json:"year"`
	Description  string               `json:"description"`
	IsActive     *bool                `json:"is_active"`
	Indicators   []CreateKPIIndicator `json:"indicators"`
}

type CreateKPIIndicator struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Target          float64 `json:"target"`
	Weight          float64 `json:"weight"`
	MeasurementUnit string  `json:"measurement_unit"`
	SortOrder       int     `json:"sort_order"`
}
