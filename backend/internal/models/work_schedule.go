package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkSchedule struct {
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	ScheduleType         string     `json:"schedule_type"`
	Description          string     `json:"description"`
	MondayStart          string     `json:"monday_start"`
	MondayEnd            string     `json:"monday_end"`
	TuesdayStart         string     `json:"tuesday_start"`
	TuesdayEnd           string     `json:"tuesday_end"`
	WednesdayStart       string     `json:"wednesday_start"`
	WednesdayEnd         string     `json:"wednesday_end"`
	ThursdayStart        string     `json:"thursday_start"`
	ThursdayEnd          string     `json:"thursday_end"`
	FridayStart          string     `json:"friday_start"`
	FridayEnd            string     `json:"friday_end"`
	SaturdayStart        *string    `json:"saturday_start"`
	SaturdayEnd          *string    `json:"saturday_end"`
	SundayStart          *string    `json:"sunday_start"`
	SundayEnd            *string    `json:"sunday_end"`
	BreakStart           string     `json:"break_start"`
	BreakEnd             string     `json:"break_end"`
	LateToleranceMinutes int        `json:"late_tolerance_minutes"`
	EarlyLeaveTolerance  int        `json:"early_leave_tolerance"`
	WeeklyHours          float64    `json:"weekly_hours"`
	IsActive             bool       `json:"is_active"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

type WorkScheduleSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ScheduleType string    `json:"schedule_type"`
	Description  string    `json:"description"`
	WeeklyHours  float64   `json:"weekly_hours"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateWorkScheduleRequest struct {
	Name                 string  `json:"name" validate:"required"`
	ScheduleType         string  `json:"schedule_type" validate:"required"`
	Description          string  `json:"description"`
	MondayStart          string  `json:"monday_start"`
	MondayEnd            string  `json:"monday_end"`
	TuesdayStart         string  `json:"tuesday_start"`
	TuesdayEnd           string  `json:"tuesday_end"`
	WednesdayStart       string  `json:"wednesday_start"`
	WednesdayEnd         string  `json:"wednesday_end"`
	ThursdayStart        string  `json:"thursday_start"`
	ThursdayEnd          string  `json:"thursday_end"`
	FridayStart          string  `json:"friday_start"`
	FridayEnd            string  `json:"friday_end"`
	SaturdayStart        *string `json:"saturday_start"`
	SaturdayEnd          *string `json:"saturday_end"`
	SundayStart          *string `json:"sunday_start"`
	SundayEnd            *string `json:"sunday_end"`
	BreakStart           string  `json:"break_start"`
	BreakEnd             string  `json:"break_end"`
	LateToleranceMinutes int     `json:"late_tolerance_minutes"`
	EarlyLeaveTolerance  int     `json:"early_leave_tolerance"`
	WeeklyHours          float64 `json:"weekly_hours"`
}

type UpdateWorkScheduleRequest struct {
	Name                 *string  `json:"name"`
	ScheduleType         *string  `json:"schedule_type"`
	Description          *string  `json:"description"`
	MondayStart          *string  `json:"monday_start"`
	MondayEnd            *string  `json:"monday_end"`
	TuesdayStart         *string  `json:"tuesday_start"`
	TuesdayEnd           *string  `json:"tuesday_end"`
	WednesdayStart       *string  `json:"wednesday_start"`
	WednesdayEnd         *string  `json:"wednesday_end"`
	ThursdayStart        *string  `json:"thursday_start"`
	ThursdayEnd          *string  `json:"thursday_end"`
	FridayStart          *string  `json:"friday_start"`
	FridayEnd            *string  `json:"friday_end"`
	SaturdayStart        *string  `json:"saturday_start"`
	SaturdayEnd          *string  `json:"saturday_end"`
	SundayStart          *string  `json:"sunday_start"`
	SundayEnd            *string  `json:"sunday_end"`
	BreakStart           *string  `json:"break_start"`
	BreakEnd             *string  `json:"break_end"`
	LateToleranceMinutes *int     `json:"late_tolerance_minutes"`
	EarlyLeaveTolerance  *int     `json:"early_leave_tolerance"`
	WeeklyHours          *float64 `json:"weekly_hours"`
	IsActive             *bool    `json:"is_active"`
}

type WorkScheduleListResponse struct {
	WorkSchedules []WorkScheduleSummary `json:"work_schedules"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
}
