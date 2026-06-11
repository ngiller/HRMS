-- +goose Up
-- ============================================================
-- Migration 00020: Shift Change Requests
-- ============================================================
-- Fitur untuk karyawan mengajukan perubahan shift melalui
-- aplikasi mobile dengan sistem approval.
--
-- 2 Jenis Request:
-- 1. Individual: Karyawan minta ganti shift untuk tanggal tertentu
-- 2. Swap: 2 karyawan saling tukar jadwal shift
-- ============================================================

-- ============================================================
-- ENUM: shift_change_status
-- ============================================================
DO $$ BEGIN
    CREATE TYPE shift_change_status AS ENUM (
        'pending',            -- Menunggu approval atasan (individual) atau konfirmasi partner (swap)
        'partner_pending',    -- Khusus swap: menunggu konfirmasi dari karyawan partner
        'approved',           -- Disetujui
        'rejected',           -- Ditolak
        'cancelled'           -- Dibatalkan oleh pengaju
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- ENUM: shift_change_type
-- ============================================================
DO $$ BEGIN
    CREATE TYPE shift_change_type AS ENUM (
        'individual',   -- Ganti shift individu
        'swap'          -- Tukar shift dengan karyawan lain
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================
-- Table: shift_change_requests
-- ============================================================
CREATE TABLE shift_change_requests (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_type            shift_change_type NOT NULL,

    -- Pengaju
    employee_id             UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,

    -- Shift yang diminta
    target_date             DATE NOT NULL,
    current_schedule_id     UUID REFERENCES work_schedules(id) ON DELETE SET NULL,
    requested_schedule_id   UUID NOT NULL REFERENCES work_schedules(id) ON DELETE RESTRICT,

    -- Untuk Swap (opsional)
    swap_partner_id         UUID REFERENCES employees(id) ON DELETE SET NULL,
    swap_partner_date       DATE,                           -- Tanggal shift partner (jika berbeda)
    swap_partner_schedule_id UUID REFERENCES work_schedules(id) ON DELETE SET NULL,

    -- Alasan
    reason                  TEXT NOT NULL,

    -- Konfirmasi Partner (khusus swap)
    swap_partner_confirmed      BOOLEAN DEFAULT FALSE,
    swap_partner_confirmed_at   TIMESTAMPTZ,

    -- Status & Approval
    status                  shift_change_status NOT NULL DEFAULT 'pending',
    approval_trail          JSONB DEFAULT '[]'::jsonb,      -- [{level, approver_id, status, note, date}]
    rejection_reason        TEXT,
    cancelled_by            UUID REFERENCES employees(id) ON DELETE SET NULL,
    cancelled_at            TIMESTAMPTZ,

    -- Timestamps
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

-- Trigger untuk auto-update updated_at
CREATE TRIGGER set_shift_change_requests_updated_at
    BEFORE UPDATE ON shift_change_requests
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Indexes
-- ============================================================
CREATE INDEX idx_shift_change_employee_id ON shift_change_requests(employee_id);
CREATE INDEX idx_shift_change_swap_partner ON shift_change_requests(swap_partner_id);
CREATE INDEX idx_shift_change_status ON shift_change_requests(status);
CREATE INDEX idx_shift_change_target_date ON shift_change_requests(target_date);
CREATE INDEX idx_shift_change_employee_date ON shift_change_requests(employee_id, target_date);

-- Prevent duplicate pending requests for same employee + date
CREATE UNIQUE INDEX idx_shift_change_unique_pending
    ON shift_change_requests(employee_id, target_date)
    WHERE status IN ('pending', 'partner_pending');

-- +goose Down
DROP TABLE IF EXISTS shift_change_requests;
DROP TYPE IF EXISTS shift_change_type;
DROP TYPE IF EXISTS shift_change_status;
