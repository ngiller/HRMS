-- +goose Up
-- ============================================================
-- Migration 00002: Companies & Company Settings
-- ============================================================

CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    legal_name      VARCHAR(255),
    address         TEXT,
    city            VARCHAR(100),
    province        VARCHAR(100),
    postal_code     VARCHAR(20),
    phone           VARCHAR(50),
    email           VARCHAR(255),
    website         VARCHAR(255),
    npwp            VARCHAR(50),
    logo_url        TEXT,
    bpjs_ks_number  VARCHAR(50),      -- BPJS Kesehatan
    bpjs_jht_number VARCHAR(50),      -- BPJS JHT
    bpjs_jp_number  VARCHAR(50),      -- BPJS JP
    bpjs_jkk_rate   DECIMAL(5,2) DEFAULT 0.54,  -- Risk-based rate (0.24% - 1.74%)
    -- Company-wide HR settings (JSONB for flexibility)
    hr_settings     JSONB DEFAULT '{}'::jsonb,
    -- Approval settings
    approval_config JSONB DEFAULT '{
        "leave_approval_limit_days": 3,
        "reimbursement_approval_limit": 5000000,
        "loan_approval_limit": 10000000,
        "overtime_max_hours_per_day": 3,
        "overtime_max_hours_per_week": 14,
        "manual_attendance_max_per_month": 3
    }'::jsonb,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Trigger untuk auto-update updated_at
CREATE TRIGGER set_companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS companies;
