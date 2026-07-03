package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	FullName         string     `json:"full_name"`
	BaseSalary       *float64   `json:"base_salary"`
	DailyWage        *float64   `json:"daily_wage"`
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
	WorkScheduleID   *uuid.UUID `json:"work_schedule_id"`
	WorkScheduleName string     `json:"work_schedule_name,omitempty"`
	Phone            string     `json:"phone"`
	PhotoURL         string     `json:"photo_url"`
	Address          string     `json:"address"`
	NIK              string     `json:"nik"`
	NPWP             string     `json:"npwp"`
	BankName         string     `json:"bank_name"`
	BankAccount      string     `json:"bank_account"`
	AddressKTP       string     `json:"address_ktp"`
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
	ID               uuid.UUID  `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	FullName         string     `json:"full_name"`
	BaseSalary       float64    `json:"base_salary"`
	Email            string     `json:"email"`
	Gender           string     `json:"gender"`
	EmploymentStatus string     `json:"employment_status"`
	IsActive         bool       `json:"is_active"`
	RoleName         string     `json:"role_name"`
	PositionName     string     `json:"position_name"`
	DepartmentName   string     `json:"department_name"`
	WorkScheduleName string     `json:"work_schedule_name,omitempty"`
	JoinDate         time.Time  `json:"join_date"`
	Phone            string     `json:"phone"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type EmployeeWorkScheduleOverride struct {
	WorkScheduleID string `json:"work_schedule_id"`
}

type DashboardResponse struct {
	TotalEmployees     int                   `json:"total_employees"`
	ActiveEmployees    int                   `json:"active_employees"`
	PresentToday       int                   `json:"present_today"`
	AttendanceRate     float64               `json:"attendance_rate"`
	PendingApprovals   int                   `json:"pending_approvals"`
	PayrollThisMonth   string                `json:"payroll_this_month"`
	AttendanceByDay    []AttendanceDay       `json:"attendance_by_day"`
	MonthlyTrend       []MonthlyTrend        `json:"monthly_trend"`
	Composition        []EmployeeComposition `json:"composition"`
	GenderBreakdown    []GenderBreakdown     `json:"gender_breakdown"`
	DepartmentStats    []DepartmentStat      `json:"department_stats"`
	RecentEmployees    []EmployeeSummary     `json:"recent_employees"`
}

type MonthlyTrend struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type GenderBreakdown struct {
	Gender string `json:"gender"`
	Count  int    `json:"count"`
}

type DepartmentStat struct {
	DepartmentName string `json:"department_name"`
	EmployeeCount  int    `json:"employee_count"`
}

type ManagerDashboardResponse struct {
	TeamSize        int                   `json:"team_size"`
	ActiveTeam      int                   `json:"active_team"`
	PendingApprovals int                  `json:"pending_approvals"`
	Composition      []EmployeeComposition `json:"composition"`
	RecentMembers    []EmployeeSummary     `json:"recent_members"`
	AttendanceToday  int                   `json:"attendance_today"`
}

type HRDashboardResponse struct {
	TotalEmployees    int                   `json:"total_employees"`
	ActiveEmployees   int                   `json:"active_employees"`
	DepartmentCount   int                   `json:"department_count"`
	HiringThisMonth   int                   `json:"hiring_this_month"`
	Composition       []EmployeeComposition `json:"composition"`
	GenderBreakdown   []GenderBreakdown     `json:"gender_breakdown"`
	DepartmentStats   []DepartmentStat      `json:"department_stats"`
	BirthdaysThisMonth []EmployeeSummary    `json:"birthdays_this_month"`
	ContractExpiring  []EmployeeSummary     `json:"contract_expiring"`
	RecentEmployees   []EmployeeSummary     `json:"recent_employees"`
}

type AttendanceDay struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type EmployeeComposition struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type CreateEmployeeRequest struct {
	EmployeeID       string   `json:"employee_id"`
	FullName         string   `json:"full_name"`
	Email            string   `json:"email"`
	Password         string   `json:"password"`
	BaseSalary       *float64 `json:"base_salary"`
	DailyWage        *float64 `json:"daily_wage"`
	Gender           string `json:"gender"`
	PlaceOfBirth     string `json:"place_of_birth"`
	DateOfBirth      string `json:"date_of_birth"`
	Religion         string `json:"religion"`
	MaritalStatus    string `json:"marital_status"`
	JoinDate         string `json:"join_date"`
	EmploymentStatus string `json:"employment_status"`
	Phone            string `json:"phone"`
	Address          string `json:"address"`
	NIK              string `json:"nik"`
	NPWP             string `json:"npwp"`
	BankName         string `json:"bank_name"`
	BankAccount      string `json:"bank_account"`
	AddressKTP       string `json:"address_ktp"`
	RoleID           string `json:"role_id"`
	PositionID       string `json:"position_id"`
	DepartmentID     string `json:"department_id"`
}

type EmployeeHistory struct {
	ID            string         `json:"id"`
	EmployeeID    string         `json:"employee_id"`
	EmployeeName  string         `json:"employee_name,omitempty"`
	ChangeType    string         `json:"change_type"`
	OldValue      map[string]any `json:"old_value,omitempty"`
	NewValue      map[string]any `json:"new_value,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	ChangedBy     string         `json:"changed_by,omitempty"`
	ChangedByName string         `json:"changed_by_name,omitempty"`
	ChangedAt     time.Time      `json:"changed_at"`
}

type EmployeeHistoryListResponse struct {
	Histories []EmployeeHistory `json:"histories"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PerPage   int               `json:"per_page"`
}

type ImportResult struct {
	Success int      `json:"success"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message"`
}

type UpdateEmployeeRequest struct {
	FullName         string   `json:"full_name"`
	Email            string   `json:"email"`
	Gender           string   `json:"gender"`
	PlaceOfBirth     string   `json:"place_of_birth"`
	DateOfBirth      string   `json:"date_of_birth"`
	Religion         string   `json:"religion"`
	MaritalStatus    string   `json:"marital_status"`
	JoinDate         string   `json:"join_date"`
	EmploymentStatus string   `json:"employment_status"`
	IsActive         *bool    `json:"is_active"`
	BaseSalary       *float64 `json:"base_salary"`
	DailyWage        *float64 `json:"daily_wage"`
	Phone            string   `json:"phone"`
	Address          string   `json:"address"`
	NIK              string   `json:"nik"`
	NPWP             string   `json:"npwp"`
	BankName         string   `json:"bank_name"`
	BankAccount      string   `json:"bank_account"`
	AddressKTP       string   `json:"address_ktp"`
	RoleID           string   `json:"role_id"`
	PositionID       string   `json:"position_id"`
	DepartmentID     string   `json:"department_id"`
	WorkScheduleID   string   `json:"work_schedule_id"`
}
