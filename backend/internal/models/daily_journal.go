package models

import (
	"github.com/google/uuid"
	"time"
)

type DailyJournal struct {
	ID                 uuid.UUID  `json:"id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	EmployeeName       string     `json:"employee_name"`
	JournalDate        string     `json:"journal_date"`
	WorkDescription    string     `json:"work_description"`
	Achievements       string     `json:"achievements"`
	Challenges         string     `json:"challenges"`
	PlanTomorrow       string     `json:"plan_tomorrow"`
	Status             string     `json:"status"`
	SubmittedAt        *time.Time `json:"submitted_at"`
	AcknowledgedBy     *uuid.UUID `json:"acknowledged_by"`
	AcknowledgedByName string     `json:"acknowledged_by_name"`
	AcknowledgedAt     *time.Time `json:"acknowledged_at"`
	DepartmentID       *uuid.UUID `json:"department_id"`
	DepartmentName     string     `json:"department_name"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type DailyJournalSummary struct {
	ID              uuid.UUID `json:"id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	EmployeeName    string    `json:"employee_name"`
	JournalDate     string    `json:"journal_date"`
	WorkDescription string    `json:"work_description"`
	Status          string    `json:"status"`
	DepartmentName  string    `json:"department_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateDailyJournalRequest struct {
	JournalDate     string `json:"journal_date"`
	WorkDescription string `json:"work_description"`
	Achievements    string `json:"achievements"`
	Challenges      string `json:"challenges"`
	PlanTomorrow    string `json:"plan_tomorrow"`
}

type DailyJournalListResponse struct {
	Journals []DailyJournalSummary `json:"journals"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PerPage  int                   `json:"per_page"`
}
