package models

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceRecord struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	Date                 time.Time  `json:"date"`
	CheckInTime          *time.Time `json:"check_in_time"`
	CheckInPhotoURL      *string    `json:"check_in_photo_url"`
	CheckInLat           *float64   `json:"check_in_lat"`
	CheckInLng           *float64   `json:"check_in_lng"`
	CheckInLocationID    *uuid.UUID `json:"check_in_location_id"`
	CheckInLocationName  *string    `json:"check_in_location_name"`
	CheckOutTime         *time.Time `json:"check_out_time"`
	CheckOutPhotoURL     *string    `json:"check_out_photo_url"`
	CheckOutLat          *float64   `json:"check_out_lat"`
	CheckOutLng          *float64   `json:"check_out_lng"`
	CheckOutLocationID   *uuid.UUID `json:"check_out_location_id"`
	CheckOutLocationName *string    `json:"check_out_location_name"`
	Photo        *string  `json:"photo"`
	Status               string     `json:"status"`
	IsLate               bool       `json:"is_late"`
	LateMinutes          int        `json:"late_minutes"`
	IsEarlyLeave         bool       `json:"is_early_leave"`
	EarlyLeaveMinutes    int        `json:"early_leave_minutes"`
	TotalWorkHours       *float64   `json:"total_work_hours"`
	IsManualEntry        bool       `json:"is_manual_entry"`
	ManualEntryReason    *string    `json:"manual_entry_reason"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

type AttendanceRecordSummary struct {
	ID                   uuid.UUID  `json:"id"`
	Date                 time.Time  `json:"date"`
	DayName              string     `json:"day_name"`
	CheckInTime          *time.Time `json:"check_in_time"`
	CheckOutTime         *time.Time `json:"check_out_time"`
	Status               string     `json:"status"`
	IsLate               bool       `json:"is_late"`
	LateMinutes          int        `json:"late_minutes"`
	TotalWorkHours       *float64   `json:"total_work_hours"`
	CheckInLocationName  *string    `json:"check_in_location_name"`
	CheckInPhotoURL      *string    `json:"check_in_photo_url"`
	CheckInLat           *float64   `json:"check_in_lat"`
	CheckInLng           *float64   `json:"check_in_lng"`
	CheckOutLocationName *string    `json:"check_out_location_name"`
	Photo        *string  `json:"photo"`
	CheckOutPhotoURL     *string    `json:"check_out_photo_url"`
	CheckOutLat          *float64   `json:"check_out_lat"`
	CheckOutLng          *float64   `json:"check_out_lng"`
	EmployeeName         string     `json:"employee_name,omitempty"`
	DepartmentName       string     `json:"department_name,omitempty"`
}

type CheckInRequest struct {
	Lat          *float64 `json:"lat"`
	Lng          *float64 `json:"lng"`
	LocationID   *string  `json:"location_id"`
	LocationName *string  `json:"location_name"`
	Photo        *string  `json:"photo"`
}

type CheckOutRequest struct {
	Lat          *float64 `json:"lat"`
	Lng          *float64 `json:"lng"`
	LocationID   *string  `json:"location_id"`
	LocationName *string  `json:"location_name"`
	Photo        *string  `json:"photo"`
}

type TodayAttendanceStatus struct {
	HasCheckedIn  bool                     `json:"has_checked_in"`
	HasCheckedOut bool                     `json:"has_checked_out"`
	Record        *AttendanceRecordSummary `json:"record,omitempty"`
	ScheduleName  string                   `json:"schedule_name"`
	ScheduleStart string                   `json:"schedule_start"`
	ScheduleEnd   string                   `json:"schedule_end"`
}

type AttendanceListResponse struct {
	Records []AttendanceRecordSummary `json:"records"`
	Total   int                       `json:"total"`
	Page    int                       `json:"page"`
	PerPage int                       `json:"per_page"`
}

type AttendanceReportFilter struct {
	EmployeeID   string `json:"employee_id"`
	DepartmentID string `json:"department_id"`
	Status       string `json:"status"`
	DateFrom     string `json:"date_from"`
	DateTo       string `json:"date_to"`
}
