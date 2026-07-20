-- Migration 00060: Add department_id to shifts
-- Description: Allow shifts to belong to a specific department or be global

-- +goose Up
-- +goose StatementBegin

-- Add department_id column
ALTER TABLE shifts ADD COLUMN IF NOT EXISTS department_id UUID REFERENCES departments(id) ON DELETE CASCADE;

-- Create index on department_id
CREATE INDEX IF NOT EXISTS idx_shifts_department ON shifts(department_id) WHERE deleted_at IS NULL;

-- Recreate unique index on shifts to allow duplicate codes across different departments
-- but ensure unique code within the same department, and unique code for global shifts
DROP INDEX IF EXISTS idx_shifts_code;

-- Unique code within the department
CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_code_dept 
ON shifts(code, department_id) 
WHERE deleted_at IS NULL AND department_id IS NOT NULL;

-- Unique code for global shifts (department_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_code_global 
ON shifts(code) 
WHERE deleted_at IS NULL AND department_id IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_shifts_code_global;
DROP INDEX IF EXISTS idx_shifts_code_dept;
DROP INDEX IF EXISTS idx_shifts_department;
ALTER TABLE shifts DROP COLUMN IF EXISTS department_id;

-- Recreate original index
CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_code ON shifts(code) WHERE deleted_at IS NULL;
-- +goose StatementEnd
