-- +goose Up
-- ============================================================
-- Migration 00025: Add daily_wage column to employees
-- ============================================================
-- Menambahkan kolom daily_wage untuk mendukung karyawan dengan
-- status "harian" (upah per hari).
-- ============================================================

ALTER TABLE employees
    ADD COLUMN daily_wage DECIMAL(15,2) DEFAULT NULL;

-- +goose Down
ALTER TABLE employees DROP COLUMN IF EXISTS daily_wage;
