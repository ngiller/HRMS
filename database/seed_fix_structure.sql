-- ============================================================
-- FIX: Struktur Organisasi Lengkap
-- Tambah data dummy + perbaiki posisi yang salah
-- ============================================================

-- Matiin trigger audit dulu
ALTER TABLE employees DISABLE TRIGGER audit_employees;
ALTER TABLE departments DISABLE TRIGGER audit_departments;
ALTER TABLE positions DISABLE TRIGGER audit_positions;

-- ============================================================
-- 1. TAMBAH POSISI YANG KURANG
-- ============================================================
INSERT INTO positions (name, department_id)
SELECT 'Finance Director', d.id FROM departments d
WHERE d.code = 'FIN'
AND NOT EXISTS (SELECT 1 FROM positions WHERE name = 'Finance Director' AND deleted_at IS NULL);

INSERT INTO positions (name, department_id)
SELECT 'Sales Director', d.id FROM departments d
WHERE d.code = 'SALES'
AND NOT EXISTS (SELECT 1 FROM positions WHERE name = 'Sales Director' AND deleted_at IS NULL);

INSERT INTO positions (name, department_id)
SELECT 'Marketing Director', d.id FROM departments d
WHERE d.code = 'MKT'
AND NOT EXISTS (SELECT 1 FROM positions WHERE name = 'Marketing Director' AND deleted_at IS NULL);

-- ============================================================
-- 2. UPDATE EMP-009 (Asep Harian → Manager IT)
-- ============================================================
UPDATE employees e
SET position_id = (SELECT id FROM positions WHERE name = 'Manager IT' LIMIT 1),
    department_id = (SELECT id FROM departments WHERE code = 'IT' LIMIT 1),
    role_id = (SELECT id FROM roles WHERE slug = 'manager' LIMIT 1),
    employee_id = 'EMP-009'
WHERE employee_id = 'EMP-009' OR (email = 'bambang@company.com' AND employee_id LIKE 'EMP-%');

-- Update nama kalo masih Asep Harian
UPDATE employees SET full_name = 'Bambang Supriyadi'
WHERE employee_id = 'EMP-009' AND full_name = 'Asep Harian';

-- ============================================================
-- 3. UPDATE EMP-004 (Dewi Lestari → HR Manager)
-- ============================================================
UPDATE employees
SET position_id = (SELECT id FROM positions WHERE name = 'HR Manager' LIMIT 1)
WHERE employee_id = 'EMP-004';

-- ============================================================
-- 4. INSERT NEW DIRECTORS (FINANCE, SALES, MARKETING)
-- ============================================================

-- Finance Director
INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active,
    role_id, position_id, department_id, work_schedule_id)
SELECT 'EMP-020', 'Fransiskus Simorangkir', 'frans@company.com',
    (SELECT password_hash FROM employees WHERE employee_id = 'EMP-001' LIMIT 1),
    'laki_laki', '2023-01-01', 'tetap', TRUE,
    (SELECT id FROM roles WHERE slug = 'director' LIMIT 1),
    (SELECT id FROM positions WHERE name = 'Finance Director' LIMIT 1),
    (SELECT id FROM departments WHERE code = 'FIN' LIMIT 1),
    (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'EMP-020');

-- Sales Director
INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active,
    role_id, position_id, department_id, work_schedule_id)
SELECT 'EMP-021', 'Hendra Permana', 'hendra.permana@company.com',
    (SELECT password_hash FROM employees WHERE employee_id = 'EMP-001' LIMIT 1),
    'laki_laki', '2023-06-01', 'tetap', TRUE,
    (SELECT id FROM roles WHERE slug = 'director' LIMIT 1),
    (SELECT id FROM positions WHERE name = 'Sales Director' LIMIT 1),
    (SELECT id FROM departments WHERE code = 'SALES' LIMIT 1),
    (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'EMP-021');

-- Marketing Director
INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active,
    role_id, position_id, department_id, work_schedule_id)
SELECT 'EMP-022', 'Vina Oktaviani', 'vina@company.com',
    (SELECT password_hash FROM employees WHERE employee_id = 'EMP-001' LIMIT 1),
    'perempuan', '2023-03-01', 'tetap', TRUE,
    (SELECT id FROM roles WHERE slug = 'director' LIMIT 1),
    (SELECT id FROM positions WHERE name = 'Marketing Director' LIMIT 1),
    (SELECT id FROM departments WHERE code = 'MKT' LIMIT 1),
    (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'EMP-022');

-- ============================================================
-- 4b. SET POSITION_ID UNTUK DIRECTORS (kalo posisi duluan ga ada)
-- ============================================================
UPDATE employees SET position_id = (SELECT id FROM positions WHERE name = 'Finance Director' LIMIT 1)
WHERE employee_id = 'EMP-020' AND (position_id IS NULL OR position_id NOT IN (SELECT id FROM positions WHERE name = 'Finance Director'));

UPDATE employees SET position_id = (SELECT id FROM positions WHERE name = 'Sales Director' LIMIT 1)
WHERE employee_id = 'EMP-021' AND (position_id IS NULL OR position_id NOT IN (SELECT id FROM positions WHERE name = 'Sales Director'));

UPDATE employees SET position_id = (SELECT id FROM positions WHERE name = 'Marketing Director' LIMIT 1)
WHERE employee_id = 'EMP-022' AND (position_id IS NULL OR position_id NOT IN (SELECT id FROM positions WHERE name = 'Marketing Director'));

-- ============================================================
-- 5. TAMBAH BEBERAPA STAFF LAGI BIAR LENGKAP
-- ============================================================

-- Staff IT tambahan
INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active,
    role_id, position_id, department_id, work_schedule_id)
SELECT 'EMP-023', 'Gilang Permadi', 'gilang@company.com',
    (SELECT password_hash FROM employees WHERE employee_id = 'EMP-001' LIMIT 1),
    'laki_laki', '2024-02-01', 'tetap', TRUE,
    (SELECT id FROM roles WHERE slug = 'employee' LIMIT 1),
    (SELECT id FROM positions WHERE name = 'Staff IT' LIMIT 1),
    (SELECT id FROM departments WHERE code = 'IT' LIMIT 1),
    (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'EMP-023');

-- Finance Staff tambahan
INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active,
    role_id, position_id, department_id, work_schedule_id)
SELECT 'EMP-024', 'Dian Puspita', 'dian@company.com',
    (SELECT password_hash FROM employees WHERE employee_id = 'EMP-001' LIMIT 1),
    'perempuan', '2024-05-01', 'tetap', TRUE,
    (SELECT id FROM roles WHERE slug = 'employee' LIMIT 1),
    (SELECT id FROM positions WHERE name = 'Finance Staff' LIMIT 1),
    (SELECT id FROM departments WHERE code = 'FIN' LIMIT 1),
    (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'EMP-024');

-- Sales tambahan
INSERT INTO employees (employee_id, full_name, email, password_hash, gender, join_date, employment_status, is_active,
    role_id, position_id, department_id, work_schedule_id)
SELECT 'EMP-025', 'Rizky Pratama', 'rizky@company.com',
    (SELECT password_hash FROM employees WHERE employee_id = 'EMP-001' LIMIT 1),
    'laki_laki', '2025-01-15', 'kontrak', TRUE,
    (SELECT id FROM roles WHERE slug = 'employee' LIMIT 1),
    (SELECT id FROM positions WHERE name = 'Sales Executive' LIMIT 1),
    (SELECT id FROM departments WHERE code = 'SALES' LIMIT 1),
    (SELECT id FROM work_schedules WHERE name = '5 Hari Kerja (Senin-Jumat)' AND deleted_at IS NULL LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM employees WHERE employee_id = 'EMP-025');

-- ============================================================
-- 6. SET APPROVAL LINES (ATASAN)
-- ============================================================

-- IT: Budi → Bambang, Rudi → Bambang, Gilang → Bambang, Bambang → Tito
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-009' LIMIT 1) WHERE employee_id = 'EMP-001';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-009' LIMIT 1) WHERE employee_id = 'EMP-005';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-009' LIMIT 1) WHERE employee_id = 'EMP-023';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-011' LIMIT 1) WHERE employee_id = 'EMP-009';

-- HR: Dewi Sartika → Dewi Lestari, Dewi Lestari → Agus
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-004' LIMIT 1) WHERE employee_id = 'EMP-010';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-013' LIMIT 1) WHERE employee_id = 'EMP-004';

-- FINANCE: Siti → Sri, Hendra → Sri, Dian → Sri, Sri → Frans
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-012' LIMIT 1) WHERE employee_id = 'EMP-002';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-012' LIMIT 1) WHERE employee_id = 'EMP-008';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-012' LIMIT 1) WHERE employee_id = 'EMP-024';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-020' LIMIT 1) WHERE employee_id = 'EMP-012';

-- SALES: Andi → Maya, Ahmad → Maya, Rizky → Maya, Maya → Hendra Permana
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-014' LIMIT 1) WHERE employee_id = 'EMP-003';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-014' LIMIT 1) WHERE employee_id = 'EMP-006';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-014' LIMIT 1) WHERE employee_id = 'EMP-025';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-021' LIMIT 1) WHERE employee_id = 'EMP-014';

-- MARKETING: Rina → Dedi, Dedi → Vina
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-015' LIMIT 1) WHERE employee_id = 'EMP-007';
UPDATE employees SET approval_line_id = (SELECT id FROM employees WHERE employee_id = 'EMP-022' LIMIT 1) WHERE employee_id = 'EMP-015';

-- ============================================================
-- 7. SET DEPARTMENT HEADS
-- ============================================================
UPDATE departments SET head_id = (SELECT id FROM employees WHERE employee_id = 'EMP-011' LIMIT 1) WHERE code = 'IT';
UPDATE departments SET head_id = (SELECT id FROM employees WHERE employee_id = 'EMP-020' LIMIT 1) WHERE code = 'FIN';
UPDATE departments SET head_id = (SELECT id FROM employees WHERE employee_id = 'EMP-013' LIMIT 1) WHERE code = 'HR';
UPDATE departments SET head_id = (SELECT id FROM employees WHERE employee_id = 'EMP-021' LIMIT 1) WHERE code = 'SALES';
UPDATE departments SET head_id = (SELECT id FROM employees WHERE employee_id = 'EMP-022' LIMIT 1) WHERE code = 'MKT';

-- ============================================================
-- 8. SEED SALARY HISTORIES UNTUK KARYAWAN BARU
-- ============================================================
INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
SELECT id, 25000000, 0, CURRENT_DATE, 'Initial Seed - Finance Director'
FROM employees WHERE employee_id = 'EMP-020'
AND NOT EXISTS (SELECT 1 FROM employee_salary_histories WHERE employee_id = (SELECT id FROM employees WHERE employee_id = 'EMP-020'));

INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
SELECT id, 23000000, 0, CURRENT_DATE, 'Initial Seed - Sales Director'
FROM employees WHERE employee_id = 'EMP-021'
AND NOT EXISTS (SELECT 1 FROM employee_salary_histories WHERE employee_id = (SELECT id FROM employees WHERE employee_id = 'EMP-021'));

INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
SELECT id, 22000000, 0, CURRENT_DATE, 'Initial Seed - Marketing Director'
FROM employees WHERE employee_id = 'EMP-022'
AND NOT EXISTS (SELECT 1 FROM employee_salary_histories WHERE employee_id = (SELECT id FROM employees WHERE employee_id = 'EMP-022'));

INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
SELECT id, 7500000, 0, CURRENT_DATE, 'Initial Seed - Staff IT'
FROM employees WHERE employee_id = 'EMP-023'
AND NOT EXISTS (SELECT 1 FROM employee_salary_histories WHERE employee_id = (SELECT id FROM employees WHERE employee_id = 'EMP-023'));

INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
SELECT id, 7000000, 0, CURRENT_DATE, 'Initial Seed - Finance Staff'
FROM employees WHERE employee_id = 'EMP-024'
AND NOT EXISTS (SELECT 1 FROM employee_salary_histories WHERE employee_id = (SELECT id FROM employees WHERE employee_id = 'EMP-024'));

INSERT INTO employee_salary_histories (employee_id, base_salary, daily_wage, effective_date, reason)
SELECT id, 5500000, 0, CURRENT_DATE, 'Initial Seed - Sales Executive'
FROM employees WHERE employee_id = 'EMP-025'
AND NOT EXISTS (SELECT 1 FROM employee_salary_histories WHERE employee_id = (SELECT id FROM employees WHERE employee_id = 'EMP-025'));

-- ============================================================
-- 9. HAPUS DUPLICATE DEPARTEMEN ANEH (contoh, fdsa)
-- ============================================================
-- Soft delete weird departments that might exist from testing
UPDATE departments SET deleted_at = NOW() WHERE code IN ('CTG', 'CTH', 'FDSA');

-- ============================================================
-- RE-ENABLE TRIGGERS
-- ============================================================
ALTER TABLE employees ENABLE TRIGGER audit_employees;
ALTER TABLE departments ENABLE TRIGGER audit_departments;
ALTER TABLE positions ENABLE TRIGGER audit_positions;

-- ============================================================
-- VERIFICATION
-- ============================================================
SELECT e.employee_id, e.full_name, p.name as position_name, d.name as dept_name, r.slug as role_slug,
    (SELECT full_name FROM employees WHERE id = e.approval_line_id) as atasan
FROM employees e
LEFT JOIN positions p ON p.id = e.position_id
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN roles r ON r.id = e.role_id
WHERE e.deleted_at IS NULL AND e.employee_id IN ('EMP-001','EMP-004','EMP-005','EMP-009','EMP-010','EMP-011','EMP-012','EMP-013','EMP-014','EMP-015','EMP-020','EMP-021','EMP-022','EMP-023','EMP-024','EMP-025')
ORDER BY d.name, p.name;
