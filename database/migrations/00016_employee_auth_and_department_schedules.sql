-- +goose Up
-- ============================================================
-- Migration 00016: Employee Auth & Department Schedules
-- ============================================================
-- 1. Create roles table for RBAC
-- 2. Add auth fields to employees (password_hash, role_id, last_login_at, is_locked)
-- 3. Add default work_schedule_id to departments
-- ============================================================

-- ============================================================
-- 1. Roles Table (Role-Based Access Control)
-- ============================================================
CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100) NOT NULL UNIQUE,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    permissions     JSONB NOT NULL DEFAULT '{}'::jsonb,  -- {module: {create, read, update, delete, approve}}
    is_system_role  BOOLEAN DEFAULT FALSE,               -- System roles cannot be deleted
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- 2. Add auth & role columns to employees
-- ============================================================
ALTER TABLE employees
    ADD COLUMN password_hash     TEXT,
    ADD COLUMN role_id           UUID REFERENCES roles(id) ON DELETE RESTRICT,
    ADD COLUMN last_login_at     TIMESTAMPTZ,
    ADD COLUMN is_locked         BOOLEAN DEFAULT FALSE,
    ADD COLUMN locked_until      TIMESTAMPTZ;

-- Ensure all employees have a unique email before adding constraint
-- (Generate email from employee_id for any existing records with NULL/duplicate email)
UPDATE employees
SET email = LOWER(REPLACE(employee_id, ' ', '_')) || '@company.com'
WHERE email IS NULL OR email = '';

-- Handle duplicate emails by appending a suffix
UPDATE employees e
SET email = e.email || '-' || SUBSTRING(MD5(e.id::TEXT)::TEXT, 1, 6)
WHERE EXISTS (
    SELECT 1 FROM employees e2
    WHERE e2.id != e.id
    AND e2.email = e.email
    AND e.deleted_at IS NULL
);

-- Set email to NOT NULL
ALTER TABLE employees
    ALTER COLUMN email SET NOT NULL;

-- Add unique constraint on email for login
ALTER TABLE employees
    ADD CONSTRAINT employees_email_unique UNIQUE (email);

-- Index for login lookup
CREATE INDEX idx_employees_email ON employees(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_role_id ON employees(role_id);

-- ============================================================
-- 3. Add default work_schedule_id to departments
-- ============================================================
ALTER TABLE departments
    ADD COLUMN work_schedule_id UUID REFERENCES work_schedules(id) ON DELETE SET NULL;

CREATE INDEX idx_departments_work_schedule_id ON departments(work_schedule_id);

-- ============================================================
-- 4. Seed default system roles
-- ============================================================
INSERT INTO roles (name, slug, description, is_system_role, permissions) VALUES
    ('Super Admin', 'super_admin', 'Administrator sistem dengan akses penuh ke semua modul dan konfigurasi', TRUE, '{
        "employee": {"create": true, "read": true, "update": true, "delete": true},
        "department": {"create": true, "read": true, "update": true, "delete": true},
        "sensitive_data": {"read": true},
        "attendance": {"create": true, "read": true, "update": true, "delete": true},
        "payroll": {"create": true, "read": true, "update": true, "delete": true},
        "leave": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "reimbursement": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "overtime": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "loan": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "kpi": {"create": true, "read": true, "update": true, "delete": true},
        "reprimand": {"create": true, "read": true, "update": true, "delete": true},
        "payslip": {"read": true},
        "announcement": {"create": true, "read": true, "update": true, "delete": true},
        "document": {"create": true, "read": true, "update": true, "delete": true},
        "company_settings": {"create": true, "read": true, "update": true, "delete": true},
        "user_management": {"create": true, "read": true, "update": true, "delete": true},
        "report": {"create": true, "read": true, "update": true, "delete": true}
    }'::jsonb),

    ('HR Manager', 'hr_manager', 'Manajer HR yang mengelola seluruh operasional SDM', TRUE, '{
        "employee": {"create": true, "read": true, "update": true, "delete": true},
        "department": {"create": true, "read": true, "update": true, "delete": true},
        "sensitive_data": {"read": true},
        "attendance": {"create": true, "read": true, "update": true, "delete": true},
        "payroll": {"create": true, "read": true, "update": true},
        "leave": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "reimbursement": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "overtime": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "loan": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "kpi": {"create": true, "read": true, "update": true, "delete": true},
        "reprimand": {"create": true, "read": true, "update": true, "delete": true},
        "payslip": {"read": true},
        "announcement": {"create": true, "read": true, "update": true, "delete": true},
        "document": {"create": true, "read": true, "update": true, "delete": true},
        "company_settings": {"read": true},
        "user_management": {"read": true},
        "report": {"create": true, "read": true, "update": true, "delete": true}
    }'::jsonb),

    ('HR Staff', 'hr_staff', 'Staf HR yang menjalankan operasional harian SDM', TRUE, '{
        "employee": {"create": true, "read": true, "update": true, "delete": true},
        "department": {"create": true, "read": true, "update": true, "delete": true},
        "sensitive_data": {"read": true},
        "attendance": {"create": true, "read": true, "update": true, "delete": true},
        "payroll": {"update": true},
        "leave": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "reimbursement": {"create": true, "read": true, "update": true, "delete": true},
        "overtime": {"create": true, "read": true, "update": true, "delete": true},
        "loan": {"create": true, "read": true, "update": true, "delete": true},
        "kpi": {"update": true},
        "reprimand": {"update": true},
        "payslip": {"read": true},
        "announcement": {"create": true, "read": true, "update": true, "delete": true},
        "document": {"create": true, "read": true, "update": true, "delete": true},
        "report": {"read": true}
    }'::jsonb),

    ('Finance', 'finance', 'Tim finance yang mengelola penggajian, pinjaman, dan reimbursement', TRUE, '{
        "employee": {"read": true},
        "department": {"read": true},
        "sensitive_data": {"read": true},
        "attendance": {"read": true},
        "payroll": {"create": true, "read": true, "update": true, "delete": true},
        "leave": {"read": true},
        "reimbursement": {"approve": true},
        "overtime": {"read": true},
        "loan": {"create": true, "read": true, "update": true, "delete": true, "approve": true},
        "payslip": {"read": true},
        "announcement": {"create": true, "read": true, "update": true, "delete": true},
        "document": {"read": true},
        "company_settings": {"read": true},
        "report": {"read": true}
    }'::jsonb),

    ('Manager', 'manager', 'Manajer / atasan yang menyetujui permintaan bawahan', TRUE, '{
        "employee": {"read": true},
        "department": {"read": true},
        "attendance": {"read": true},
        "leave": {"approve": true},
        "reimbursement": {"approve": true},
        "overtime": {"approve": true},
        "loan": {"approve": true},
        "kpi": {"create": true, "read": true, "update": true},
        "reprimand": {"read": true},
        "payslip": {"read": true},
        "announcement": {"create": true, "read": true, "update": true, "delete": true},
        "document": {"read": true},
        "report": {"read": true}
    }'::jsonb),

    ('Karyawan', 'employee', 'Karyawan reguler dengan akses self-service', TRUE, '{
        "employee": {"read": true},
        "department": {"read": true},
        "attendance": {"read": true},
        "leave": {"create": true, "read": true},
        "reimbursement": {"create": true, "read": true},
        "overtime": {"create": true, "read": true},
        "loan": {"create": true, "read": true},
        "kpi": {"create": true, "read": true},
        "reprimand": {"read": true},
        "payslip": {"read": true},
        "announcement": {"read": true},
        "document": {"create": true, "read": true, "update": true}
    }'::jsonb),

    ('Direktur', 'director', 'Pimpinan perusahaan dengan akses dashboard & laporan', TRUE, '{
        "employee": {"read": true},
        "department": {"read": true},
        "sensitive_data": {"read": true},
        "attendance": {"read": true},
        "payroll": {"read": true},
        "leave": {"read": true},
        "reimbursement": {"read": true},
        "overtime": {"read": true},
        "loan": {"approve": true},
        "kpi": {"read": true},
        "reprimand": {"read": true},
        "payslip": {"read": true},
        "announcement": {"create": true, "read": true, "update": true, "delete": true},
        "document": {"read": true},
        "company_settings": {"read": true},
        "report": {"read": true}
    }'::jsonb);

-- ============================================================
-- 5. Create Super Admin employee (default admin account)
-- Note: password harus di-reset setelah first login
-- ============================================================
-- Create a default company first if none exists
INSERT INTO companies (name, address, npwp)
SELECT 'Perusahaan Saya', 'Alamat Perusahaan', '00.000.000.0-000.000'
WHERE NOT EXISTS (SELECT 1 FROM companies LIMIT 1);

-- Disable audit trigger during seeding
ALTER TABLE employees DISABLE TRIGGER audit_employees;

-- Create default admin
INSERT INTO employees (
    employee_id,
    full_name,
    email,
    password_hash,
    gender,
    join_date,
    employment_status,
    is_active
)
SELECT
    'ADMIN-001',
    'Super Admin',
    'admin@company.com',
    'PLACEHOLDER_HASH_RESET_ON_FIRST_LOGIN',
    'laki_laki',
    CURRENT_DATE,
    'tetap',
    TRUE
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'ADMIN-001');

-- Assign Super Admin role to default admin
UPDATE employees
SET role_id = (SELECT id FROM roles WHERE slug = 'super_admin')
WHERE employee_id = 'ADMIN-001' AND role_id IS NULL;

-- ============================================================
-- 6. Update existing employees with default 'Karyawan' role
-- ============================================================
UPDATE employees
SET role_id = (SELECT id FROM roles WHERE slug = 'employee')
WHERE role_id IS NULL AND employee_id != 'ADMIN-001' AND is_active = TRUE;

ALTER TABLE employees ENABLE TRIGGER audit_employees;

-- +goose Down
-- ============================================================
-- Rollback migration 00016
-- ============================================================

-- Remove columns from departments
ALTER TABLE departments DROP COLUMN IF EXISTS work_schedule_id;
DROP INDEX IF EXISTS idx_departments_work_schedule_id;

-- Remove columns from employees
ALTER TABLE employees DROP COLUMN IF EXISTS password_hash;
ALTER TABLE employees DROP COLUMN IF EXISTS role_id;
ALTER TABLE employees DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE employees DROP COLUMN IF EXISTS is_locked;
ALTER TABLE employees DROP COLUMN IF EXISTS locked_until;

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_email_unique;

DROP INDEX IF EXISTS idx_employees_email;
DROP INDEX IF EXISTS idx_employees_role_id;

-- Delete default admin
ALTER TABLE employees DISABLE TRIGGER audit_employees;
DELETE FROM employees WHERE employee_id = 'ADMIN-001';
ALTER TABLE employees ENABLE TRIGGER audit_employees;

-- Drop roles table
DROP TABLE IF EXISTS roles;
