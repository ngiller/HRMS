-- +goose Up
-- ============================================================
-- Migration 00037: Add salary range columns to position_grades
-- ============================================================
-- Menambahkan kolom min_salary dan max_salary ke tabel
-- position_grades agar setiap golongan jabatan memiliki
-- rentang gaji yang jelas. Ini penting untuk:
-- - Validasi gaji saat input/edit karyawan
-- - Payroll budgeting
-- - Struktur penggajian yang transparan
-- ============================================================

ALTER TABLE position_grades
    ADD COLUMN IF NOT EXISTS min_salary DECIMAL(15,2),
    ADD COLUMN IF NOT EXISTS max_salary DECIMAL(15,2);

-- Update existing position grades with sensible salary ranges
UPDATE position_grades SET 
    min_salary = CASE level
        WHEN 1  THEN 4000000   -- Staff
        WHEN 2  THEN 6000000   -- Senior Staff
        WHEN 3  THEN 9000000   -- Supervisor
        WHEN 4  THEN 12000000  -- Senior Supervisor
        WHEN 5  THEN 15000000  -- Assistant Manager
        WHEN 6  THEN 18000000  -- Manager
        WHEN 7  THEN 25000000  -- Senior Manager
        WHEN 8  THEN 35000000  -- General Manager
        WHEN 9  THEN 50000000  -- Director
        WHEN 10 THEN 75000000  -- President Director
        ELSE 0
    END,
    max_salary = CASE level
        WHEN 1  THEN 7000000   -- Staff
        WHEN 2  THEN 10000000  -- Senior Staff
        WHEN 3  THEN 14000000  -- Supervisor
        WHEN 4  THEN 18000000  -- Senior Supervisor
        WHEN 5  THEN 25000000  -- Assistant Manager
        WHEN 6  THEN 30000000  -- Manager
        WHEN 7  THEN 40000000  -- Senior Manager
        WHEN 8  THEN 55000000  -- General Manager
        WHEN 9  THEN 80000000  -- Director
        WHEN 10 THEN 150000000 -- President Director
        ELSE 0
    END
WHERE min_salary IS NULL AND max_salary IS NULL;

-- +goose Down
ALTER TABLE position_grades
    DROP COLUMN IF EXISTS min_salary,
    DROP COLUMN IF EXISTS max_salary;
