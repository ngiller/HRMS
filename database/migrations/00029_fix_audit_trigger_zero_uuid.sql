-- +goose Up
-- ============================================================
-- Migration 00029: Fix audit trigger — handle zero UUID user_id
-- ============================================================
-- Masalah: app.current_user_id default-nya '00000000-0000-0000-0000-000000000000'
-- (UUID all-zeros). Trigger audit mencoba INSERT user_id ini ke
-- activity_logs, tapi FK constraint activity_logs_user_id_fkey
-- nge-refer ke employees(id), dan UUID all-zeros tidak ada di tabel.
--
-- Solusi: Jika current_user_id adalah all-zeros, gunakan NULL.
-- ============================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    entity_name_val VARCHAR(255);
    user_id_val UUID;
BEGIN
    -- Use NULL when current_user_id is the default zero UUID (system action)
    user_id_val := current_setting('app.current_user_id')::UUID;
    IF user_id_val = '00000000-0000-0000-0000-000000000000' THEN
        user_id_val := NULL;
    END IF;

    -- Determine entity name based on table
    IF TG_TABLE_NAME = 'employees' THEN
        entity_name_val := NEW.full_name;
    ELSIF TG_TABLE_NAME = 'leave_requests' THEN
        entity_name_val := 'Leave request from ' || (SELECT full_name FROM employees WHERE id = NEW.employee_id);
    ELSE
        entity_name_val := TG_TABLE_NAME || ' ' || NEW.id::TEXT;
    END IF;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, new_values)
        VALUES (user_id_val, 'create', TG_TABLE_NAME, NEW.id, entity_name_val, row_to_json(NEW)::jsonb);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, old_values, new_values)
        VALUES (user_id_val, 'update', TG_TABLE_NAME, NEW.id, entity_name_val, row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO activity_logs (user_id, action, entity_type, entity_id, entity_name, old_values)
        VALUES (user_id_val, 'delete', TG_TABLE_NAME, OLD.id, entity_name_val, row_to_json(OLD)::jsonb);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    entity_name_val VARCHAR(255);
BEGIN
    -- Determine entity name based on table
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
