-- +goose Up
-- ============================================================
-- Migration 00004: Work Schedules & Attendance Locations
-- ============================================================

-- ============================================================
-- Work Schedules
-- 5-day (Senin-Jumat) or 6-day (Senin-Sabtu) with shift support
-- ============================================================
CREATE TABLE work_schedules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    schedule_type   work_schedule_type NOT NULL DEFAULT '5_day',
    description     TEXT,

    -- Monday
    monday_start    TIME NOT NULL DEFAULT '08:00',
    monday_end      TIME NOT NULL DEFAULT '17:00',
    -- Tuesday
    tuesday_start   TIME NOT NULL DEFAULT '08:00',
    tuesday_end     TIME NOT NULL DEFAULT '17:00',
    -- Wednesday
    wednesday_start TIME NOT NULL DEFAULT '08:00',
    wednesday_end   TIME NOT NULL DEFAULT '17:00',
    -- Thursday
    thursday_start  TIME NOT NULL DEFAULT '08:00',
    thursday_end    TIME NOT NULL DEFAULT '17:00',
    -- Friday
    friday_start    TIME NOT NULL DEFAULT '08:00',
    friday_end      TIME NOT NULL DEFAULT '17:00',
    -- Saturday
    saturday_start  TIME DEFAULT NULL,
    saturday_end    TIME DEFAULT NULL,
    -- Sunday
    sunday_start    TIME DEFAULT NULL,
    sunday_end      TIME DEFAULT NULL,

    -- Break time
    break_start     TIME NOT NULL DEFAULT '12:00',
    break_end       TIME NOT NULL DEFAULT '13:00',

    -- Tolerance settings
    late_tolerance_minutes  INTEGER NOT NULL DEFAULT 15,  -- Grazing period
    early_leave_tolerance   INTEGER NOT NULL DEFAULT 15,

    -- Work hours calculation
    weekly_hours    DECIMAL(4,1) NOT NULL DEFAULT 40.0,

    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_work_schedules_updated_at
    BEFORE UPDATE ON work_schedules
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Attendance Locations (GPS based)
-- Bisa lebih dari satu lokasi absensi
-- ============================================================
CREATE TABLE attendance_locations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,       -- Kantor Pusat, Cabang A, Gudang, dll
    address         TEXT,
    latitude        DECIMAL(10,7) NOT NULL,
    longitude       DECIMAL(10,7) NOT NULL,
    radius_meters   INTEGER NOT NULL DEFAULT 100,  -- Radius toleransi GPS
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_attendance_locations_updated_at
    BEFORE UPDATE ON attendance_locations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Seed default schedules
-- ============================================================
INSERT INTO work_schedules (name, schedule_type,
    monday_start, monday_end,
    tuesday_start, tuesday_end,
    wednesday_start, wednesday_end,
    thursday_start, thursday_end,
    friday_start, friday_end,
    break_start, break_end,
    weekly_hours)
VALUES
    ('5 Hari Kerja (Senin-Jumat)', '5_day',
     '08:00', '17:00',
     '08:00', '17:00',
     '08:00', '17:00',
     '08:00', '17:00',
     '08:00', '17:00',
     '12:00', '13:00',
     40.0);

INSERT INTO work_schedules (name, schedule_type,
    monday_start, monday_end,
    tuesday_start, tuesday_end,
    wednesday_start, wednesday_end,
    thursday_start, thursday_end,
    friday_start, friday_end,
    saturday_start, saturday_end,
    break_start, break_end,
    weekly_hours)
VALUES
    ('6 Hari Kerja (Senin-Sabtu)', '6_day',
     '08:00', '16:00',
     '08:00', '16:00',
     '08:00', '16:00',
     '08:00', '16:00',
     '08:00', '16:00',
     '08:00', '13:00',
     '12:00', '13:00',
     40.0);

-- Indexes
CREATE INDEX idx_attendance_locations_coordinates
    ON attendance_locations(latitude, longitude);

-- +goose Down
DROP TABLE IF EXISTS attendance_locations;
DROP TABLE IF EXISTS work_schedules;
