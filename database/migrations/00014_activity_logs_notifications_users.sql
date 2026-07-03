-- +goose Up
-- ============================================================
-- Migration 00014: Activity Logs, Notifications, Users & Auth
-- ============================================================

-- ============================================================
-- Activity / Audit Trail
-- ============================================================
CREATE TABLE activity_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID REFERENCES employees(id) ON DELETE SET NULL,
    action          VARCHAR(50) NOT NULL,                    -- create, update, delete, approve, reject, view, login, logout
    entity_type     VARCHAR(50) NOT NULL,                    -- employee, leave, payroll, reimbursement, dll
    entity_id       UUID,
    entity_name     VARCHAR(255),                            -- Nama/deskripsi entity untuk display
    old_values      JSONB,
    new_values      JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partition by month for better performance
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_entity ON activity_logs(entity_type, entity_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);

-- ============================================================
-- Notifications
-- ============================================================
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    notification_type notification_type NOT NULL,
    title           VARCHAR(255) NOT NULL,
    body            TEXT,
    
    -- For deep linking
    data            JSONB DEFAULT '{}'::jsonb,              -- {entity_type, entity_id, url}
    
    -- Status
    is_read         BOOLEAN DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    
    -- Push notification delivery
    is_pushed       BOOLEAN DEFAULT FALSE,
    pushed_at       TIMESTAMPTZ,
    
    is_email_sent   BOOLEAN DEFAULT FALSE,
    email_sent_at   TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id) WHERE is_read = FALSE;

-- ============================================================
-- Notification Preferences (per user)
-- ============================================================
CREATE TABLE notification_preferences (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    notification_type notification_type NOT NULL,
    in_app          BOOLEAN DEFAULT TRUE,
    push            BOOLEAN DEFAULT TRUE,
    email           BOOLEAN DEFAULT TRUE,
    whatsapp        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, notification_type)
);

CREATE TRIGGER set_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Seed default notification preferences
INSERT INTO notification_preferences (user_id, notification_type, in_app, push, email)
SELECT e.id, nt.type_name::notification_type, TRUE, TRUE, TRUE
FROM employees e
CROSS JOIN (VALUES 
    ('approval_request'), ('approved'), ('rejected'), 
    ('announcement'), ('reminder'), ('system')
) AS nt(type_name);

-- ============================================================
-- User Sessions (for JWT refresh token tracking)
-- ============================================================
CREATE TABLE user_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    refresh_token   TEXT NOT NULL,
    device_info     JSONB DEFAULT '{}'::jsonb,
    ip_address      INET,
    is_active       BOOLEAN DEFAULT TRUE,
    last_used_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id, is_active);
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token);

-- ============================================================
-- Audit Trail Trigger Function (auto-log changes)
-- ============================================================
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

-- Apply audit trigger to employees table
CREATE TRIGGER audit_employees
    AFTER INSERT OR UPDATE OR DELETE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_function();

-- ============================================================
-- Password Reset Tokens
-- ============================================================
CREATE TABLE password_reset_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token           VARCHAR(255) NOT NULL UNIQUE,
    is_used         BOOLEAN DEFAULT FALSE,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_employee ON password_reset_tokens(employee_id);

-- Account lockout tracking
CREATE TABLE login_attempts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    ip_address      INET,
    is_successful   BOOLEAN NOT NULL,
    attempted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_login_attempts_employee ON login_attempts(employee_id, attempted_at DESC);

-- +goose Down
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS activity_logs;

DROP TRIGGER IF EXISTS audit_employees ON employees;
DROP FUNCTION IF EXISTS audit_trigger_function;
