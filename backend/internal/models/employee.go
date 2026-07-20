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
	ApprovalLineID   *uuid.UUID `json:"approval_line_id"`
	ApprovalLineName string     `json:"approval_line_name,omitempty"`
	Phone            string     `json:"phone"`
	PhotoURL         string     `json:"photo_url"`
	Address          string     `json:"address"`
	NIK              string     `json:"nik"`
	NPWP             string     `json:"npwp"`
	BankName         string     `json:"bank_name"`
	BankAccount      string     `json:"bank_account"`
	AddressKTP       string     `json:"address_ktp"`
	PTKPStatus       string     `json:"ptkp_status"`
	IsPregnant       bool       `json:"is_pregnant"`
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
	ManagerID        string     `json:"manager_id,omitempty"`
	ManagerName      string     `json:"manager_name,omitempty"`
}

type EmployeeWorkScheduleOverride struct {
	WorkScheduleID string `json:"work_schedule_id"`
}

type AbsentEmployee struct {
	EmployeeID     string `json:"employee_id"`
	FullName       string `json:"full_name"`
	DepartmentName string `json:"department_name"`
	AbsenceReason  string `json:"absence_reason"`
	LeaveReason    string `json:"leave_reason"`
}

type DashboardResponse struct {
	TotalEmployees   int                   `json:"total_employees"`
	ActiveEmployees  int                   `json:"active_employees"`
	PresentToday     int                   `json:"present_today"`
	AttendanceRate   float64               `json:"attendance_rate"`
	PendingApprovals int                   `json:"pending_approvals"`
	PayrollThisMonth string                `json:"payroll_this_month"`
	AttendanceByDay  []AttendanceDay       `json:"attendance_by_day"`
	MonthlyTrend     []MonthlyTrend        `json:"monthly_trend"`
	Composition      []EmployeeComposition `json:"composition"`
	GenderBreakdown  []GenderBreakdown     `json:"gender_breakdown"`
	DepartmentStats  []DepartmentStat      `json:"department_stats"`
	RecentEmployees  []EmployeeSummary     `json:"recent_employees"`
	AbsentToday      []AbsentEmployee      `json:"absent_today"`
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
	TeamSize         int                   `json:"team_size"`
	ActiveTeam       int                   `json:"active_team"`
	PendingApprovals int                   `json:"pending_approvals"`
	Composition      []EmployeeComposition `json:"composition"`
	RecentMembers    []EmployeeSummary     `json:"recent_members"`
	AttendanceToday  int                   `json:"attendance_today"`
}

type HRDashboardResponse struct {
	TotalEmployees     int                   `json:"total_employees"`
	ActiveEmployees    int                   `json:"active_employees"`
	DepartmentCount    int                   `json:"department_count"`
	HiringThisMonth    int                   `json:"hiring_this_month"`
	Composition        []EmployeeComposition `json:"composition"`
	GenderBreakdown    []GenderBreakdown     `json:"gender_breakdown"`
	DepartmentStats    []DepartmentStat      `json:"department_stats"`
	BirthdaysThisMonth []EmployeeSummary     `json:"birthdays_this_month"`
	ContractExpiring   []EmployeeSummary     `json:"contract_expiring"`
	RecentEmployees    []EmployeeSummary     `json:"recent_employees"`
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
	Gender           string   `json:"gender"`
	PlaceOfBirth     string   `json:"place_of_birth"`
	DateOfBirth      string   `json:"date_of_birth"`
	Religion         string   `json:"religion"`
	MaritalStatus    string   `json:"marital_status"`
	JoinDate         string   `json:"join_date"`
	EmploymentStatus string   `json:"employment_status"`
	IsPregnant       bool     `json:"is_pregnant"`
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
	BloodType        string   `json:"blood_type"`
	PTKPStatus       string   `json:"ptkp_status"`
	EndDate          string   `json:"end_date"`
	PositionGradeID  string   `json:"position_grade_id"`
	ApprovalLineID   string   `json:"approval_line_id"`
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

type EmployeeExportFull struct {
	EmployeeID              string `json:"employee_id"`
	FullName                string `json:"full_name"`
	Barcode                 string `json:"barcode"`
	Organization            string `json:"organization"`
	JobPosition             string `json:"job_position"`
	JobLevel                string `json:"job_level"`
	JoinDate                string `json:"join_date"`
	ResignDate              string `json:"resign_date"`
	StatusEmployee          string `json:"status_employee"`
	EndDate                 string `json:"end_date"`
	SignDate                string `json:"sign_date"`
	Email                   string `json:"email"`
	BirthDate               string `json:"birth_date"`
	Age                     string `json:"age"`
	BirthPlace              string `json:"birth_place"`
	CitizenIDAddress        string `json:"citizen_id_address"`
	ResidentialAddress      string `json:"residential_address"`
	NPWP                    string `json:"npwp"`
	PTKPStatus              string `json:"ptkp_status"`
	EmployeeTaxStatus       string `json:"employee_tax_status"`
	TaxConfig               string `json:"tax_config"`
	BankName                string `json:"bank_name"`
	BankAccount             string `json:"bank_account"`
	BankAccountHolder       string `json:"bank_account_holder"`
	BPJSTK                  string `json:"bpjs_tk"`
	BPJSKesehatan           string `json:"bpjs_kesehatan"`
	NIK                     string `json:"nik"`
	MobilePhone             string `json:"mobile_phone"`
	Phone                   string `json:"phone"`
	BranchName              string `json:"branch_name"`
	ParentBranchName        string `json:"parent_branch_name"`
	Religion                string `json:"religion"`
	Gender                  string `json:"gender"`
	MaritalStatus           string `json:"marital_status"`
	BloodType               string `json:"blood_type"`
	NationalityCode         string `json:"nationality_code"`
	Currency                string `json:"currency"`
	LengthOfService         string `json:"length_of_service"`
	PaymentSchedule         string `json:"payment_schedule"`
	ApprovalLine            string `json:"approval_line"`
	Manager                 string `json:"manager"`
	Grade                   string `json:"grade"`
	Class                   string `json:"class"`
	ProfilePicture          string `json:"profile_picture"`
	CostCenter              string `json:"cost_center"`
	CostCenterCategory      string `json:"cost_center_category"`
	SBU                     string `json:"sbu"`
	NPWPBaru                string `json:"npwp_baru"`
	Passport                string `json:"passport"`
	PassportExpirationDate  string `json:"passport_expiration_date"`
	JenisDokReferensi       string `json:"jenis_dok_referensi"`
	NomorDokReferensi       string `json:"nomor_dok_referensi"`
	TanggalDokReferensi     string `json:"tanggal_dok_referensi"`
	TIN                     string `json:"tin"`
	UkuranBaju              string `json:"ukuran_baju"`
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
	ApprovalLineID   *string  `json:"approval_line_id"`
	IsPregnant       *bool    `json:"is_pregnant"`
}
