-- +goose Up
-- ============================================================
-- Migration 00011: Reprimands, Announcements & Company Holidays
-- ============================================================

-- ============================================================
-- Reprimands (Surat Peringatan / Teguran)
-- ============================================================
CREATE TABLE reprimands (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    reprimand_type      reprimand_type NOT NULL,
    
    -- Content
    title               VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL,
    violation_date      DATE,                                -- Tanggal pelanggaran
    violation_details   TEXT,
    
    -- Issuer
    issued_by           UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    issued_date         DATE NOT NULL DEFAULT CURRENT_DATE,
    
    -- Acknowledgment
    acknowledgment_date TIMESTAMPTZ,
    acknowledgment_note TEXT,
    
    -- Document
    document_url        TEXT,                                -- PDF surat peringatan
    
    -- Period & Status
    effective_period_months INTEGER DEFAULT 6,               -- Masa berlaku peringatan
    status              reprimand_status NOT NULL DEFAULT 'issued',
    expired_at          DATE,
    
    -- Escalation (auto next level)
    escalated_from_id   UUID REFERENCES reprimands(id) ON DELETE SET NULL,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE TRIGGER set_reprimands_updated_at
    BEFORE UPDATE ON reprimands
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Announcements (Pengumuman)
-- ============================================================
CREATE TABLE announcements (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title               VARCHAR(255) NOT NULL,
    content             TEXT NOT NULL,
    announcement_type   announcement_type NOT NULL DEFAULT 'general',
    
    -- Targeting
    target_department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    target_position_grade_id UUID REFERENCES position_grades(id) ON DELETE SET NULL,
    target_all           BOOLEAN DEFAULT TRUE,               -- Jika TRUE, tampilkan ke semua
    
    -- Attachment
    attachment_urls     TEXT[],
    
    -- Display settings
    is_pinned           BOOLEAN DEFAULT FALSE,
    pin_priority        INTEGER DEFAULT 0,                   -- Higher = more prominent
    
    -- Schedule
    published_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at          TIMESTAMPTZ,
    
    -- Author
    created_by          UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE TRIGGER set_announcements_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Announcement Read Tracking
-- ============================================================
CREATE TABLE announcement_reads (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    announcement_id     UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    read_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(announcement_id, employee_id)
);

-- ============================================================
-- Company Holidays (Kalender Hari Libur)
-- ============================================================
CREATE TABLE company_holidays (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date                DATE NOT NULL,
    name                VARCHAR(255) NOT NULL,
    holiday_type        holiday_type NOT NULL DEFAULT 'national',
    is_recurring_yearly BOOLEAN DEFAULT FALSE,               -- Libur nasional yang tetap setiap tahun
    description         TEXT,
    is_active           BOOLEAN DEFAULT TRUE,
    created_by          UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(date, name)
);

CREATE TRIGGER set_company_holidays_updated_at
    BEFORE UPDATE ON company_holidays
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Seed 2026 Indonesian national holidays
-- NOTE: Tanggal-tanggal ini adalah estimasi berdasarkan pola tahun sebelumnya.
-- Harap diverifikasi dengan SKB 3 Menteri resmi sebelum go-live.
INSERT INTO company_holidays (date, name, holiday_type, is_recurring_yearly) VALUES
    ('2026-01-01', 'Tahun Baru Masehi', 'national', TRUE),
    ('2026-01-27', 'Isra Miraj Nabi Muhammad SAW', 'national', TRUE),
    ('2026-01-28', 'Tahun Baru Imlek 2577', 'national', FALSE),
    ('2026-02-11', 'Cuti Bersama Tahun Baru Imlek', 'joint', FALSE),
    ('2026-03-01', 'Hari Suci Nyepi Tahun Baru Saka 1948', 'national', TRUE),
    ('2026-03-31', 'Wafat Isa Almasih', 'national', TRUE),
    ('2026-04-02', 'Cuti Bersama Hari Raya Idul Fitri', 'joint', FALSE),
    ('2026-04-03', 'Cuti Bersama Hari Raya Idul Fitri', 'joint', FALSE),
    ('2026-04-06', 'Cuti Bersama Hari Raya Idul Fitri', 'joint', FALSE),
    ('2026-04-07', 'Cuti Bersama Hari Raya Idul Fitri', 'joint', FALSE),
    ('2026-04-10', 'Hari Raya Idul Fitri 1447 Hijriah', 'national', TRUE),
    ('2026-04-11', 'Hari Raya Idul Fitri 1447 Hijriah', 'national', TRUE),
    ('2026-05-01', 'Hari Buruh Internasional', 'national', TRUE),
    ('2026-05-07', 'Cuti Bersama Idul Fitri', 'joint', FALSE),
    ('2026-05-08', 'Cuti Bersama Idul Fitri', 'joint', FALSE),
    ('2026-05-21', 'Kenaikan Isa Almasih', 'national', TRUE),
    ('2026-05-28', 'Cuti Bersama Waisak', 'joint', FALSE),
    ('2026-06-01', 'Hari Lahir Pancasila', 'national', TRUE),
    ('2026-06-04', 'Hari Raya Waisak 2570', 'national', TRUE),
    ('2026-06-17', 'Idul Adha 1447 Hijriah', 'national', TRUE),
    ('2026-06-22', 'Cuti Bersama Idul Adha', 'joint', FALSE),
    ('2026-07-06', 'Tahun Baru Islam 1448 Hijriah', 'national', TRUE),
    ('2026-08-17', 'Hari Kemerdekaan RI ke-81', 'national', TRUE),
    ('2026-09-05', 'Maulid Nabi Muhammad SAW', 'national', TRUE),
    ('2026-12-25', 'Hari Raya Natal', 'national', TRUE),
    ('2026-12-26', 'Cuti Bersama Natal', 'joint', FALSE);

-- Indexes
CREATE INDEX idx_reprimands_employee_id ON reprimands(employee_id);
CREATE INDEX idx_reprimands_status ON reprimands(status);
CREATE INDEX idx_reprimands_issued_by ON reprimands(issued_by);
CREATE INDEX idx_announcements_type ON announcements(announcement_type);
CREATE INDEX idx_announcements_published_at ON announcements(published_at DESC);
CREATE INDEX idx_announcements_pinned ON announcements(is_pinned) WHERE is_pinned = TRUE;
CREATE INDEX idx_company_holidays_date ON company_holidays(date);
CREATE INDEX idx_company_holidays_type ON company_holidays(holiday_type);

-- +goose Down
DROP TABLE IF EXISTS company_holidays;
DROP TABLE IF EXISTS announcement_reads;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS reprimands;
