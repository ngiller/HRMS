-- +goose Up
-- ============================================================
-- Migration 00017: Employee Salary Components & History
-- ============================================================
-- Adds master data for per-employee salary components
-- (allowances & deductions) with change history tracking.
-- ============================================================

-- ============================================================
-- 1. Create salary_component_type enum
-- ============================================================
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE salary_component_type AS ENUM ('allowance', 'deduction');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- ============================================================
-- 2. Employee Salary Components
-- Master data: setiap karyawan punya daftar komponen gaji sendiri
-- ============================================================
CREATE TABLE employee_salary_components (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    component_name      VARCHAR(100) NOT NULL,               -- Tunj. Jabatan, Tunj. Transport, dll
    component_type      salary_component_type NOT NULL,      -- allowance / deduction
    amount              DECIMAL(15,2) NOT NULL DEFAULT 0,
    is_active           BOOLEAN DEFAULT TRUE,
    effective_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    
    -- Audit
    created_by          UUID REFERENCES employees(id) ON DELETE SET NULL,
    updated_by          UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE TRIGGER set_employee_salary_components_updated_at
    BEFORE UPDATE ON employee_salary_components
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- 3. Employee Salary Component Histories
-- Riwayat perubahan komponen gaji
-- ============================================================
CREATE TABLE employee_salary_component_histories (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    component_id        UUID REFERENCES employee_salary_components(id) ON DELETE SET NULL,
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    component_name      VARCHAR(100) NOT NULL,
    component_type      salary_component_type NOT NULL,
    old_amount          DECIMAL(15,2),
    new_amount          DECIMAL(15,2) NOT NULL,
    change_reason       TEXT,
    changed_by          UUID REFERENCES employees(id) ON DELETE SET NULL,
    changed_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 4. Function: Auto-log salary component changes
-- ============================================================
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION log_salary_component_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO employee_salary_component_histories (
            component_id, employee_id, component_name, component_type,
            old_amount, new_amount, change_reason, changed_by
        ) VALUES (
            NEW.id, NEW.employee_id, NEW.component_name, NEW.component_type,
            NULL, NEW.amount, 'Initial creation',
            NEW.created_by
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'UPDATE' THEN
        -- Log if amount changed
        IF OLD.amount IS DISTINCT FROM NEW.amount THEN
            INSERT INTO employee_salary_component_histories (
                component_id, employee_id, component_name, component_type,
                old_amount, new_amount, change_reason, changed_by
            ) VALUES (
                NEW.id, NEW.employee_id, NEW.component_name, NEW.component_type,
                OLD.amount, NEW.amount, 'Amount changed',
                NEW.updated_by
            );
        END IF;
        
        -- Log if deactivated
        IF OLD.is_active = TRUE AND NEW.is_active = FALSE THEN
            INSERT INTO employee_salary_component_histories (
                component_id, employee_id, component_name, component_type,
                old_amount, new_amount, change_reason, changed_by
            ) VALUES (
                NEW.id, NEW.employee_id, NEW.component_name, NEW.component_type,
                NEW.amount, 0, 'Component deactivated',
                NEW.updated_by
            );
        END IF;
        
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Apply trigger
CREATE TRIGGER log_salary_component_changes
    AFTER INSERT OR UPDATE ON employee_salary_components
    FOR EACH ROW
    EXECUTE FUNCTION log_salary_component_change();

-- ============================================================
-- 5. Indexes
-- ============================================================
CREATE INDEX idx_emp_salary_comp_employee_id ON employee_salary_components(employee_id);
CREATE INDEX idx_emp_salary_comp_active ON employee_salary_components(employee_id, is_active);
CREATE INDEX idx_emp_salary_comp_type ON employee_salary_components(component_type);
CREATE INDEX idx_emp_salary_comp_hist_employee ON employee_salary_component_histories(employee_id);
CREATE INDEX idx_emp_salary_comp_hist_changed_at ON employee_salary_component_histories(changed_at DESC);

-- +goose Down
DROP TRIGGER IF EXISTS log_salary_component_changes ON employee_salary_components;
DROP FUNCTION IF EXISTS log_salary_component_change();
DROP TABLE IF EXISTS employee_salary_component_histories;
DROP TABLE IF EXISTS employee_salary_components;
DROP TYPE IF EXISTS salary_component_type;
