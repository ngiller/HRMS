-- +goose Up
-- ============================================================
-- Migration 00033: Add missing columns to overtime_requests
-- ============================================================
-- Bug: migration 00008 (overtime_requests) tidak memiliki
-- rejection_reason, cancelled_by, dan cancelled_at.
-- Tapi GetOvertimeRequestByID di repository merujuk kolom-kolom
-- ini, menyebabkan error di endpoint GET /api/overtime-requests/:id.
-- ============================================================

ALTER TABLE overtime_requests
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES employees(id),
    ADD COLUMN IF NOT EXISTS cancelled_by UUID,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE overtime_requests
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS rejected_at,
    DROP COLUMN IF EXISTS rejected_by,
    DROP COLUMN IF EXISTS cancelled_by,
    DROP COLUMN IF EXISTS cancelled_at;
