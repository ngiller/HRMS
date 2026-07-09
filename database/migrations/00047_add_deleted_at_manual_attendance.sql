-- +goose Up
-- ============================================================
-- Migration 00047: Add deleted_at to manual_attendance_requests
-- ============================================================
-- Alasan: Tabel manual_attendance_requests tidak memiliki kolom
-- deleted_at, sehingga LEFT JOIN di entityJoinSQL gagal karena
-- mereferensi kolom yang tidak ada (mar.deleted_at IS NULL).
-- Migration ini menambahkan kolom deleted_at untuk konsistensi
-- dengan tabel entity lainnya (leave_requests, overtime_requests, dll).
-- ============================================================

ALTER TABLE manual_attendance_requests
    ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX idx_manual_attendance_requests_deleted_at
    ON manual_attendance_requests(deleted_at)
    WHERE deleted_at IS NULL;

-- +goose Down
ALTER TABLE manual_attendance_requests
    DROP COLUMN IF EXISTS deleted_at;
