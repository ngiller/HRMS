-- +goose Up
-- ============================================================
-- Migration 00027: Add base_salary column to employees
-- ============================================================
-- Sebelumnya: base_salary hanya disimpan di employee_salary_histories.
-- Akibatnya:
--   - Tidak bisa lihat gaji pokok di halaman daftar karyawan
--   - API create/update karyawan tidak bisa set base_salary
--
-- Sesudah: Setiap karyawan punya base_salary langsung di tabel employees.
-- employee_salary_histories tetap sebagai riwayat perubahan.
-- ============================================================

-- 1. Tambah kolom base_salary
ALTER TABLE employees
    ADD COLUMN base_salary DECIMAL(15,2) DEFAULT NULL;

COMMENT ON COLUMN employees.base_salary IS
'Gaji pokok karyawan saat ini. Riwayat perubahan disimpan di employee_salary_histories.';

-- 2. Update kolom daily_wage — pastikan ada comment
COMMENT ON COLUMN employees.daily_wage IS
'Upah harian untuk karyawan dengan status harian. Riwayat perubahan disimpan di employee_salary_histories.';

-- 3. Disable audit triggers temporarily (running outside application context)
ALTER TABLE employees DISABLE TRIGGER audit_employees;
ALTER TABLE employee_salary_components DISABLE TRIGGER audit_employee_salary_components;
ALTER TABLE employee_salary_components DISABLE TRIGGER log_salary_component_changes;

-- 4. Backfill base_salary dari employee_salary_histories
-- Ambil base_salary terbaru per employee (berdasarkan effective_date, lalu created_at)
UPDATE employees e
SET base_salary = sub.latest_salary
FROM (
    SELECT DISTINCT ON (employee_id)
        employee_id,
        base_salary AS latest_salary
    FROM employee_salary_histories
    ORDER BY employee_id, effective_date DESC NULLS LAST, created_at DESC
) sub
WHERE e.id = sub.employee_id;

-- 5. Backfill daily_wage dari employee_salary_histories (jika employees.daily_wage masih NULL)
UPDATE employees e
SET daily_wage = sub.latest_daily_wage
FROM (
    SELECT DISTINCT ON (employee_id)
        employee_id,
        daily_wage AS latest_daily_wage
    FROM employee_salary_histories
    WHERE daily_wage IS NOT NULL AND daily_wage > 0
    ORDER BY employee_id, effective_date DESC NULLS LAST, created_at DESC
) sub
WHERE e.id = sub.employee_id
AND (e.daily_wage IS NULL OR e.daily_wage = 0);

-- 6. Re-enable audit triggers
ALTER TABLE employees ENABLE TRIGGER audit_employees;
ALTER TABLE employee_salary_components ENABLE TRIGGER audit_employee_salary_components;
ALTER TABLE employee_salary_components ENABLE TRIGGER log_salary_component_changes;

-- +goose Down
ALTER TABLE employees DROP COLUMN IF EXISTS base_salary;
