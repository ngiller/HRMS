-- +goose Up
-- ============================================================
-- Migration 00015: Final Indexes, Triggers & Functions
-- ============================================================

-- ============================================================
-- NOTE: trigger_set_updated_at() sudah didefinisikan di
-- migration 00001. Tidak perlu dibuat ulang di sini.
-- ============================================================

-- ============================================================
-- Function: Calculate payroll (to be called by application)
-- ============================================================
CREATE OR REPLACE FUNCTION calculate_employee_payroll(
    p_payroll_period_id UUID,
    p_employee_id UUID,
    p_base_salary DECIMAL,
    p_daily_wage DECIMAL DEFAULT 0,
    p_total_days_worked INTEGER DEFAULT 0,
    p_allowances JSONB DEFAULT '[]'::jsonb,
    p_overtime_pay DECIMAL DEFAULT 0,
    p_thr_amount DECIMAL DEFAULT 0,
    p_bonus_amount DECIMAL DEFAULT 0,
    p_loan_deduction DECIMAL DEFAULT 0,
    p_other_deductions DECIMAL DEFAULT 0
)
RETURNS UUID AS $$
DECLARE
    v_gross_salary DECIMAL(15,2);
    v_total_allowances DECIMAL(15,2);
    v_bpjs_kesehatan DECIMAL(15,2);
    v_bpjs_jht DECIMAL(15,2);
    v_bpjs_jp DECIMAL(15,2);
    v_bpjs_ks_company DECIMAL(15,2);
    v_bpjs_jht_company DECIMAL(15,2);
    v_bpjs_jp_company DECIMAL(15,2);
    v_bpjs_jkk DECIMAL(15,2);
    v_bpjs_jkm DECIMAL(15,2);
    v_total_deductions DECIMAL(15,2);
    v_net_salary DECIMAL(15,2);
    v_company_cost JSONB;
    v_pph21 DECIMAL(15,2);
    v_jkk_rate DECIMAL(5,2);
    v_payroll_item_id UUID;
    v_bpjs_ceiling DECIMAL(15,2);
    v_bpjs_jp_ceiling DECIMAL(15,2) := 10000000;
BEGIN
    -- Hitung total tunjangan
    SELECT COALESCE(SUM((value->>'amount')::DECIMAL), 0)
    INTO v_total_allowances
    FROM jsonb_array_elements(p_allowances);
    
    -- Gross Salary
    v_gross_salary := p_base_salary + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;
    
    -- Get BPJS config from company settings FIRST (before using the values)
    SELECT
        COALESCE(hr_settings->>'bpjs_kesehatan_ceiling', '12000000')::DECIMAL,
        COALESCE(hr_settings->>'bpjs_jp_ceiling', '10000000')::DECIMAL,
        bpjs_jkk_rate
    INTO v_bpjs_ceiling, v_bpjs_jp_ceiling, v_jkk_rate
    FROM companies LIMIT 1;
    
    -- Default values if no company config found
    IF v_bpjs_ceiling IS NULL THEN v_bpjs_ceiling := 12000000; END IF;
    IF v_bpjs_jp_ceiling IS NULL THEN v_bpjs_jp_ceiling := 10000000; END IF;
    IF v_jkk_rate IS NULL THEN v_jkk_rate := 0.54; END IF;
    
    -- BPJS (Pekerja) - menggunakan LEAST(gaji, ceiling)
    v_bpjs_kesehatan := LEAST(v_gross_salary, v_bpjs_ceiling) * 0.01;
    v_bpjs_jht := v_gross_salary * 0.02;
    v_bpjs_jp := LEAST(v_gross_salary, v_bpjs_jp_ceiling) * 0.01;
    
    -- BPJS (Perusahaan)
    v_bpjs_ks_company := LEAST(v_gross_salary, v_bpjs_ceiling) * 0.04;
    v_bpjs_jht_company := v_gross_salary * 0.037;
    v_bpjs_jp_company := LEAST(v_gross_salary, v_bpjs_jp_ceiling) * 0.02;
    v_bpjs_jkk := v_gross_salary * (v_jkk_rate / 100);
    v_bpjs_jkm := v_gross_salary * 0.003;
    
    -- PPh 21 placeholder (dihitung by application dengan TER)
    v_pph21 := 0;
    
    -- Total Deductions
    v_total_deductions := v_bpjs_kesehatan + v_bpjs_jht + v_bpjs_jp + v_pph21 + p_loan_deduction + p_other_deductions;
    
    -- Net Salary
    v_net_salary := v_gross_salary - v_total_deductions;
    
    -- Company cost
    v_company_cost := jsonb_build_array(
        jsonb_build_object('name', 'BPJS Kesehatan', 'amount', v_bpjs_ks_company),
        jsonb_build_object('name', 'BPJS JHT', 'amount', v_bpjs_jht_company),
        jsonb_build_object('name', 'BPJS JP', 'amount', v_bpjs_jp_company),
        jsonb_build_object('name', 'BPJS JKK', 'amount', v_bpjs_jkk),
        jsonb_build_object('name', 'BPJS JKM', 'amount', v_bpjs_jkm)
    );
    
    -- Upsert payroll item
    INSERT INTO payroll_items (
        payroll_period_id, employee_id,
        base_salary, daily_wage, total_days_worked,
        allowances, overtime_pay, thr_amount, bonus_amount,
        gross_salary,
        deductions, pph21_amount, bpjs_kesehatan, bpjs_jht, bpjs_jp,
        loan_deduction, other_deductions, total_deductions,
        net_salary,
        company_cost,
        bpjs_kesehatan_company, bpjs_jht_company, bpjs_jp_company, bpjs_jkk, bpjs_jkm,
        status
    ) VALUES (
        p_payroll_period_id, p_employee_id,
        p_base_salary, p_daily_wage, p_total_days_worked,
        p_allowances, p_overtime_pay, p_thr_amount, p_bonus_amount,
        v_gross_salary,
        jsonb_build_array(
            jsonb_build_object('name', 'PPh 21', 'amount', v_pph21),
            jsonb_build_object('name', 'BPJS Kesehatan', 'amount', v_bpjs_kesehatan),
            jsonb_build_object('name', 'BPJS JHT', 'amount', v_bpjs_jht),
            jsonb_build_object('name', 'BPJS JP', 'amount', v_bpjs_jp),
            jsonb_build_object('name', 'Pinjaman', 'amount', p_loan_deduction),
            jsonb_build_object('name', 'Lain-lain', 'amount', p_other_deductions)
        ),
        v_pph21, v_bpjs_kesehatan, v_bpjs_jht, v_bpjs_jp,
        p_loan_deduction, p_other_deductions, v_total_deductions,
        v_net_salary,
        v_company_cost,
        v_bpjs_ks_company, v_bpjs_jht_company, v_bpjs_jp_company, v_bpjs_jkk, v_bpjs_jkm,
        'calculated'
    )
    ON CONFLICT (payroll_period_id, employee_id) 
    DO UPDATE SET
        base_salary = EXCLUDED.base_salary,
        gross_salary = EXCLUDED.gross_salary,
        net_salary = EXCLUDED.net_salary,
        total_deductions = EXCLUDED.total_deductions,
        updated_at = NOW()
    RETURNING id INTO v_payroll_item_id;
    
    RETURN v_payroll_item_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Additional Composite Indexes for Query Performance
-- ============================================================

-- Attendance queries: by employee + month
CREATE INDEX idx_attendance_records_employee_month_status
    ON attendance_records(employee_id, (DATE_TRUNC('month', date)), status);

-- Leave queries: overlapping date check
CREATE INDEX idx_leave_requests_date_range
    ON leave_requests(employee_id, start_date, end_date);

-- Payroll queries
CREATE INDEX idx_payroll_items_employee_period
    ON payroll_items(employee_id, payroll_period_id);

-- Notification queries
CREATE INDEX idx_notifications_user_created
    ON notifications(user_id, created_at DESC);

-- Employee search
CREATE INDEX idx_employees_name_search
    ON employees USING gin(to_tsvector('simple', full_name));

-- Activity logs by date range
CREATE INDEX idx_activity_logs_date_range
    ON activity_logs(created_at DESC);

-- ============================================================
-- Views for Reporting
-- ============================================================

-- Employee Headcount View
CREATE OR REPLACE VIEW v_employee_headcount AS
SELECT
    d.id AS department_id,
    d.name AS department_name,
    COUNT(e.id) AS total_employees,
    COUNT(e.id) FILTER (WHERE e.employment_status = 'tetap') AS tetap,
    COUNT(e.id) FILTER (WHERE e.employment_status = 'kontrak') AS kontrak,
    COUNT(e.id) FILTER (WHERE e.employment_status = 'percobaan') AS percobaan,
    COUNT(e.id) FILTER (WHERE e.employment_status = 'harian') AS harian,
    COUNT(e.id) FILTER (WHERE e.gender = 'laki_laki') AS laki_laki,
    COUNT(e.id) FILTER (WHERE e.gender = 'perempuan') AS perempuan
FROM departments d
LEFT JOIN employees e ON e.department_id = d.id AND e.deleted_at IS NULL AND e.is_active = TRUE
WHERE d.deleted_at IS NULL
GROUP BY d.id, d.name
ORDER BY d.name;

-- Attendance Monthly Report View
CREATE OR REPLACE VIEW v_attendance_monthly AS
SELECT
    e.id AS employee_id,
    e.employee_id AS nip,
    e.full_name,
    d.name AS department_name,
    DATE_TRUNC('month', ar.date) AS month,
    COUNT(*) AS total_days,
    COUNT(*) FILTER (WHERE ar.status = 'hadir') AS hadir,
    COUNT(*) FILTER (WHERE ar.status = 'terlambat') AS terlambat,
    COUNT(*) FILTER (WHERE ar.status = 'sakit') AS sakit,
    COUNT(*) FILTER (WHERE ar.status = 'izin') AS izin,
    COUNT(*) FILTER (WHERE ar.status = 'tanpa_keterangan') AS alpa,
    ROUND(AVG(ar.total_work_hours), 2) AS avg_work_hours,
    ROUND(AVG(ar.late_minutes) FILTER (WHERE ar.is_late), 1) AS avg_late_minutes
FROM employees e
JOIN attendance_records ar ON ar.employee_id = e.id
LEFT JOIN departments d ON d.id = e.department_id
WHERE e.deleted_at IS NULL AND ar.deleted_at IS NULL
GROUP BY e.id, e.employee_id, e.full_name, d.name, DATE_TRUNC('month', ar.date);

-- Leave Balance Summary View
CREATE OR REPLACE VIEW v_leave_balance_summary AS
SELECT
    e.id AS employee_id,
    e.employee_id AS nip,
    e.full_name,
    lt.name AS leave_type_name,
    lb.year,
    lb.total_quota,
    lb.used,
    lb.remaining,
    lb.rolled_over_from
FROM employees e
JOIN leave_balances lb ON lb.employee_id = e.id
JOIN leave_types lt ON lt.id = lb.leave_type_id
WHERE e.deleted_at IS NULL AND e.is_active = TRUE
ORDER BY e.full_name, lt.sort_order;

-- ============================================================
-- Cleanup: Update departments.head_id FK after employees exist
-- Note: This must be run after employees table is populated
-- ============================================================
-- Add FK constraint for departments.head_id
ALTER TABLE departments
    ADD CONSTRAINT fk_departments_head
    FOREIGN KEY (head_id) REFERENCES employees(id) ON DELETE SET NULL;

-- ============================================================
-- Function: Auto-create leave balances at the start of each year
-- ============================================================
CREATE OR REPLACE FUNCTION auto_create_leave_balances()
RETURNS VOID AS $$
DECLARE
    v_year INTEGER := EXTRACT(YEAR FROM CURRENT_DATE);
    v_employee RECORD;
    v_leave_type RECORD;
BEGIN
    FOR v_employee IN SELECT id FROM employees WHERE is_active = TRUE AND deleted_at IS NULL LOOP
        FOR v_leave_type IN SELECT id, default_quota FROM leave_types WHERE is_active = TRUE AND default_quota > 0 LOOP
            INSERT INTO leave_balances (employee_id, leave_type_id, year, total_quota, used, remaining)
            VALUES (v_employee.id, v_leave_type.id, v_year, v_leave_type.default_quota, 0, v_leave_type.default_quota)
            ON CONFLICT (employee_id, leave_type_id, year) DO NOTHING;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- +goose Down
DROP VIEW IF EXISTS v_leave_balance_summary;
DROP VIEW IF EXISTS v_attendance_monthly;
DROP VIEW IF EXISTS v_employee_headcount;
DROP FUNCTION IF EXISTS auto_create_leave_balances;
DROP FUNCTION IF EXISTS calculate_employee_payroll;
DROP FUNCTION IF EXISTS trigger_set_updated_at;
