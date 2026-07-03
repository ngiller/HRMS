-- +goose NO TRANSACTION
-- +goose Up
-- ============================================================
-- Migration 00038: Cleanup payroll_status ENUM
-- Remove unused 'completed' value, keep only active statuses
-- ============================================================
-- Note: PostgreSQL cannot remove enum values directly.
-- We recreate the type without 'completed'.
--
-- The 'completed' value from original enum is never used by
-- code (which uses 'calculated' instead via migration 00035).

-- Step 0: Drop dependent views and defaults that reference the payroll_status type
DROP VIEW IF EXISTS payslip_view CASCADE;

ALTER TABLE payroll_periods ALTER COLUMN status DROP DEFAULT;
ALTER TABLE payroll_items ALTER COLUMN status DROP DEFAULT;

-- Step 1: Alter tables that use payroll_status to use text temporarily
ALTER TABLE payroll_periods ALTER COLUMN status TYPE text;
ALTER TABLE payroll_items ALTER COLUMN status TYPE text;

-- Step 2: Drop old enum (columns already converted to text, so safe)
DROP TYPE IF EXISTS payroll_status;

-- Step 3: Recreate enum without 'completed'
CREATE TYPE payroll_status AS ENUM (
    'draft',
    'calculated',
    'approved',
    'paid'
);

-- Step 4: Cast columns back to new enum
ALTER TABLE payroll_periods ALTER COLUMN status TYPE payroll_status USING status::payroll_status;
ALTER TABLE payroll_items ALTER COLUMN status TYPE payroll_status USING status::payroll_status;

-- Step 5: Restore default values
ALTER TABLE payroll_periods ALTER COLUMN status SET DEFAULT 'draft'::payroll_status;
ALTER TABLE payroll_items ALTER COLUMN status SET DEFAULT 'draft'::payroll_status;

-- Step 6: Recreate payslip_view
CREATE OR REPLACE VIEW payslip_view AS
SELECT
    pi.id AS payroll_item_id,
    pp.month,
    pp.year,
    e.id AS employee_id,
    e.employee_id AS nip,
    e.full_name,
    d.name AS department_name,
    pos.name AS position_name,
    pi.base_salary,
    pi.daily_wage,
    pi.allowances,
    pi.overtime_pay,
    pi.thr_amount,
    pi.bonus_amount,
    pi.gross_salary,
    pi.pph21_amount,
    pi.bpjs_kesehatan,
    pi.bpjs_jht,
    pi.bpjs_jp,
    pi.loan_deduction,
    pi.total_deductions,
    pi.net_salary,
    pi.company_cost,
    pi.bpjs_kesehatan_company,
    pi.bpjs_jht_company,
    pi.bpjs_jp_company,
    pi.bpjs_jkk,
    pi.bpjs_jkm,
    pi.status,
    pp.status AS period_status
FROM payroll_items pi
JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
JOIN employees e ON e.id = pi.employee_id
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN positions pos ON pos.id = e.position_id
WHERE e.deleted_at IS NULL;

-- +goose Down
-- Drop payslip_view before altering types
DROP VIEW IF EXISTS payslip_view CASCADE;

-- Drop default values first
ALTER TABLE payroll_periods ALTER COLUMN status DROP DEFAULT;
ALTER TABLE payroll_items ALTER COLUMN status DROP DEFAULT;

ALTER TABLE payroll_periods ALTER COLUMN status TYPE text;
ALTER TABLE payroll_items ALTER COLUMN status TYPE text;

DROP TYPE IF EXISTS payroll_status;

CREATE TYPE payroll_status AS ENUM (
    'draft',
    'completed',
    'approved',
    'paid'
);

ALTER TABLE payroll_periods ALTER COLUMN status TYPE payroll_status USING status::payroll_status;
ALTER TABLE payroll_items ALTER COLUMN status TYPE payroll_status USING status::payroll_status;

ALTER TABLE payroll_periods ALTER COLUMN status SET DEFAULT 'draft'::payroll_status;
ALTER TABLE payroll_items ALTER COLUMN status SET DEFAULT 'draft'::payroll_status;

-- Recreate payslip_view
CREATE OR REPLACE VIEW payslip_view AS
SELECT
    pi.id AS payroll_item_id,
    pp.month,
    pp.year,
    e.id AS employee_id,
    e.employee_id AS nip,
    e.full_name,
    d.name AS department_name,
    pos.name AS position_name,
    pi.base_salary,
    pi.daily_wage,
    pi.allowances,
    pi.overtime_pay,
    pi.thr_amount,
    pi.bonus_amount,
    pi.gross_salary,
    pi.pph21_amount,
    pi.bpjs_kesehatan,
    pi.bpjs_jht,
    pi.bpjs_jp,
    pi.loan_deduction,
    pi.total_deductions,
    pi.net_salary,
    pi.company_cost,
    pi.bpjs_kesehatan_company,
    pi.bpjs_jht_company,
    pi.bpjs_jp_company,
    pi.bpjs_jkk,
    pi.bpjs_jkm,
    pi.status,
    pp.status AS period_status
FROM payroll_items pi
JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
JOIN employees e ON e.id = pi.employee_id
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN positions pos ON pos.id = e.position_id
WHERE e.deleted_at IS NULL;
