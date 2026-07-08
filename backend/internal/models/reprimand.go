package models

import (
	"github.com/google/uuid"
	"time"
)

type Reprimand struct {
	ID                    uuid.UUID  `json:"id"`
	EmployeeID            uuid.UUID  `json:"employee_id"`
	EmployeeName          string     `json:"employee_name"`
	ReprimandType         string     `json:"reprimand_type"`
	Title                 string     `json:"title"`
	Description           string     `json:"description"`
	ViolationDate         *time.Time `json:"violation_date"`
	ViolationDetails      string     `json:"violation_details"`
	IssuedBy              *uuid.UUID `json:"issued_by"`
	IssuedByName          string     `json:"issued_by_name"`
	IssuedDate            time.Time  `json:"issued_date"`
	AcknowledgmentDate    *time.Time `json:"acknowledgment_date"`
	AcknowledgmentNote    string     `json:"acknowledgment_note"`
	DocumentURL           string     `json:"document_url"`
	EffectivePeriodMonths int        `json:"effective_period_months"`
	Status                string     `json:"status"`
	ExpiredAt             *time.Time `json:"expired_at"`
	EscalatedFromID       *uuid.UUID `json:"escalated_from_id"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type ReprimandSummary struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	ReprimandType string    `json:"reprimand_type"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	IssuedDate    time.Time `json:"issued_date"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateReprimandRequest struct {
	EmployeeID            string `json:"employee_id"`
	ReprimandType         string `json:"reprimand_type"`
	Title                 string `json:"title"`
	Description           string `json:"description"`
	ViolationDate         string `json:"violation_date"`
	ViolationDetails      string `json:"violation_details"`
	DocumentURL           string `json:"document_url"`
	EffectivePeriodMonths int    `json:"effective_period_months"`
}

type UpdateReprimandStatusRequest struct {
	AcknowledgmentNote string `json:"acknowledgment_note"`
}

type ReprimandListResponse struct {
	Reprimands []ReprimandSummary `json:"reprimands"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
}
