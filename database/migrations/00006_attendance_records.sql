-- +goose Up
-- ============================================================
-- Migration 00006: Attendance Records
-- ============================================================

CREATE TABLE attendance_records (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id             UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date                    DATE NOT NULL,
    
    -- Check In
    check_in_time           TIMESTAMPTZ,
    check_in_photo_url      TEXT,
    check_in_lat            DECIMAL(10,7),
    check_in_lng            DECIMAL(10,7),
    check_in_location_id    UUID REFERENCES attendance_locations(id) ON DELETE SET NULL,
    check_in_location_name  VARCHAR(255),
    face_match_score        DECIMAL(5,2),                  -- 0-100 score
    
    -- Check Out
    check_out_time          TIMESTAMPTZ,
    check_out_photo_url     TEXT,
    check_out_lat           DECIMAL(10,7),
    check_out_lng           DECIMAL(10,7),
    check_out_location_id   UUID REFERENCES attendance_locations(id) ON DELETE SET NULL,
    check_out_location_name VARCHAR(255),
    
    -- Status & Metadata
    status                  attendance_status NOT NULL DEFAULT 'hadir',
    is_late                 BOOLEAN DEFAULT FALSE,
    late_minutes            INTEGER DEFAULT 0,
    is_early_leave          BOOLEAN DEFAULT FALSE,
    early_leave_minutes     INTEGER DEFAULT 0,
    total_work_hours        DECIMAL(5,2),                  -- Total jam kerja
    
    -- Manual Entry
    is_manual_entry         BOOLEAN DEFAULT FALSE,
    manual_entry_reason     TEXT,
    manual_entry_approved_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    
    -- Device Info
    device_info             JSONB,                         -- {user_agent, platform, device_name, app_version}
    
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint: 1 record per employee per day
    UNIQUE(employee_id, date)
);

CREATE TRIGGER set_attendance_records_updated_at
    BEFORE UPDATE ON attendance_records
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Manual Attendance Requests
-- ============================================================
CREATE TABLE manual_attendance_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    check_in_time   TIMESTAMPTZ,
    check_out_time  TIMESTAMPTZ,
    reason          TEXT NOT NULL,
    status          leave_status NOT NULL DEFAULT 'pending',
    approved_by     UUID REFERENCES employees(id) ON DELETE SET NULL,
    approved_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_manual_attendance_requests_updated_at
    BEFORE UPDATE ON manual_attendance_requests
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Attendance Summary (Materialized view untuk dashboard)
-- ============================================================
CREATE MATERIALIZED VIEW attendance_summary AS
SELECT
    employee_id,
    DATE_TRUNC('month', date) AS month,
    COUNT(*) FILTER (WHERE status = 'hadir') AS total_hadir,
    COUNT(*) FILTER (WHERE status = 'terlambat') AS total_terlambat,
    COUNT(*) FILTER (WHERE status = 'izin') AS total_izin,
    COUNT(*) FILTER (WHERE status = 'sakit') AS total_sakit,
    COUNT(*) FILTER (WHERE status = 'tanpa_keterangan') AS total_alpa,
    COUNT(*) FILTER (WHERE is_late = TRUE) AS total_late_days,
    AVG(total_work_hours) AS avg_work_hours,
    AVG(late_minutes) FILTER (WHERE is_late = TRUE) AS avg_late_minutes
FROM attendance_records
WHERE deleted_at IS NULL
GROUP BY employee_id, DATE_TRUNC('month', date);

CREATE UNIQUE INDEX idx_attendance_summary_employee_month
    ON attendance_summary(employee_id, month);

-- Indexes
CREATE INDEX idx_attendance_records_employee_date 
    ON attendance_records(employee_id, date DESC);
CREATE INDEX idx_attendance_records_date 
    ON attendance_records(date);
CREATE INDEX idx_attendance_records_status 
    ON attendance_records(status);
CREATE INDEX idx_attendance_records_employee_month 
    ON attendance_records(employee_id, (DATE_TRUNC('month', date)));

-- +goose Down
DROP MATERIALIZED VIEW IF EXISTS attendance_summary;
DROP TABLE IF EXISTS manual_attendance_requests;
DROP TABLE IF EXISTS attendance_records;
