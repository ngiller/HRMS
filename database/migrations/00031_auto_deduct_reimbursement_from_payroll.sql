-- +goose Up
-- ============================================================
-- Migration 00031: Auto-deduct reimbursement from payroll
-- ============================================================
-- Sekarang calculate_employee_payroll() secara otomatis membaca
-- reimbursement yang sudah paid dengan payment_method = 'payroll'
-- dalam periode payroll yang sama, lalu memasukkannya ke dalam
-- komponen potongan gaji (deductions + other_deductions).
--
-- Changes:
-- 1. Query reimbursements yang paid via payroll dalam periode ini
-- 2. Tambah ke other_deductions dan deductions JSONB array
-- 3. Total deductions otomatis bertambah
-- ============================================================

-- +goose StatementBegin
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
    v_payroll_item_id UUID;

    v_kesehatan_emp_rate       DECIMAL;
    v_kesehatan_comp_rate      DECIMAL;
    v_kesehatan_ceiling        DECIMAL(15,2);
    v_jht_emp_rate             DECIMAL;
    v_jht_comp_rate            DECIMAL;
    v_jp_emp_rate              DECIMAL;
    v_jp_comp_rate             DECIMAL;
    v_jp_ceiling               DECIMAL(15,2);
    v_jkm_comp_rate            DECIMAL;
    v_jkk_rate                 DECIMAL(5,2);

    v_bpjs_config              JSONB;
    v_kesehatan_enabled        BOOLEAN := TRUE;
    v_jht_enabled              BOOLEAN := TRUE;
    v_jp_enabled               BOOLEAN := TRUE;
    v_jkk_enabled              BOOLEAN := TRUE;
    v_jkm_enabled              BOOLEAN := TRUE;

    -- NEW: Reimbursement deduction
    v_reimbursement_total      DECIMAL(15,2) := 0;
    v_last_payroll_start       TIMESTAMPTZ := '1970-01-01'::timestamptz;
    v_current_period_start     DATE;
    v_total_other_deductions   DECIMAL(15,2);
BEGIN
    -- Hitung total tunjangan
    SELECT COALESCE(SUM((value->>'amount')::DECIMAL), 0)
    INTO v_total_allowances
    FROM jsonb_array_elements(p_allowances);

    v_gross_salary := p_base_salary + p_daily_wage + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;

    -- ============================================================
    -- REIMBURSEMENT: Auto-deduct paid payroll-method reimbursements
    -- sejak payroll terakhir (hanya period yg end_date <= current period start)
    -- ============================================================
    -- Get current period start date
    SELECT start_date INTO v_current_period_start
    FROM payroll_periods
    WHERE id = p_payroll_period_id;

    -- Cari start_date payroll terakhir (hanya period yg sudah selesai)
    SELECT COALESCE(MAX(pp.start_date), '1970-01-01'::date)::timestamptz
    INTO v_last_payroll_start
    FROM payroll_items pi
    JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
    WHERE pi.employee_id = p_employee_id
      AND pi.status IN ('calculated', 'approved', 'paid')
      AND pp.end_date <= v_current_period_start;

    -- Hitung total reimbursement yang paid via payroll sejak last payroll
    SELECT COALESCE(SUM(r.amount), 0)
    INTO v_reimbursement_total
    FROM reimbursements r
    WHERE r.employee_id = p_employee_id
      AND r.status = 'paid'
      AND r.payment_method = 'payroll'
      AND r.deleted_at IS NULL
      AND r.paid_at IS NOT NULL
      AND r.paid_at > v_last_payroll_start;

    -- Combine passed-in other_deductions with auto-detected reimbursement
    v_total_other_deductions := p_other_deductions + v_reimbursement_total;
    -- ============================================================

    -- BACA BPJS CONFIG DARI COMPANY SETTINGS
    SELECT
        COALESCE((hr_settings #>> '{bpjs,kesehatan,employee_rate}')::DECIMAL, 0.01),
        COALESCE((hr_settings #>> '{bpjs,kesehatan,company_rate}')::DECIMAL, 0.04),
        COALESCE((hr_settings #>> '{bpjs,kesehatan,ceiling}')::DECIMAL, 12000000),
        COALESCE((hr_settings #>> '{bpjs,jht,employee_rate}')::DECIMAL, 0.02),
        COALESCE((hr_settings #>> '{bpjs,jht,company_rate}')::DECIMAL, 0.037),
        COALESCE((hr_settings #>> '{bpjs,jp,employee_rate}')::DECIMAL, 0.01),
        COALESCE((hr_settings #>> '{bpjs,jp,company_rate}')::DECIMAL, 0.02),
        COALESCE((hr_settings #>> '{bpjs,jp,ceiling}')::DECIMAL, 10000000),
        COALESCE((hr_settings #>> '{bpjs,jkm,company_rate}')::DECIMAL, 0.003),
        bpjs_jkk_rate
    INTO
        v_kesehatan_emp_rate, v_kesehatan_comp_rate, v_kesehatan_ceiling,
        v_jht_emp_rate, v_jht_comp_rate,
        v_jp_emp_rate, v_jp_comp_rate, v_jp_ceiling,
        v_jkm_comp_rate,
        v_jkk_rate
    FROM companies LIMIT 1;

    IF v_kesehatan_emp_rate IS NULL THEN v_kesehatan_emp_rate := 0.01; END IF;
    IF v_kesehatan_comp_rate IS NULL THEN v_kesehatan_comp_rate := 0.04; END IF;
    IF v_kesehatan_ceiling IS NULL THEN v_kesehatan_ceiling := 12000000; END IF;
    IF v_jht_emp_rate IS NULL THEN v_jht_emp_rate := 0.02; END IF;
    IF v_jht_comp_rate IS NULL THEN v_jht_comp_rate := 0.037; END IF;
    IF v_jp_emp_rate IS NULL THEN v_jp_emp_rate := 0.01; END IF;
    IF v_jp_comp_rate IS NULL THEN v_jp_comp_rate := 0.02; END IF;
    IF v_jp_ceiling IS NULL THEN v_jp_ceiling := 10000000; END IF;
    IF v_jkm_comp_rate IS NULL THEN v_jkm_comp_rate := 0.003; END IF;
    IF v_jkk_rate IS NULL THEN v_jkk_rate := 0.54; END IF;

    -- BACA PER-EMPLOYEE BPJS OVERRIDE
    SELECT bpjs_config INTO v_bpjs_config
    FROM employees
    WHERE id = p_employee_id;

    IF v_bpjs_config IS NOT NULL THEN
        v_kesehatan_enabled := COALESCE((v_bpjs_config #>> '{kesehatan,enabled}')::BOOLEAN, TRUE);
        v_jht_enabled       := COALESCE((v_bpjs_config #>> '{jht,enabled}')::BOOLEAN, TRUE);
        v_jp_enabled        := COALESCE((v_bpjs_config #>> '{jp,enabled}')::BOOLEAN, TRUE);
        v_jkk_enabled       := COALESCE((v_bpjs_config #>> '{jkk,enabled}')::BOOLEAN, TRUE);
        v_jkm_enabled       := COALESCE((v_bpjs_config #>> '{jkm,enabled}')::BOOLEAN, TRUE);

        IF v_bpjs_config #>> '{kesehatan,employee_rate}' IS NOT NULL THEN
            v_kesehatan_emp_rate := (v_bpjs_config #>> '{kesehatan,employee_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{kesehatan,company_rate}' IS NOT NULL THEN
            v_kesehatan_comp_rate := (v_bpjs_config #>> '{kesehatan,company_rate}')::DECIMAL;
        END IF;

        IF v_bpjs_config #>> '{jht,employee_rate}' IS NOT NULL THEN
            v_jht_emp_rate := (v_bpjs_config #>> '{jht,employee_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jht,company_rate}' IS NOT NULL THEN
            v_jht_comp_rate := (v_bpjs_config #>> '{jht,company_rate}')::DECIMAL;
        END IF;

        IF v_bpjs_config #>> '{jp,employee_rate}' IS NOT NULL THEN
            v_jp_emp_rate := (v_bpjs_config #>> '{jp,employee_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jp,company_rate}' IS NOT NULL THEN
            v_jp_comp_rate := (v_bpjs_config #>> '{jp,company_rate}')::DECIMAL;
        END IF;

        IF v_bpjs_config #>> '{jkm,company_rate}' IS NOT NULL THEN
            v_jkm_comp_rate := (v_bpjs_config #>> '{jkm,company_rate}')::DECIMAL;
        END IF;
    END IF;

    -- HITUNG BPJS (PEKERJA)
    IF v_kesehatan_enabled THEN
        v_bpjs_kesehatan := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_emp_rate;
    ELSE
        v_bpjs_kesehatan := 0;
    END IF;

    IF v_jht_enabled THEN
        v_bpjs_jht := v_gross_salary * v_jht_emp_rate;
    ELSE
        v_bpjs_jht := 0;
    END IF;

    IF v_jp_enabled THEN
        v_bpjs_jp := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_emp_rate;
    ELSE
        v_bpjs_jp := 0;
    END IF;

    -- HITUNG BPJS (PERUSAHAAN)
    IF v_kesehatan_enabled THEN
        v_bpjs_ks_company := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_comp_rate;
    ELSE
        v_bpjs_ks_company := 0;
    END IF;

    IF v_jht_enabled THEN
        v_bpjs_jht_company := v_gross_salary * v_jht_comp_rate;
    ELSE
        v_bpjs_jht_company := 0;
    END IF;

    IF v_jp_enabled THEN
        v_bpjs_jp_company := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_comp_rate;
    ELSE
        v_bpjs_jp_company := 0;
    END IF;

    IF v_jkk_enabled THEN
        v_bpjs_jkk := v_gross_salary * (v_jkk_rate / 100);
    ELSE
        v_bpjs_jkk := 0;
    END IF;

    IF v_jkm_enabled THEN
        v_bpjs_jkm := v_gross_salary * v_jkm_comp_rate;
    ELSE
        v_bpjs_jkm := 0;
    END IF;

    v_pph21 := 0;

    -- Use v_total_other_deductions (which includes reimbursement) instead of p_other_deductions
    v_total_deductions := v_bpjs_kesehatan + v_bpjs_jht + v_bpjs_jp + v_pph21 + p_loan_deduction + v_total_other_deductions;
    v_net_salary := v_gross_salary - v_total_deductions;

    v_company_cost := jsonb_build_array(
        jsonb_build_object('name', 'BPJS Kesehatan', 'amount', v_bpjs_ks_company),
        jsonb_build_object('name', 'BPJS JHT', 'amount', v_bpjs_jht_company),
        jsonb_build_object('name', 'BPJS JP', 'amount', v_bpjs_jp_company),
        jsonb_build_object('name', 'BPJS JKK', 'amount', v_bpjs_jkk),
        jsonb_build_object('name', 'BPJS JKM', 'amount', v_bpjs_jkm)
    );

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
            jsonb_build_object('name', 'Reimbursement', 'amount', v_reimbursement_total),
            jsonb_build_object('name', 'Lain-lain', 'amount', p_other_deductions)
        ),
        v_pph21, v_bpjs_kesehatan, v_bpjs_jht, v_bpjs_jp,
        p_loan_deduction, v_total_other_deductions, v_total_deductions,
        v_net_salary,
        v_company_cost,
        v_bpjs_ks_company, v_bpjs_jht_company, v_bpjs_jp_company, v_bpjs_jkk, v_bpjs_jkm,
        'calculated'
    )
    ON CONFLICT (payroll_period_id, employee_id)
    DO UPDATE SET
        base_salary = EXCLUDED.base_salary,
        daily_wage = EXCLUDED.daily_wage,
        allowances = EXCLUDED.allowances,
        overtime_pay = EXCLUDED.overtime_pay,
        thr_amount = EXCLUDED.thr_amount,
        bonus_amount = EXCLUDED.bonus_amount,
        gross_salary = EXCLUDED.gross_salary,
        deductions = EXCLUDED.deductions,
        pph21_amount = EXCLUDED.pph21_amount,
        bpjs_kesehatan = EXCLUDED.bpjs_kesehatan,
        bpjs_jht = EXCLUDED.bpjs_jht,
        bpjs_jp = EXCLUDED.bpjs_jp,
        loan_deduction = EXCLUDED.loan_deduction,
        other_deductions = EXCLUDED.other_deductions,
        total_deductions = EXCLUDED.total_deductions,
        net_salary = EXCLUDED.net_salary,
        bpjs_kesehatan_company = EXCLUDED.bpjs_kesehatan_company,
        bpjs_jht_company = EXCLUDED.bpjs_jht_company,
        bpjs_jp_company = EXCLUDED.bpjs_jp_company,
        bpjs_jkk = EXCLUDED.bpjs_jkk,
        bpjs_jkm = EXCLUDED.bpjs_jkm,
        company_cost = EXCLUDED.company_cost,
        status = 'calculated',
        updated_at = NOW()
    RETURNING id INTO v_payroll_item_id;

    RETURN v_payroll_item_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
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
    v_payroll_item_id UUID;

    v_kesehatan_emp_rate       DECIMAL;
    v_kesehatan_comp_rate      DECIMAL;
    v_kesehatan_ceiling        DECIMAL(15,2);
    v_jht_emp_rate             DECIMAL;
    v_jht_comp_rate            DECIMAL;
    v_jp_emp_rate              DECIMAL;
    v_jp_comp_rate             DECIMAL;
    v_jp_ceiling               DECIMAL(15,2);
    v_jkm_comp_rate            DECIMAL;
    v_jkk_rate                 DECIMAL(5,2);

    v_bpjs_config              JSONB;
    v_kesehatan_enabled        BOOLEAN := TRUE;
    v_jht_enabled              BOOLEAN := TRUE;
    v_jp_enabled               BOOLEAN := TRUE;
    v_jkk_enabled              BOOLEAN := TRUE;
    v_jkm_enabled              BOOLEAN := TRUE;
BEGIN
    SELECT COALESCE(SUM((value->>'amount')::DECIMAL), 0)
    INTO v_total_allowances
    FROM jsonb_array_elements(p_allowances);

    v_gross_salary := p_base_salary + p_daily_wage + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;

    SELECT
        COALESCE((hr_settings #>> '{bpjs,kesehatan,employee_rate}')::DECIMAL, 0.01),
        COALESCE((hr_settings #>> '{bpjs,kesehatan,company_rate}')::DECIMAL, 0.04),
        COALESCE((hr_settings #>> '{bpjs,kesehatan,ceiling}')::DECIMAL, 12000000),
        COALESCE((hr_settings #>> '{bpjs,jht,employee_rate}')::DECIMAL, 0.02),
        COALESCE((hr_settings #>> '{bpjs,jht,company_rate}')::DECIMAL, 0.037),
        COALESCE((hr_settings #>> '{bpjs,jp,employee_rate}')::DECIMAL, 0.01),
        COALESCE((hr_settings #>> '{bpjs,jp,company_rate}')::DECIMAL, 0.02),
        COALESCE((hr_settings #>> '{bpjs,jp,ceiling}')::DECIMAL, 10000000),
        COALESCE((hr_settings #>> '{bpjs,jkm,company_rate}')::DECIMAL, 0.003),
        bpjs_jkk_rate
    INTO
        v_kesehatan_emp_rate, v_kesehatan_comp_rate, v_kesehatan_ceiling,
        v_jht_emp_rate, v_jht_comp_rate,
        v_jp_emp_rate, v_jp_comp_rate, v_jp_ceiling,
        v_jkm_comp_rate,
        v_jkk_rate
    FROM companies LIMIT 1;

    IF v_kesehatan_emp_rate IS NULL THEN v_kesehatan_emp_rate := 0.01; END IF;
    IF v_kesehatan_comp_rate IS NULL THEN v_kesehatan_comp_rate := 0.04; END IF;
    IF v_kesehatan_ceiling IS NULL THEN v_kesehatan_ceiling := 12000000; END IF;
    IF v_jht_emp_rate IS NULL THEN v_jht_emp_rate := 0.02; END IF;
    IF v_jht_comp_rate IS NULL THEN v_jht_comp_rate := 0.037; END IF;
    IF v_jp_emp_rate IS NULL THEN v_jp_emp_rate := 0.01; END IF;
    IF v_jp_comp_rate IS NULL THEN v_jp_comp_rate := 0.02; END IF;
    IF v_jp_ceiling IS NULL THEN v_jp_ceiling := 10000000; END IF;
    IF v_jkm_comp_rate IS NULL THEN v_jkm_comp_rate := 0.003; END IF;
    IF v_jkk_rate IS NULL THEN v_jkk_rate := 0.54; END IF;

    SELECT bpjs_config INTO v_bpjs_config
    FROM employees
    WHERE id = p_employee_id;

    IF v_bpjs_config IS NOT NULL THEN
        v_kesehatan_enabled := COALESCE((v_bpjs_config #>> '{kesehatan,enabled}')::BOOLEAN, TRUE);
        v_jht_enabled       := COALESCE((v_bpjs_config #>> '{jht,enabled}')::BOOLEAN, TRUE);
        v_jp_enabled        := COALESCE((v_bpjs_config #>> '{jp,enabled}')::BOOLEAN, TRUE);
        v_jkk_enabled       := COALESCE((v_bpjs_config #>> '{jkk,enabled}')::BOOLEAN, TRUE);
        v_jkm_enabled       := COALESCE((v_bpjs_config #>> '{jkm,enabled}')::BOOLEAN, TRUE);

        IF v_bpjs_config #>> '{kesehatan,employee_rate}' IS NOT NULL THEN
            v_kesehatan_emp_rate := (v_bpjs_config #>> '{kesehatan,employee_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{kesehatan,company_rate}' IS NOT NULL THEN
            v_kesehatan_comp_rate := (v_bpjs_config #>> '{kesehatan,company_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jht,employee_rate}' IS NOT NULL THEN
            v_jht_emp_rate := (v_bpjs_config #>> '{jht,employee_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jht,company_rate}' IS NOT NULL THEN
            v_jht_comp_rate := (v_bpjs_config #>> '{jht,company_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jp,employee_rate}' IS NOT NULL THEN
            v_jp_emp_rate := (v_bpjs_config #>> '{jp,employee_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jp,company_rate}' IS NOT NULL THEN
            v_jp_comp_rate := (v_bpjs_config #>> '{jp,company_rate}')::DECIMAL;
        END IF;
        IF v_bpjs_config #>> '{jkm,company_rate}' IS NOT NULL THEN
            v_jkm_comp_rate := (v_bpjs_config #>> '{jkm,company_rate}')::DECIMAL;
        END IF;
    END IF;

    IF v_kesehatan_enabled THEN
        v_bpjs_kesehatan := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_emp_rate;
    ELSE
        v_bpjs_kesehatan := 0;
    END IF;

    IF v_jht_enabled THEN
        v_bpjs_jht := v_gross_salary * v_jht_emp_rate;
    ELSE
        v_bpjs_jht := 0;
    END IF;

    IF v_jp_enabled THEN
        v_bpjs_jp := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_emp_rate;
    ELSE
        v_bpjs_jp := 0;
    END IF;

    IF v_kesehatan_enabled THEN
        v_bpjs_ks_company := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_comp_rate;
    ELSE
        v_bpjs_ks_company := 0;
    END IF;

    IF v_jht_enabled THEN
        v_bpjs_jht_company := v_gross_salary * v_jht_comp_rate;
    ELSE
        v_bpjs_jht_company := 0;
    END IF;

    IF v_jp_enabled THEN
        v_bpjs_jp_company := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_comp_rate;
    ELSE
        v_bpjs_jp_company := 0;
    END IF;

    IF v_jkk_enabled THEN
        v_bpjs_jkk := v_gross_salary * (v_jkk_rate / 100);
    ELSE
        v_bpjs_jkk := 0;
    END IF;

    IF v_jkm_enabled THEN
        v_bpjs_jkm := v_gross_salary * v_jkm_comp_rate;
    ELSE
        v_bpjs_jkm := 0;
    END IF;

    v_pph21 := 0;

    v_total_deductions := v_bpjs_kesehatan + v_bpjs_jht + v_bpjs_jp + v_pph21 + p_loan_deduction + p_other_deductions;
    v_net_salary := v_gross_salary - v_total_deductions;

    v_company_cost := jsonb_build_array(
        jsonb_build_object('name', 'BPJS Kesehatan', 'amount', v_bpjs_ks_company),
        jsonb_build_object('name', 'BPJS JHT', 'amount', v_bpjs_jht_company),
        jsonb_build_object('name', 'BPJS JP', 'amount', v_bpjs_jp_company),
        jsonb_build_object('name', 'BPJS JKK', 'amount', v_bpjs_jkk),
        jsonb_build_object('name', 'BPJS JKM', 'amount', v_bpjs_jkm)
    );

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
        daily_wage = EXCLUDED.daily_wage,
        gross_salary = EXCLUDED.gross_salary,
        net_salary = EXCLUDED.net_salary,
        total_deductions = EXCLUDED.total_deductions,
        allowances = EXCLUDED.allowances,
        updated_at = NOW()
    RETURNING id INTO v_payroll_item_id;

    RETURN v_payroll_item_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
