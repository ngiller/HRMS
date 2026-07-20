package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// SHIFT DEFINITIONS
// ============================================================

type Shift struct {
	ID           uuid.UUID  `json:"id"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	StartTime    string     `json:"start_time"`
	EndTime      string     `json:"end_time"`
	BreakStart   string     `json:"break_start"`
	BreakEnd     string     `json:"break_end"`
	Color        string     `json:"color"`
	Description  string     `json:"description"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type ShiftSummary struct {
	ID           uuid.UUID  `json:"id"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	StartTime    string     `json:"start_time"`
	EndTime      string     `json:"end_time"`
	Color        string     `json:"color"`
	IsActive     bool       `json:"is_active"`
	Description  string     `json:"description"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateShiftRequest struct {
	DepartmentID string `json:"department_id,omitempty"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakStart   string `json:"break_start"`
	BreakEnd     string `json:"break_end"`
	Color        string `json:"color"`
	Description  string `json:"description"`
}

type UpdateShiftRequest struct {
	DepartmentID *string `json:"department_id,omitempty"`
	Name         *string `json:"name"`
	Code         *string `json:"code"`
	StartTime    *string `json:"start_time"`
	EndTime      *string `json:"end_time"`
	BreakStart   *string `json:"break_start"`
	BreakEnd     *string `json:"break_end"`
	Color        *string `json:"color"`
	Description  *string `json:"description"`
	IsActive     *bool   `json:"is_active"`
}

type ShiftListResponse struct {
	Shifts []ShiftSummary `json:"shifts"`
	Total  int            `json:"total"`
	Page   int            `json:"page"`
	PerPage int           `json:"per_page"`
}

// ============================================================
// DEPARTMENT ROSTERS
// ============================================================

type DepartmentRoster struct {
	ID           uuid.UUID  `json:"id"`
	DepartmentID uuid.UUID  `json:"department_id"`
	DepartmentName string  `json:"department_name,omitempty"`
	Name         string     `json:"name"`
	Month        int        `json:"month"`
	Year         int        `json:"year"`
	IsPublished  bool       `json:"is_published"`
	Notes        string     `json:"notes"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Entries      []RosterEntry `json:"entries,omitempty"`
}

type DepartmentRosterSummary struct {
	ID             uuid.UUID `json:"id"`
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	Name           string    `json:"name"`
	Month          int       `json:"month"`
	Year           int       `json:"year"`
	IsPublished    bool      `json:"is_published"`
	Notes          string    `json:"notes"`
	EntryCount     int       `json:"entry_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateDepartmentRosterRequest struct {
	DepartmentID string `json:"department_id"`
	Name         string `json:"name"`
	Month        int    `json:"month"`
	Year         int    `json:"year"`
	Notes        string `json:"notes"`
}

type UpdateDepartmentRosterRequest struct {
	Name        *string `json:"name"`
	Month       *int    `json:"month"`
	Year        *int    `json:"year"`
	IsPublished *bool   `json:"is_published"`
	Notes       *string `json:"notes"`
}

type DepartmentRosterListResponse struct {
	Rosters []DepartmentRosterSummary `json:"rosters"`
	Total   int                       `json:"total"`
	Page    int                       `json:"page"`
	PerPage int                       `json:"per_page"`
}

// ============================================================
// ROSTER ENTRIES
// ============================================================

type RosterEntry struct {
	ID           uuid.UUID `json:"id"`
	RosterID     uuid.UUID `json:"roster_id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name,omitempty"`
	Date         string    `json:"date"`
	ShiftID      uuid.UUID `json:"shift_id"`
	ShiftName    string    `json:"shift_name,omitempty"`
	ShiftCode    string    `json:"shift_code,omitempty"`
	ShiftColor   string    `json:"shift_color,omitempty"`
	StartTime    string    `json:"start_time,omitempty"`
	EndTime      string    `json:"end_time,omitempty"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateRosterEntryRequest struct {
	RosterID   string `json:"roster_id"`
	EmployeeID string `json:"employee_id"`
	Date       string `json:"date"`
	ShiftID    string `json:"shift_id"`
	Notes      string `json:"notes"`
}

type UpdateRosterEntryRequest struct {
	ShiftID *string `json:"shift_id"`
	Notes   *string `json:"notes"`
}

type BulkRosterEntryRequest struct {
	RosterID      string                     `json:"roster_id"`
	Entries       []CreateRosterEntryRequest `json:"entries"`
	ClearExisting bool                       `json:"clear_existing"`
}

// Roster calendar — employees grouped with their daily shifts for a month
type RosterCalendarEntry struct {
	EmployeeID   string                  `json:"employee_id"`
	EmployeeName string                  `json:"employee_name"`
	Days         map[string]RosterDayInfo `json:"days"` // key: "YYYY-MM-DD"
}

type RosterDayInfo struct {
	ShiftID    string `json:"shift_id,omitempty"`
	ShiftName  string `json:"shift_name,omitempty"`
	ShiftCode  string `json:"shift_code,omitempty"`
	ShiftColor string `json:"shift_color,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
	EntryID    string `json:"entry_id,omitempty"`
	Notes      string `json:"notes,omitempty"`
}
