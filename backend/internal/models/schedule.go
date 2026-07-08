package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Schedule Template (Level 3 — reusable schedule patterns)
// ============================================================

type ScheduleTemplate struct {
	ID           uuid.UUID             `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	ScheduleType string                `json:"schedule_type"` // weekly, shift, flexible
	IsActive     bool                  `json:"is_active"`
	Days         []ScheduleTemplateDay `json:"days,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	DeletedAt    *time.Time            `json:"deleted_at,omitempty"`
}

type ScheduleTemplateSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ScheduleType string    `json:"schedule_type"`
	IsActive     bool      `json:"is_active"`
	DayCount     int       `json:"day_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type ScheduleTemplateDay struct {
	ID                   uuid.UUID `json:"id"`
	TemplateID           uuid.UUID `json:"template_id"`
	DayOfWeek            *int      `json:"day_of_week"` // 0=Senin..6=Minggu, NULL=all days
	StartTime            string    `json:"start_time"`
	EndTime              string    `json:"end_time"`
	BreakStart           string    `json:"break_start"`
	BreakEnd             string    `json:"break_end"`
	LateToleranceMinutes int       `json:"late_tolerance_minutes"`
	EarlyLeaveTolerance  int       `json:"early_leave_tolerance"`
	IsActive             bool      `json:"is_active"`
	SortOrder            int       `json:"sort_order"`
}

type CreateScheduleTemplateRequest struct {
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	ScheduleType string                      `json:"schedule_type"`
	Days         []CreateScheduleTemplateDay `json:"days"`
}

type CreateScheduleTemplateDay struct {
	DayOfWeek            *int   `json:"day_of_week"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	BreakStart           string `json:"break_start"`
	BreakEnd             string `json:"break_end"`
	LateToleranceMinutes int    `json:"late_tolerance_minutes"`
	EarlyLeaveTolerance  int    `json:"early_leave_tolerance"`
}

type UpdateScheduleTemplateRequest struct {
	Name         *string                     `json:"name"`
	Description  *string                     `json:"description"`
	ScheduleType *string                     `json:"schedule_type"`
	IsActive     *bool                       `json:"is_active"`
	Days         []CreateScheduleTemplateDay `json:"days"`
}

// ============================================================
// Employee Schedule (Level 2 + Level 3 — actual assignments)
// ============================================================

type EmployeeSchedule struct {
	ID             uuid.UUID                  `json:"id"`
	EmployeeID     uuid.UUID                  `json:"employee_id"`
	TemplateID     *uuid.UUID                 `json:"template_id,omitempty"`
	TemplateName   string                     `json:"template_name,omitempty"`
	DayOfWeek      *int                       `json:"day_of_week"`
	SpecificDate   *string                    `json:"specific_date"`
	StartTime      *string                    `json:"start_time"`
	EndTime        *string                    `json:"end_time"`
	BreakStart     *string                    `json:"break_start"`
	BreakEnd       *string                    `json:"break_end"`
	IsRemote       bool                       `json:"is_remote"`
	EffectiveFrom  string                     `json:"effective_from"`
	EffectiveUntil *string                    `json:"effective_until"`
	Priority       int                        `json:"priority"`
	Reason         string                     `json:"reason"`
	IsActive       bool                       `json:"is_active"`
	Locations      []EmployeeScheduleLocation `json:"locations,omitempty"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

type EmployeeScheduleSummary struct {
	ID             uuid.UUID `json:"id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	TemplateName   string    `json:"template_name"`
	DayOfWeek      *int      `json:"day_of_week"`
	SpecificDate   *string   `json:"specific_date"`
	StartTime      *string   `json:"start_time"`
	EndTime        *string   `json:"end_time"`
	IsRemote       bool      `json:"is_remote"`
	EffectiveFrom  string    `json:"effective_from"`
	EffectiveUntil *string   `json:"effective_until"`
	Priority       int       `json:"priority"`
	IsActive       bool      `json:"is_active"`
	LocationNames  string    `json:"location_names"`
	CreatedAt      time.Time `json:"created_at"`
}

type EmployeeScheduleLocation struct {
	ID                   uuid.UUID `json:"id"`
	EmployeeScheduleID   uuid.UUID `json:"employee_schedule_id"`
	AttendanceLocationID uuid.UUID `json:"attendance_location_id"`
	LocationName         string    `json:"location_name"`
	DayOfWeek            *int      `json:"day_of_week"`
	IsPrimary            bool      `json:"is_primary"`
	SortOrder            int       `json:"sort_order"`
}

type CreateEmployeeScheduleRequest struct {
	EmployeeID     string                   `json:"employee_id"`
	TemplateID     string                   `json:"template_id,omitempty"`
	DayOfWeek      *int                     `json:"day_of_week"`
	SpecificDate   *string                  `json:"specific_date"`
	StartTime      *string                  `json:"start_time"`
	EndTime        *string                  `json:"end_time"`
	BreakStart     *string                  `json:"break_start"`
	BreakEnd       *string                  `json:"break_end"`
	IsRemote       *bool                    `json:"is_remote"`
	EffectiveFrom  string                   `json:"effective_from"`
	EffectiveUntil *string                  `json:"effective_until"`
	Priority       *int                     `json:"priority"`
	Reason         string                   `json:"reason"`
	Locations      []CreateScheduleLocation `json:"locations"`
}

type CreateScheduleLocation struct {
	AttendanceLocationID string `json:"attendance_location_id"`
	DayOfWeek            *int   `json:"day_of_week"`
	IsPrimary            *bool  `json:"is_primary"`
}

type UpdateEmployeeScheduleRequest struct {
	TemplateID     *string                  `json:"template_id"`
	StartTime      *string                  `json:"start_time"`
	EndTime        *string                  `json:"end_time"`
	BreakStart     *string                  `json:"break_start"`
	BreakEnd       *string                  `json:"break_end"`
	IsRemote       *bool                    `json:"is_remote"`
	EffectiveFrom  *string                  `json:"effective_from"`
	EffectiveUntil *string                  `json:"effective_until"`
	Priority       *int                     `json:"priority"`
	Reason         *string                  `json:"reason"`
	IsActive       *bool                    `json:"is_active"`
	Locations      []CreateScheduleLocation `json:"locations"`
}

// ============================================================
// Resolved Schedule (hasil lookup untuk tanggal tertentu)
// ============================================================

type ResolvedSchedule struct {
	ScheduleID uuid.UUID         `json:"schedule_id"`
	Source     string            `json:"source"` // date_override, weekly_schedule, work_schedule, department_schedule
	StartTime  string            `json:"start_time"`
	EndTime    string            `json:"end_time"`
	BreakStart string            `json:"break_start"`
	BreakEnd   string            `json:"break_end"`
	IsRemote   bool              `json:"is_remote"`
	Location   *ResolvedLocation `json:"location,omitempty"`
}

type ResolvedLocation struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
}

// ============================================================
// List Responses
// ============================================================

type ScheduleTemplateListResponse struct {
	Templates []ScheduleTemplateSummary `json:"templates"`
	Total     int                       `json:"total"`
	Page      int                       `json:"page"`
	PerPage   int                       `json:"per_page"`
}

type EmployeeScheduleListResponse struct {
	Schedules []EmployeeScheduleSummary `json:"schedules"`
	Total     int                       `json:"total"`
	Page      int                       `json:"page"`
	PerPage   int                       `json:"per_page"`
}
