package models

import (
	"time"

	"github.com/google/uuid"
)

// Role — full role object from database
type Role struct {
	ID           uuid.UUID                  `json:"id"`
	Name         string                     `json:"name"`
	Slug         string                     `json:"slug"`
	Description  string                     `json:"description"`
	Permissions  map[string]map[string]bool `json:"permissions"`
	IsSystemRole bool                       `json:"is_system_role"`
	IsActive     bool                       `json:"is_active"`
	CreatedAt    time.Time                  `json:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

// RoleSummary — for list views
type RoleSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	IsSystemRole bool      `json:"is_system_role"`
	IsActive     bool      `json:"is_active"`
	EmployeeCnt  int       `json:"employee_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateRoleRequest — payload for creating a role
type CreateRoleRequest struct {
	Name        string                     `json:"name" validate:"required"`
	Slug        string                     `json:"slug" validate:"required"`
	Description string                     `json:"description"`
	Permissions map[string]map[string]bool `json:"permissions"`
}

// UpdateRoleRequest — payload for updating a role
type UpdateRoleRequest struct {
	Name        *string                     `json:"name"`
	Slug        *string                     `json:"slug"`
	Description *string                     `json:"description"`
	Permissions *map[string]map[string]bool `json:"permissions"`
	IsActive    *bool                       `json:"is_active"`
}

// RoleListResponse — paginated response
type RoleListResponse struct {
	Roles   []RoleSummary `json:"roles"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
}

// PermissionTemplate — struktural petunjuk modul & action yang tersedia
type PermissionModule struct {
	Module  string   `json:"module"`
	Label   string   `json:"label"`
	Actions []string `json:"actions"`
}

// DefaultPermissionTemplate returns the known modules and their available actions.
func DefaultPermissionTemplate() []PermissionModule {
	return []PermissionModule{
		{Module: "employee", Label: "Karyawan", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "department", Label: "Departemen", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "sensitive_data", Label: "Data Sensitif", Actions: []string{"read"}},
		{Module: "attendance", Label: "Absensi", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "payroll", Label: "Payroll", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "leave", Label: "Cuti", Actions: []string{"create", "read", "update", "delete", "approve"}},
		{Module: "reimbursement", Label: "Reimbursement", Actions: []string{"create", "read", "update", "delete", "approve"}},
		{Module: "overtime", Label: "Lembur", Actions: []string{"create", "read", "update", "delete", "approve"}},
		{Module: "loan", Label: "Pinjaman", Actions: []string{"create", "read", "update", "delete", "approve"}},
		{Module: "kpi", Label: "KPI", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "reprimand", Label: "Reprimand", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "payslip", Label: "Slip Gaji", Actions: []string{"read"}},
		{Module: "announcement", Label: "Pengumuman", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "document", Label: "Dokumen", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "company_settings", Label: "Pengaturan Perusahaan", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "user_management", Label: "Manajemen Pengguna", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "report", Label: "Laporan", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "position_grade", Label: "Golongan Jabatan", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "position", Label: "Posisi Jabatan", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "work_schedule", Label: "Jadwal Kerja", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "attendance_location", Label: "Lokasi Absensi", Actions: []string{"create", "read", "update", "delete"}},
		{Module: "shift_change", Label: "Permintaan Shift", Actions: []string{"create", "read", "update", "delete"}},
	}
}
