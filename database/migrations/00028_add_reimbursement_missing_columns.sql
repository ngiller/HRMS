-- +goose Up
-- ============================================================
-- Migration 00028: Add missing columns to reimbursements
-- ============================================================
-- Menambahkan kolom yang hilang dari tabel reimbursements:
-- 1. rejection_reason — alasan penolakan
-- 2. cancelled_by — siapa yang membatalkan
-- 3. cancelled_at — kapan dibatalkan
-- ============================================================

ALTER TABLE reimbursements
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS cancelled_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE reimbursements
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS cancelled_by,
    DROP COLUMN IF EXISTS rejection_reason;
