-- Migration 00040: Daily Working Journal
-- Description: Create table for daily work journal entries per employee

-- +goose Up
-- +goose StatementBegin

-- Create enum for journal status
DO $$ BEGIN
    CREATE TYPE journal_status AS ENUM ('draft', 'submitted', 'acknowledged');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Create daily_journals table
CREATE TABLE IF NOT EXISTS daily_journals (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    journal_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    work_description TEXT NOT NULL DEFAULT '',
    achievements    TEXT DEFAULT '',
    challenges      TEXT DEFAULT '',
    plan_tomorrow   TEXT DEFAULT '',
    status          journal_status NOT NULL DEFAULT 'draft',
    submitted_at    TIMESTAMPTZ,
    acknowledged_by UUID REFERENCES employees(id),
    acknowledged_at TIMESTAMPTZ,
    department_id   UUID REFERENCES departments(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    -- One journal entry per employee per day
    CONSTRAINT uq_daily_journal_employee_date UNIQUE (employee_id, journal_date)
);

-- Trigger for updated_at
CREATE TRIGGER trg_daily_journals_updated_at
    BEFORE UPDATE ON daily_journals
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Indexes for performance
CREATE INDEX idx_daily_journals_date ON daily_journals(journal_date DESC);
CREATE INDEX idx_daily_journals_employee ON daily_journals(employee_id);
CREATE INDEX idx_daily_journals_department ON daily_journals(department_id);
CREATE INDEX idx_daily_journals_status ON daily_journals(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS daily_journals;
DROP TYPE IF EXISTS journal_status;
-- +goose StatementEnd
