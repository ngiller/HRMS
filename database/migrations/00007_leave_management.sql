-- +goose Up
-- ============================================================
-- Migration 00007: Leave Management
-- ============================================================

-- ============================================================
-- Leave Types
-- ============================================================
CREATE TABLE leave_types (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                    VARCHAR(255) NOT NULL,
    code                    VARCHAR(50) UNIQUE NOT NULL,
    default_quota           INTEGER NOT NULL DEFAULT 0,     -- Per year
    is_paid                 BOOLEAN NOT NULL DEFAULT TRUE,
    requires_document       BOOLEAN DEFAULT FALSE,
    max_consecutive_days    INTEGER,                        -- NULL = unlimited
    can_rollover            BOOLEAN DEFAULT FALSE,
    rollover_max_days       INTEGER DEFAULT 0,
    can_cashout             BOOLEAN DEFAULT FALSE,
    is_active               BOOLEAN DEFAULT TRUE,
    sort_order              INTEGER DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE TRIGGER set_leave_types_updated_at
    BEFORE UPDATE ON leave_types
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Leave Balances (per employee per year)
-- ============================================================
CREATE TABLE leave_balances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id   UUID NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    year            INTEGER NOT NULL,
    total_quota     INTEGER NOT NULL,
    used            INTEGER NOT NULL DEFAULT 0,
    remaining       INTEGER NOT NULL,
    rolled_over_from INTEGER DEFAULT 0,     -- Sisa cuti tahun sebelumnya
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(employee_id, leave_type_id, year)
);

CREATE TRIGGER set_leave_balances_updated_at
    BEFORE UPDATE ON leave_balances
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Leave Requests (dengan approval bertingkat)
-- ============================================================
CREATE TABLE leave_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id   UUID NOT NULL REFERENCES leave_types(id) ON DELETE RESTRICT,
    
    -- Date range
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    total_days      INTEGER NOT NULL,
    is_half_day     BOOLEAN DEFAULT FALSE,                 -- Cuti setengah hari
    
    -- Details
    reason          TEXT NOT NULL,
    document_url    TEXT,                                   -- Surat dokter, undangan, dll
    contact_during_leave VARCHAR(100),                      -- No. HP selama cuti
    
    -- Approval (multilevel, disimpan sebagai JSON array)
    approval_trail  JSONB DEFAULT '[]'::jsonb,             -- [{level, approver_id, status, note, date}]
    status          leave_status NOT NULL DEFAULT 'pending',
    
    -- Cancellation
    cancelled_at    TIMESTAMPTZ,
    cancel_reason   TEXT,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_leave_requests_updated_at
    BEFORE UPDATE ON leave_requests
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Seed default leave types for Indonesia
-- ============================================================
INSERT INTO leave_types (name, code, default_quota, is_paid, requires_document, max_consecutive_days, can_rollover, rollover_max_days, can_cashout, sort_order) VALUES
    ('Cuti Tahunan',                'tahunan',     12,  TRUE,  FALSE, NULL,   TRUE,  6,  TRUE,  1),
    ('Cuti Sakit',                  'sakit',       14,  TRUE,  FALSE, NULL,   FALSE, 0,  FALSE, 2),
    ('Cuti Hamil/Melahirkan',       'melahirkan',  90,  TRUE,  TRUE,  90,     FALSE, 0,  FALSE, 3),
    ('Cuti Keguguran',              'keguguran',   30,  TRUE,  TRUE,  30,     FALSE, 0,  FALSE, 4),
    ('Cuti Menikah',                'menikah',     3,   TRUE,  TRUE,  3,      FALSE, 0,  FALSE, 5),
    ('Cuti Menikahkan Anak',        'nikah_anak',  2,   TRUE,  TRUE,  2,      FALSE, 0,  FALSE, 6),
    ('Cuti Khitanan/Baptis Anak',   'khitanan',    2,   TRUE,  FALSE, 2,      FALSE, 0,  FALSE, 7),
    ('Cuti Keluarga Meninggal',     'meninggal',   2,   TRUE,  FALSE, 2,      FALSE, 0,  FALSE, 8),
    ('Cuti Ibadah Haji/Umroh',      'haji',        40,  TRUE,  TRUE,  40,     FALSE, 0,  FALSE, 9),
    ('Cuti Alasan Penting',         'penting',     0,   FALSE, FALSE, NULL,   FALSE, 0,  FALSE, 10);

-- Indexes
CREATE INDEX idx_leave_requests_employee_id ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
CREATE INDEX idx_leave_requests_dates ON leave_requests(start_date, end_date);
CREATE INDEX idx_leave_balances_employee_year ON leave_balances(employee_id, year);
CREATE INDEX idx_leave_types_code ON leave_types(code);

-- +goose Down
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS leave_balances;
DROP TABLE IF EXISTS leave_types;
