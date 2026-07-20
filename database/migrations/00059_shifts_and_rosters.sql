-- Migration 00059: Shifts & Department Rosters
-- Description: Shift definitions + monthly department rosters for shift scheduling
-- Flow: Create Shift → Create Roster (monthly) → Assign employees to shifts per date

-- +goose Up
-- +goose StatementBegin

-- ─── 1. SHIFTS ──────────────────────────────────────────────
-- Definisi shift reusable: Pagi, Siang, Malam, dll
CREATE TABLE IF NOT EXISTS shifts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    break_start     TIME DEFAULT '12:00',
    break_end       TIME DEFAULT '13:00',
    color           VARCHAR(7) DEFAULT '#3B82F6',
    description     TEXT DEFAULT '',
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_code ON shifts(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shifts_active ON shifts(is_active) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_shifts_updated_at ON shifts;
CREATE TRIGGER set_shifts_updated_at
    BEFORE UPDATE ON shifts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Seed default shifts
INSERT INTO shifts (name, code, start_time, end_time, break_start, break_end, color, description) VALUES
    ('Pagi', 'morning', '06:00', '14:00', '10:00', '10:30', '#3B82F6', 'Shift Pagi: 06:00 - 14:00'),
    ('Siang', 'afternoon', '14:00', '22:00', '18:00', '18:30', '#F59E0B', 'Shift Siang: 14:00 - 22:00'),
    ('Malam', 'night', '22:00', '06:00', '01:00', '01:30', '#8B5CF6', 'Shift Malam: 22:00 - 06:00'),
    ('Non-Shift', 'regular', '08:00', '17:00', '12:00', '13:00', '#10B981', 'Jam kerja reguler: 08:00 - 17:00')
ON CONFLICT DO NOTHING;

-- ─── 2. DEPARTMENT ROSTERS ─────────────────────────────────
-- Roster bulanan per departemen
CREATE TABLE IF NOT EXISTS department_rosters (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department_id   UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    month           INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year            INTEGER NOT NULL,
    is_published    BOOLEAN DEFAULT FALSE,
    notes           TEXT DEFAULT '',
    created_by      UUID REFERENCES employees(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Drop old unique constraint if it exists (which doesn't allow soft-delete duplicates)
ALTER TABLE department_rosters DROP CONSTRAINT IF EXISTS department_rosters_department_id_name_key;

-- Create partial unique index (allows duplicate names for soft-deleted rosters)
CREATE UNIQUE INDEX IF NOT EXISTS idx_rosters_dept_name_active ON department_rosters(department_id, name) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_rosters_department ON department_rosters(department_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_rosters_period ON department_rosters(year, month, deleted_at);

DROP TRIGGER IF EXISTS set_department_rosters_updated_at ON department_rosters;
CREATE TRIGGER set_department_rosters_updated_at
    BEFORE UPDATE ON department_rosters
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ─── 3. ROSTER ENTRIES ─────────────────────────────────────
-- Assign karyawan ke shift pada tanggal tertentu dalam roster
CREATE TABLE IF NOT EXISTS roster_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roster_id       UUID NOT NULL REFERENCES department_rosters(id) ON DELETE CASCADE,
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    shift_id        UUID NOT NULL REFERENCES shifts(id),
    notes           TEXT DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Drop old unique constraint if it exists
ALTER TABLE roster_entries DROP CONSTRAINT IF EXISTS roster_entries_roster_id_employee_id_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_roster_entries_unique ON roster_entries(roster_id, employee_id, date);

CREATE INDEX IF NOT EXISTS idx_roster_entries_roster ON roster_entries(roster_id);
CREATE INDEX IF NOT EXISTS idx_roster_entries_employee ON roster_entries(employee_id, date);
CREATE INDEX IF NOT EXISTS idx_roster_entries_date ON roster_entries(date);

DROP TRIGGER IF EXISTS set_roster_entries_updated_at ON roster_entries;
CREATE TRIGGER set_roster_entries_updated_at
    BEFORE UPDATE ON roster_entries
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

COMMENT ON TABLE shifts IS 'Master data shift kerja (Pagi, Siang, Malam, dll)';
COMMENT ON TABLE department_rosters IS 'Roster bulanan per departemen';
COMMENT ON TABLE roster_entries IS 'Assign karyawan ke shift per tanggal dalam roster';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS roster_entries CASCADE;
DROP TABLE IF EXISTS department_rosters CASCADE;
DROP TABLE IF EXISTS shifts CASCADE;
-- +goose StatementEnd
