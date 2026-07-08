-- +goose Up
-- ============================================================
-- Migration 00044: Resign & Exit Management
-- ============================================================

-- Resign Requests
CREATE TABLE resign_requests (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id       UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    resign_date       DATE NOT NULL DEFAULT CURRENT_DATE,
    last_working_date DATE NOT NULL,
    reason            TEXT NOT NULL,
    resign_type       VARCHAR(50) NOT NULL DEFAULT 'voluntary',  -- voluntary, termination, retirement, mutual
    status            VARCHAR(20) NOT NULL DEFAULT 'pending',    -- pending, approved, rejected, processed
    approved_by       UUID REFERENCES employees(id) ON DELETE SET NULL,
    approved_at       TIMESTAMPTZ,
    rejection_reason  TEXT,
    approval_trail    JSONB DEFAULT '[]'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE TRIGGER set_resign_requests_updated_at
    BEFORE UPDATE ON resign_requests
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE INDEX idx_resign_requests_employee ON resign_requests(employee_id);
CREATE INDEX idx_resign_requests_status ON resign_requests(status);

-- Exit Clearance Items
CREATE TABLE exit_clearance_items (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resign_id      UUID NOT NULL REFERENCES resign_requests(id) ON DELETE CASCADE,
    item_name      VARCHAR(255) NOT NULL,
    description    TEXT,
    is_checked     BOOLEAN DEFAULT FALSE,
    checked_by     UUID REFERENCES employees(id) ON DELETE SET NULL,
    checked_at     TIMESTAMPTZ,
    sort_order     INTEGER DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_exit_clearance_items_updated_at
    BEFORE UPDATE ON exit_clearance_items
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE INDEX idx_exit_clearance_resign ON exit_clearance_items(resign_id);

-- +goose Down
DROP TABLE IF EXISTS exit_clearance_items;
DROP TABLE IF EXISTS resign_requests;
