-- +goose Up
-- ============================================================
-- Migration 00021: Fix Triggers & Add Audit Coverage
-- ============================================================
-- Perbaikan:
-- 1. Fix log_salary_component_change — pakai current_setting
--    instead of NEW.created_by / NEW.updated_by agar konsisten
--    dengan audit trigger yang sudah menggunakan current_setting.
-- 2. Fix audit_trigger_function — entity_name_val pakai OLD
--    untuk DELETE (NEW = NULL saat DELETE!), bukan NEW.
-- 3. Tambah audit trigger ke tabel kunci lainnya.
-- 4. Tambah function refresh_attendance_summary().
-- ============================================================

-- ============================================================
-- 1. Fix: log_salary_component_change()
--    Ganti NEW.created_by / NEW.updated_by →
--    current_setting('app.current_user_id')::UUID
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
            current_setting('app.current_user_id')::UUID
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
                current_setting('app.current_user_id')::UUID
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
                current_setting('app.current_user_id')::UUID
            );
        END IF;

        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- ============================================================
-- 2. Fix: audit_trigger_function()
--    entity_name_val harus menggunakan OLD untuk operasi DELETE,
--    karena NEW = NULL saat DELETE!
-- ============================================================
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    entity_name_val VARCHAR(255);
BEGIN
    -- Determine entity name based on table and operation
    IF TG_OP = 'DELETE' THEN
        -- For DELETE, use OLD values (NEW is NULL)
        IF TG_TABLE_NAME = 'employees' THEN
            entity_name_val := OLD.full_name;
        ELSIF TG_TABLE_NAME = 'leave_requests' THEN
            entity_name_val := 'Leave request from ' || (SELECT full_name FROM employees WHERE id = OLD.employee_id);
        ELSIF TG_TABLE_NAME = 'departments' THEN
            entity_name_val := OLD.name;
        ELSIF TG_TABLE_NAME = 'positions' THEN
            entity_name_val := OLD.name;
        ELSIF TG_TABLE_NAME = 'roles' THEN
            entity_name_val := OLD.name;
        ELSIF TG_TABLE_NAME = 'employee_salary_components' THEN
            entity_name_val := OLD.component_name;
        ELSE
            entity_name_val := TG_TABLE_NAME || ' ' || OLD.id::TEXT;
        END IF;

        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, old_values)
        VALUES (current_setting('app.current_user_id')::UUID, 'delete', TG_TABLE_NAME, OLD.id, entity_name_val, row_to_json(OLD)::jsonb);
        RETURN OLD;
    ELSE
        -- For INSERT and UPDATE, use NEW values
        IF TG_TABLE_NAME = 'employees' THEN
            entity_name_val := NEW.full_name;
        ELSIF TG_TABLE_NAME = 'leave_requests' THEN
            entity_name_val := 'Leave request from ' || (SELECT full_name FROM employees WHERE id = NEW.employee_id);
        ELSIF TG_TABLE_NAME = 'departments' THEN
            entity_name_val := NEW.name;
        ELSIF TG_TABLE_NAME = 'positions' THEN
            entity_name_val := NEW.name;
        ELSIF TG_TABLE_NAME = 'roles' THEN
            entity_name_val := NEW.name;
        ELSIF TG_TABLE_NAME = 'employee_salary_components' THEN
            entity_name_val := NEW.component_name;
        ELSE
            entity_name_val := TG_TABLE_NAME || ' ' || NEW.id::TEXT;
        END IF;

        IF TG_OP = 'INSERT' THEN
            INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, new_values)
            VALUES (current_setting('app.current_user_id')::UUID, 'create', TG_TABLE_NAME, NEW.id, entity_name_val, row_to_json(NEW)::jsonb);
            RETURN NEW;
        ELSE
            INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, old_values, new_values)
            VALUES (current_setting('app.current_user_id')::UUID, 'update', TG_TABLE_NAME, NEW.id, entity_name_val, row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb);
            RETURN NEW;
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- ============================================================
-- 3. Tambah Audit Triggers ke Tabel Kunci
-- ============================================================

-- Departments
CREATE TRIGGER audit_departments
    AFTER INSERT OR UPDATE OR DELETE ON departments
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Positions
CREATE TRIGGER audit_positions
    AFTER INSERT OR UPDATE OR DELETE ON positions
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Leave Requests
CREATE TRIGGER audit_leave_requests
    AFTER INSERT OR UPDATE OR DELETE ON leave_requests
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Payroll Items
CREATE TRIGGER audit_payroll_items
    AFTER INSERT OR UPDATE OR DELETE ON payroll_items
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Reimbursements
CREATE TRIGGER audit_reimbursements
    AFTER INSERT OR UPDATE OR DELETE ON reimbursements
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Overtime Requests
CREATE TRIGGER audit_overtime_requests
    AFTER INSERT OR UPDATE OR DELETE ON overtime_requests
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Loans
CREATE TRIGGER audit_loans
    AFTER INSERT OR UPDATE OR DELETE ON loans
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Employee Salary Components
CREATE TRIGGER audit_employee_salary_components
    AFTER INSERT OR UPDATE OR DELETE ON employee_salary_components
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- Roles
CREATE TRIGGER audit_roles
    AFTER INSERT OR UPDATE OR DELETE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- ============================================================
-- 4. Function: refresh_attendance_summary()
--    Untuk merefresh materialized view attendance_summary.
--    Panggil secara periodik via cron atau endpoint.
-- ============================================================
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION refresh_attendance_summary()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY attendance_summary;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- ============================================================
-- Rollback
-- ============================================================

-- 4. Hapus function refresh
DROP FUNCTION IF EXISTS refresh_attendance_summary();

-- 3. Hapus audit triggers (urutan terbalik)
DROP TRIGGER IF EXISTS audit_roles ON roles;
DROP TRIGGER IF EXISTS audit_employee_salary_components ON employee_salary_components;
DROP TRIGGER IF EXISTS audit_loans ON loans;
DROP TRIGGER IF EXISTS audit_overtime_requests ON overtime_requests;
DROP TRIGGER IF EXISTS audit_reimbursements ON reimbursements;
DROP TRIGGER IF EXISTS audit_payroll_items ON payroll_items;
DROP TRIGGER IF EXISTS audit_leave_requests ON leave_requests;
DROP TRIGGER IF EXISTS audit_positions ON positions;
DROP TRIGGER IF EXISTS audit_departments ON departments;

-- 2. Restore original audit_trigger_function (dengan DELETE bug)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    entity_name_val VARCHAR(255);
BEGIN
    IF TG_TABLE_NAME = 'employees' THEN
        entity_name_val := NEW.full_name;
    ELSIF TG_TABLE_NAME = 'leave_requests' THEN
        entity_name_val := 'Leave request from ' || (SELECT full_name FROM employees WHERE id = NEW.employee_id);
    ELSE
        entity_name_val := TG_TABLE_NAME || ' ' || NEW.id::TEXT;
    END IF;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, new_values)
        VALUES (current_setting('app.current_user_id')::UUID, 'create', TG_TABLE_NAME, NEW.id, entity_name_val, row_to_json(NEW)::jsonb);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, old_values, new_values)
        VALUES (current_setting('app.current_user_id')::UUID, 'update', TG_TABLE_NAME, NEW.id, entity_name_val, row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, old_values)
        VALUES (current_setting('app.current_user_id')::UUID, 'delete', TG_TABLE_NAME, OLD.id, entity_name_val, row_to_json(OLD)::jsonb);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- 1. Restore original log_salary_component_change (menggunakan created_by/updated_by)
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

