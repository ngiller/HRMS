package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	FullName         string     `json:"full_name"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	Gender           string     `json:"gender"`
	PlaceOfBirth     string     `json:"place_of_birth"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	Religion         string     `json:"religion"`
	MaritalStatus    string     `json:"marital_status"`
	JoinDate         time.Time  `json:"join_date"`
	EmploymentStatus string     `json:"employment_status"`
	IsActive         bool       `json:"is_active"`
	RoleID           *uuid.UUID `json:"role_id"`
	RoleSlug         string     `json:"role_slug,omitempty"`
	RoleName         string     `json:"role_name,omitempty"`
	PositionID       *uuid.UUID `json:"position_id"`
	PositionName     string     `json:"position_name,omitempty"`
	DepartmentID     *uuid.UUID `json:"department_id"`
	DepartmentName   string     `json:"department_name,omitempty"`
	Phone            string     `json:"phone"`
	Address          string     `json:"address"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	IsLocked         bool       `json:"is_locked"`
	LockedUntil      *time.Time `json:"locked_until"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type EmployeeListResponse struct {
	Employees []EmployeeSummary `json:"employees"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PerPage   int               `json:"per_page"`
}

type EmployeeSummary struct {
	ID               uuid.UUID `json:"id"`
	EmployeeID       string    `json:"employee_id"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	Gender           string    `json:"gender"`
	EmploymentStatus string    `json:"employment_status"`
	IsActive         bool      `json:"is_active"`
	RoleName         string    `json:"role_name"`
	PositionName     string    `json:"position_name"`
	DepartmentName   string    `json:"department_name"`
	JoinDate         time.Time `json:"join_date"`
	Phone            string    `json:"phone"`
}

type DashboardResponse struct {
	TotalEmployees   int                   `json:"total_employees"`
	PresentToday     int                   `json:"present_today"`
	PendingApprovals int                   `json:"pending_approvals"`
	PayrollThisMonth string                `json:"payroll_this_month"`
	AttendanceByDay  []AttendanceDay       `json:"attendance_by_day"`
	Composition      []EmployeeComposition `json:"composition"`
	RecentEmployees  []EmployeeSummary     `json:"recent_employees"`
}

type AttendanceDay struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type EmployeeComposition struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
