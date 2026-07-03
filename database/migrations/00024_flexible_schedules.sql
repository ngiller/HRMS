-- +goose Up
-- ============================================================
-- Migration 00024: Flexible Schedule Templates & Employee Schedules
-- ============================================================
-- Level 2 (Daily Schedule + Location) + Level 3 (Date-Based Schedule)
-- 
-- Features:
-- - Schedule templates (reusable patterns: "5 Hari", "Shift Pagi", "Ramadhan")
-- - Per-day-of-week OR per-specific-date assignments
-- - Location assignment per schedule (different location each day)
-- - Effective period (from-until dates)
-- - Priority-based override resolution
-- - Full backward compatibility with existing work_schedules
-- ============================================================

-- ============================================================
-- 1. Schedule Templates (Level 3)
-- Template jadwal yang bisa dipakai ulang oleh banyak karyawan
-- ============================================================
CREATE TABLE schedule_templates (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    schedule_type       VARCHAR(20) NOT NULL DEFAULT 'weekly',
    -- weekly  = jadwal tetap per hari dalam seminggu (Level 2)
    -- shift   = jadwal shift (bisa rotasi)
    -- flexible = kombinasi bebas
    
    is_active           BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE TRIGGER set_schedule_templates_updated_at
    BEFORE UPDATE ON schedule_templates
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- 2. Schedule Template Days (Level 2 + Level 3)
-- Detail jam kerja per hari dalam template
-- day_of_week: 0=Senin, 1=Selasa, ..., 6=Minggu (NULL untuk shift/flexible)
-- ============================================================
CREATE TABLE schedule_template_days (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id             UUID NOT NULL REFERENCES schedule_templates(id) ON DELETE CASCADE,
    day_of_week             INTEGER CHECK (day_of_week IS NULL OR day_of_week BETWEEN 0 AND 6),
    -- NULL = berlaku untuk semua hari (untuk shift)
    
    start_time              TIME NOT NULL DEFAULT '08:00',
    end_time                TIME NOT NULL DEFAULT '17:00',
    break_start             TIME DEFAULT '12:00',
    break_end               TIME DEFAULT '13:00',
    
    late_tolerance_minutes  INTEGER NOT NULL DEFAULT 15,
    early_leave_tolerance   INTEGER NOT NULL DEFAULT 15,
    
    is_active               BOOLEAN DEFAULT TRUE,
    sort_order              INTEGER DEFAULT 0,
    
    UNIQUE(template_id, day_of_week)
);

-- ============================================================
-- 3. Employee Schedules (Jadwal Aktual per Karyawan)
-- 
-- Bisa refer ke template, atau override langsung jam kerja.
-- Bisa di-set per day_of_week (berlaku periodik) ATAU per specific_date.
-- 
-- Level 2: day_of_week diisi, effective_from/until diisi → "Setiap Senin pakai jadwal ini"
-- Level 3: specific_date diisi → "Tanggal X pakai jadwal ini"
-- ============================================================
CREATE TABLE employee_schedules (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Template reference (opsional — jika pakai template)
    template_id         UUID REFERENCES schedule_templates(id) ON DELETE SET NULL,
    
    -- Day/Date specification
    day_of_week         INTEGER CHECK (day_of_week IS NULL OR day_of_week BETWEEN 0 AND 6),
    -- 0=Senin..6=Minggu. NULL jika pakai specific_date atau template override
    specific_date       DATE,
    -- Tanggal spesifik. NULL jika periodik (day_of_week)
    
    -- Override langsung (jika tidak pakai template, atau override dari template)
    start_time          TIME,
    end_time            TIME,
    break_start         TIME,
    break_end           TIME,
    is_remote           BOOLEAN DEFAULT FALSE,
    -- TRUE = WFH / remote, tidak perlu validasi GPS
    
    -- Effective period
    effective_from      DATE NOT NULL,
    effective_until     DATE,    -- NULL = berlaku terus
    
    -- Prioritas: makin tinggi angka, makin menang
    priority            INTEGER DEFAULT 0,
    
    -- Keterangan/alasan
    reason              TEXT,
    
    is_active           BOOLEAN DEFAULT TRUE,
    created_by          UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraint: either day_of_week or specific_date, not both
    CONSTRAINT chk_schedule_date CHECK (
        (day_of_week IS NOT NULL AND specific_date IS NULL)
        OR (specific_date IS NOT NULL AND day_of_week IS NULL)
        OR (day_of_week IS NULL AND specific_date IS NULL)
    ),
    
    -- Unique per employee per specific_date
    CONSTRAINT unique_employee_date UNIQUE (employee_id, specific_date)
);

CREATE TRIGGER set_employee_schedules_updated_at
    BEFORE UPDATE ON employee_schedules
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- 4. Employee Schedule Locations (Lokasi per Jadwal)
-- Setiap jadwal bisa punya beberapa lokasi absensi
-- Level 2: satu lokasi per hari; Level 3: multi-lokasi per jadwal
-- ============================================================
CREATE TABLE employee_schedule_locations (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_schedule_id    UUID NOT NULL REFERENCES employee_schedules(id) ON DELETE CASCADE,
    attendance_location_id  UUID NOT NULL REFERENCES attendance_locations(id) ON DELETE RESTRICT,
    day_of_week             INTEGER CHECK (day_of_week IS NULL OR day_of_week BETWEEN 0 AND 6),
    -- NULL = berlaku untuk semua hari dalam jadwal itu
    
    is_primary              BOOLEAN DEFAULT TRUE,
    sort_order              INTEGER DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 5. Function: Resolve employee schedule for a given date
-- ============================================================
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION resolve_employee_schedule(
    p_employee_id UUID,
    p_date DATE
)
RETURNS TABLE (
    schedule_id         UUID,
    start_time          TIME,
    end_time            TIME,
    break_start         TIME,
    break_end           TIME,
    is_remote           BOOLEAN,
    location_id         UUID,
    location_name       VARCHAR(255),
    latitude            DECIMAL(10,7),
    longitude           DECIMAL(10,7),
    radius_meters       INTEGER,
    source              TEXT
) AS $$
DECLARE
    v_day_of_week INTEGER := EXTRACT(DOW FROM p_date);
    -- PostgreSQL DOW: 0=Sunday, 1=Monday, ... 6=Saturday
    -- Kami pake: 0=Senin, 1=Selasa, ... 6=Minggu
    v_dow_local INTEGER := CASE WHEN v_day_of_week = 0 THEN 6 ELSE v_day_of_week - 1 END;
BEGIN
    RETURN QUERY
    -- Step 1: Cari jadwal dengan specific_date match (highest priority)
    SELECT 
        es.id,
        COALESCE(es.start_time, std.start_time, '08:00'::time),
        COALESCE(es.end_time, std.end_time, '17:00'::time),
        COALESCE(es.break_start, std.break_start, '12:00'::time),
        COALESCE(es.break_end, std.break_end, '13:00'::time),
        COALESCE(es.is_remote, FALSE),
        al.id,
        al.name,
        al.latitude,
        al.longitude,
        al.radius_meters,
        'date_override'
    FROM employee_schedules es
    LEFT JOIN schedule_templates st ON st.id = es.template_id AND st.is_active = TRUE
    LEFT JOIN schedule_template_days std ON std.template_id = st.id 
        AND (std.day_of_week IS NULL OR std.day_of_week = v_dow_local)
    LEFT JOIN employee_schedule_locations esl ON esl.employee_schedule_id = es.id
        AND (esl.day_of_week IS NULL OR esl.day_of_week = v_dow_local)
        AND esl.is_primary = TRUE
    LEFT JOIN attendance_locations al ON al.id = esl.attendance_location_id AND al.is_active = TRUE
    WHERE es.employee_id = p_employee_id
        AND es.is_active = TRUE
        AND es.specific_date = p_date
        AND es.deleted_at IS NULL
    
    UNION ALL
    
    -- Step 2: Cari jadwal periodik (day_of_week) yang aktif di rentang tanggal
    SELECT 
        es.id,
        COALESCE(es.start_time, std.start_time, '08:00'::time),
        COALESCE(es.end_time, std.end_time, '17:00'::time),
        COALESCE(es.break_start, std.break_start, '12:00'::time),
        COALESCE(es.break_end, std.break_end, '13:00'::time),
        COALESCE(es.is_remote, FALSE),
        al.id,
        al.name,
        al.latitude,
        al.longitude,
        al.radius_meters,
        'weekly_schedule'
    FROM employee_schedules es
    LEFT JOIN schedule_templates st ON st.id = es.template_id AND st.is_active = TRUE
    LEFT JOIN schedule_template_days std ON std.template_id = st.id 
        AND (std.day_of_week IS NULL OR std.day_of_week = v_dow_local)
    LEFT JOIN employee_schedule_locations esl ON esl.employee_schedule_id = es.id
        AND (esl.day_of_week IS NULL OR esl.day_of_week = v_dow_local)
        AND esl.is_primary = TRUE
    LEFT JOIN attendance_locations al ON al.id = esl.attendance_location_id AND al.is_active = TRUE
    WHERE es.employee_id = p_employee_id
        AND es.is_active = TRUE
        AND es.day_of_week = v_dow_local
        AND es.effective_from <= p_date
        AND (es.effective_until IS NULL OR es.effective_until >= p_date)
        AND es.specific_date IS NULL
        AND es.deleted_at IS NULL
    
    ORDER BY priority DESC, source ASC
    LIMIT 1;
    
    -- Jika tidak ada hasil, function ini return empty set
    -- Caller bisa fallback ke employees.work_schedule_id → departments.work_schedule_id
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- ============================================================
-- 6. Seed default templates from existing work_schedules
-- ============================================================
INSERT INTO schedule_templates (name, description, schedule_type)
SELECT 
    name,
    COALESCE(description, 'Auto-migrated from work_schedules'),
    CASE 
        WHEN schedule_type::text = '5_day' THEN 'weekly'
        WHEN schedule_type::text = '6_day' THEN 'weekly'
        ELSE 'shift'
    END
FROM work_schedules
WHERE deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM schedule_templates);

-- Seed template days from work_schedules
INSERT INTO schedule_template_days (template_id, day_of_week, start_time, end_time, break_start, break_end, late_tolerance_minutes, early_leave_tolerance)
SELECT 
    st.id, 0, ws.monday_start, ws.monday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 0)
UNION ALL
SELECT st.id, 1, ws.tuesday_start, ws.tuesday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 1)
UNION ALL
SELECT st.id, 2, ws.wednesday_start, ws.wednesday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 2)
UNION ALL
SELECT st.id, 3, ws.thursday_start, ws.thursday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 3)
UNION ALL
SELECT st.id, 4, ws.friday_start, ws.friday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 4)
UNION ALL
SELECT st.id, 5, ws.saturday_start, ws.saturday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL AND ws.saturday_start IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 5)
UNION ALL
SELECT st.id, 6, ws.sunday_start, ws.sunday_end, ws.break_start, ws.break_end, 
    ws.late_tolerance_minutes, ws.early_leave_tolerance
FROM work_schedules ws
JOIN schedule_templates st ON st.name = ws.name AND st.description LIKE 'Auto-migrated%'
WHERE ws.deleted_at IS NULL AND ws.sunday_start IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM schedule_template_days WHERE template_id = st.id AND day_of_week = 6);

-- ============================================================
-- 7. Indexes
-- ============================================================
CREATE INDEX idx_employee_schedules_employee ON employee_schedules(employee_id, is_active);
CREATE INDEX idx_employee_schedules_date ON employee_schedules(employee_id, specific_date) WHERE specific_date IS NOT NULL;
CREATE INDEX idx_employee_schedules_period ON employee_schedules(employee_id, effective_from, effective_until) WHERE specific_date IS NULL;
CREATE INDEX idx_employee_schedules_dow ON employee_schedules(employee_id, day_of_week) WHERE day_of_week IS NOT NULL;
CREATE INDEX idx_schedule_template_days_template ON schedule_template_days(template_id);
CREATE INDEX idx_employee_schedule_locations_schedule ON employee_schedule_locations(employee_schedule_id);

-- +goose Down
DROP FUNCTION IF EXISTS resolve_employee_schedule(UUID, DATE);
DROP TABLE IF EXISTS employee_schedule_locations;
DROP TABLE IF EXISTS employee_schedules;
DROP TABLE IF EXISTS schedule_template_days;
DROP TABLE IF EXISTS schedule_templates;
