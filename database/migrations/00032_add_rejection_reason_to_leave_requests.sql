-- +goose Up
-- ============================================================
-- Migration 00032: Add rejection_reason to leave_requests
-- ============================================================
-- Bug: UpdateLeaveStatus di leave_repo.go mencoba SET rejection_reason = $3
-- tapi kolom tersebut tidak ada di tabel leave_requests (migration 00007).
-- Ini menyebabkan error saat approve/reject cuti:
--   ERROR: column "rejection_reason" of relation "leave_requests" does not exist
-- ============================================================

ALTER TABLE leave_requests
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES employees(id);

-- +goose Down
ALTER TABLE leave_requests
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS rejected_at,
    DROP COLUMN IF EXISTS rejected_by;
