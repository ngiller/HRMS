-- +goose Up
-- ============================================================
-- Migration 00021: Partial Unique Indexes for Soft Delete
-- ============================================================
-- Mengubah UNIQUE constraint menjadi partial unique index
-- agar record yang soft-delete (deleted_at IS NOT NULL)
-- tidak memblokir data baru dengan email/employee_id yang sama.
-- ============================================================

-- Hapus unique constraint email (dari migration 00016)
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_email_unique;

-- Buat partial unique index untuk email — hanya untuk record aktif
CREATE UNIQUE INDEX employees_email_unique_active
    ON employees(email) WHERE deleted_at IS NULL;

-- Hapus unique constraint employee_id (UNIQUE inline dari migration 00005)
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_employee_id_key;

-- Buat partial unique index untuk employee_id — hanya untuk record aktif
CREATE UNIQUE INDEX employees_employee_id_unique_active
    ON employees(employee_id) WHERE deleted_at IS NULL;

-- ============================================================
-- Pastikan auth_repo tetap bisa login efisien
-- (index WHERE deleted_at IS NULL sudah ada dari migration 00016)
-- ============================================================

-- +goose Down
-- ============================================================
-- Rollback migration 00021
-- ============================================================

DROP INDEX IF EXISTS employees_email_unique_active;
DROP INDEX IF EXISTS employees_employee_id_unique_active;

-- Kembalikan unique constraint biasa
ALTER TABLE employees
    ADD CONSTRAINT employees_email_unique UNIQUE (email);

-- employee_id: karena didefinisikan inline di CREATE TABLE,
-- kita tambahkan UNIQUE constraint dengan nama yang sama seperti sebelumnya
ALTER TABLE employees
    ADD CONSTRAINT employees_employee_id_key UNIQUE (employee_id);
