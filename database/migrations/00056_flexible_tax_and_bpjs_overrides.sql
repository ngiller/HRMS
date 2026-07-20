-- +goose Up
-- ============================================================
-- Migration 00056: Flexible Tax & BPJS Overrides (Rate & Nominal)
-- ============================================================

-- 1. Add tax_config column to employees table
ALTER TABLE employees ADD COLUMN tax_config JSONB DEFAULT NULL;

COMMENT ON COLUMN employees.tax_config IS
'Per-employee PPh 21 tax overrides. Format JSONB:
{
  "override_type": "rate"|"nominal"|"none"|"free",
  "override_rate": 0.05,        -- Custom tax rate (5%)
  "override_nominal": 150000.00 -- Custom fixed tax amount
}';

-- 2. Create calculate_ter_rate helper function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_ter_rate(p_category VARCHAR, p_gross_salary NUMERIC)
RETURNS NUMERIC AS $$
BEGIN
    -- Category A: TK/0, TK/1, K/0
    IF p_category = 'A' THEN
        IF p_gross_salary <= 5400000 THEN RETURN 0;
        ELSIF p_gross_salary <= 5650000 THEN RETURN 0.0025;
        ELSIF p_gross_salary <= 5950000 THEN RETURN 0.005;
        ELSIF p_gross_salary <= 6300000 THEN RETURN 0.0075;
        ELSIF p_gross_salary <= 6750000 THEN RETURN 0.01;
        ELSIF p_gross_salary <= 7500000 THEN RETURN 0.0125;
        ELSIF p_gross_salary <= 8550000 THEN RETURN 0.015;
        ELSIF p_gross_salary <= 9650000 THEN RETURN 0.0175;
        ELSIF p_gross_salary <= 10950000 THEN RETURN 0.02;
        ELSIF p_gross_salary <= 12300000 THEN RETURN 0.0225;
        ELSIF p_gross_salary <= 13800000 THEN RETURN 0.025;
        ELSIF p_gross_salary <= 15500000 THEN RETURN 0.03;
        ELSIF p_gross_salary <= 17500000 THEN RETURN 0.04;
        ELSIF p_gross_salary <= 20000000 THEN RETURN 0.05;
        ELSIF p_gross_salary <= 24000000 THEN RETURN 0.06;
        ELSIF p_gross_salary <= 29000000 THEN RETURN 0.07;
        ELSIF p_gross_salary <= 35000000 THEN RETURN 0.08;
        ELSIF p_gross_salary <= 40000000 THEN RETURN 0.09;
        ELSIF p_gross_salary <= 47000000 THEN RETURN 0.10;
        ELSIF p_gross_salary <= 57000000 THEN RETURN 0.11;
        ELSIF p_gross_salary <= 73000000 THEN RETURN 0.12;
        ELSIF p_gross_salary <= 93000000 THEN RETURN 0.13;
        ELSIF p_gross_salary <= 120000000 THEN RETURN 0.14;
        ELSIF p_gross_salary <= 151000000 THEN RETURN 0.15;
        ELSIF p_gross_salary <= 189000000 THEN RETURN 0.16;
        ELSIF p_gross_salary <= 236000000 THEN RETURN 0.17;
        ELSIF p_gross_salary <= 293000000 THEN RETURN 0.18;
        ELSIF p_gross_salary <= 365000000 THEN RETURN 0.19;
        ELSIF p_gross_salary <= 452000000 THEN RETURN 0.20;
        ELSIF p_gross_salary <= 557000000 THEN RETURN 0.21;
        ELSIF p_gross_salary <= 683000000 THEN RETURN 0.22;
        ELSIF p_gross_salary <= 836000000 THEN RETURN 0.23;
        ELSIF p_gross_salary <= 1022000000 THEN RETURN 0.24;
        ELSE RETURN 0.25;
        END IF;
    -- Category B: TK/2, TK/3, K/1, K/2
    ELSIF p_category = 'B' THEN
        IF p_gross_salary <= 6200000 THEN RETURN 0;
        ELSIF p_gross_salary <= 6500000 THEN RETURN 0.0025;
        ELSIF p_gross_salary <= 6850000 THEN RETURN 0.005;
        ELSIF p_gross_salary <= 7300000 THEN RETURN 0.0075;
        ELSIF p_gross_salary <= 8000000 THEN RETURN 0.01;
        ELSIF p_gross_salary <= 9050000 THEN RETURN 0.0125;
        ELSIF p_gross_salary <= 10300000 THEN RETURN 0.015;
        ELSIF p_gross_salary <= 11800000 THEN RETURN 0.0175;
        ELSIF p_gross_salary <= 13800000 THEN RETURN 0.02;
        ELSIF p_gross_salary <= 15600000 THEN RETURN 0.0225;
        ELSIF p_gross_salary <= 17800000 THEN RETURN 0.025;
        ELSIF p_gross_salary <= 20300000 THEN RETURN 0.03;
        ELSIF p_gross_salary <= 23300000 THEN RETURN 0.04;
        ELSIF p_gross_salary <= 27300000 THEN RETURN 0.05;
        ELSIF p_gross_salary <= 32300000 THEN RETURN 0.06;
        ELSIF p_gross_salary <= 38300000 THEN RETURN 0.07;
        ELSIF p_gross_salary <= 45400000 THEN RETURN 0.08;
        ELSIF p_gross_salary <= 53600000 THEN RETURN 0.09;
        ELSIF p_gross_salary <= 63300000 THEN RETURN 0.10;
        ELSIF p_gross_salary <= 74800000 THEN RETURN 0.11;
        ELSIF p_gross_salary <= 88500000 THEN RETURN 0.12;
        ELSIF p_gross_salary <= 104700000 THEN RETURN 0.13;
        ELSIF p_gross_salary <= 123900000 THEN RETURN 0.14;
        ELSIF p_gross_salary <= 146500000 THEN RETURN 0.15;
        ELSIF p_gross_salary <= 173300000 THEN RETURN 0.16;
        ELSIF p_gross_salary <= 204900000 THEN RETURN 0.17;
        ELSIF p_gross_salary <= 242200000 THEN RETURN 0.18;
        ELSIF p_gross_salary <= 286200000 THEN RETURN 0.19;
        ELSIF p_gross_salary <= 338200000 THEN RETURN 0.20;
        ELSIF p_gross_salary <= 399700000 THEN RETURN 0.21;
        ELSIF p_gross_salary <= 472300000 THEN RETURN 0.22;
        ELSIF p_gross_salary <= 558100000 THEN RETURN 0.23;
        ELSIF p_gross_salary <= 659700000 THEN RETURN 0.24;
        ELSE RETURN 0.25;
        END IF;
    -- Category C: K/3
    ELSIF p_category = 'C' THEN
        IF p_gross_salary <= 6600000 THEN RETURN 0;
        ELSIF p_gross_salary <= 6950000 THEN RETURN 0.0025;
        ELSIF p_gross_salary <= 7350000 THEN RETURN 0.005;
        ELSIF p_gross_salary <= 7800000 THEN RETURN 0.0075;
        ELSIF p_gross_salary <= 8350000 THEN RETURN 0.01;
        ELSIF p_gross_salary <= 9050000 THEN RETURN 0.0125;
        ELSIF p_gross_salary <= 9850000 THEN RETURN 0.015;
        ELSIF p_gross_salary <= 10750000 THEN RETURN 0.0175;
        ELSIF p_gross_salary <= 11800000 THEN RETURN 0.02;
        ELSIF p_gross_salary <= 12950000 THEN RETURN 0.0225;
        ELSIF p_gross_salary <= 14250000 THEN RETURN 0.025;
        ELSIF p_gross_salary <= 15700000 THEN RETURN 0.03;
        ELSIF p_gross_salary <= 17300000 THEN RETURN 0.04;
        ELSIF p_gross_salary <= 19100000 THEN RETURN 0.05;
        ELSIF p_gross_salary <= 21100000 THEN RETURN 0.06;
        ELSIF p_gross_salary <= 23300000 THEN RETURN 0.07;
        ELSIF p_gross_salary <= 25700000 THEN RETURN 0.08;
        ELSIF p_gross_salary <= 29000000 THEN RETURN 0.09;
        ELSIF p_gross_salary <= 32700000 THEN RETURN 0.10;
        ELSIF p_gross_salary <= 36900000 THEN RETURN 0.11;
        ELSIF p_gross_salary <= 41700000 THEN RETURN 0.12;
        ELSIF p_gross_salary <= 47100000 THEN RETURN 0.13;
        ELSIF p_gross_salary <= 53200000 THEN RETURN 0.14;
        ELSIF p_gross_salary <= 60100000 THEN RETURN 0.15;
        ELSIF p_gross_salary <= 67900000 THEN RETURN 0.16;
        ELSIF p_gross_salary <= 76700000 THEN RETURN 0.17;
        ELSIF p_gross_salary <= 86700000 THEN RETURN 0.18;
        ELSIF p_gross_salary <= 98000000 THEN RETURN 0.19;
        ELSIF p_gross_salary <= 110800000 THEN RETURN 0.20;
        ELSIF p_gross_salary <= 125200000 THEN RETURN 0.21;
        ELSIF p_gross_salary <= 141500000 THEN RETURN 0.22;
        ELSIF p_gross_salary <= 159900000 THEN RETURN 0.23;
        ELSIF p_gross_salary <= 180700000 THEN RETURN 0.24;
        ELSE RETURN 0.25;
        END IF;
    ELSE
        RETURN 0;
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- 3. Drop function cascade to ensure ownership resets to running user
DROP FUNCTION IF EXISTS calculate_employee_payroll(uuid, uuid, numeric, numeric, integer, jsonb, numeric, numeric, numeric, numeric, numeric) CASCADE;

-- 4. Recreate function calculate_employee_payroll with PPh 21 and BPJS nominal overrides
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

    -- Default BPJS variables from Company Settings
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

    -- BPJS overrides
    v_bpjs_config              JSONB;
    v_kesehatan_enabled        BOOLEAN := TRUE;
    v_jht_enabled              BOOLEAN := TRUE;
    v_jp_enabled               BOOLEAN := TRUE;
    v_jkk_enabled              BOOLEAN := TRUE;
    v_jkm_enabled              BOOLEAN := TRUE;

    -- Reimbursement & Auto deductions
    v_reimbursement_total      DECIMAL(15,2) := 0;
    v_last_payroll_start       TIMESTAMPTZ := '1970-01-01'::timestamptz;
    v_current_period_start     DATE;
    v_current_period_end       DATE;
    v_total_other_deductions   DECIMAL(15,2);
    v_auto_overtime_pay        DECIMAL(15,2) := 0;
    v_overtime_hours           DECIMAL(5,2) := 0;
    v_auto_loan_deduction      DECIMAL(15,2) := 0;

    -- PPh 21 variables
    v_ptkp_status              VARCHAR(10);
    v_npwp                     BYTEA;
    v_tax_config               JSONB;
    v_ter_category             VARCHAR(5) := 'A';
    v_ter_rate                 NUMERIC := 0;
    v_tax_override_type        VARCHAR(20) := 'none';
    v_tax_override_rate        NUMERIC;
    v_tax_override_nominal     NUMERIC;
    v_has_npwp                 BOOLEAN := FALSE;
    v_pph21_desc               TEXT := '';
BEGIN
    -- Hitung total tunjangan
    SELECT COALESCE(SUM((value->>'amount')::DECIMAL), 0)
    INTO v_total_allowances
    FROM jsonb_array_elements(p_allowances);

    v_gross_salary := p_base_salary + p_daily_wage + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;

    -- ============================================================
    -- REIMBURSEMENT
    -- ============================================================
    SELECT start_date, end_date INTO v_current_period_start, v_current_period_end
    FROM payroll_periods
    WHERE id = p_payroll_period_id;

    SELECT COALESCE(MAX(pp.start_date), '1970-01-01'::date)::timestamptz
    INTO v_last_payroll_start
    FROM payroll_items pi
    JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
    WHERE pi.employee_id = p_employee_id
      AND pi.status IN ('calculated', 'approved', 'paid')
      AND pp.end_date <= v_current_period_start;

    SELECT COALESCE(SUM(r.amount), 0)
    INTO v_reimbursement_total
    FROM reimbursements r
    WHERE r.employee_id = p_employee_id
      AND r.status = 'paid'
      AND r.payment_method = 'payroll'
      AND r.deleted_at IS NULL
      AND r.paid_at IS NOT NULL
      AND r.paid_at > v_last_payroll_start;

    v_total_other_deductions := p_other_deductions + v_reimbursement_total;

    -- ============================================================
    -- OVERTIME
    -- ============================================================
    SELECT COALESCE(SUM(overtime_pay), 0), COALESCE(SUM(total_hours), 0)
    INTO v_auto_overtime_pay, v_overtime_hours
    FROM overtime_calculation
    WHERE employee_id = p_employee_id
      AND date >= v_current_period_start
      AND date <= v_current_period_end;
      
    p_overtime_pay := p_overtime_pay + v_auto_overtime_pay;
    v_gross_salary := p_base_salary + p_daily_wage + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;

    -- ============================================================
    -- LOAN
    -- ============================================================
    SELECT COALESCE(SUM(LEAST(installment_amount, remaining_balance)), 0)
    INTO v_auto_loan_deduction
    FROM loans
    WHERE employee_id = p_employee_id
      AND payment_method = 'payroll_deduction'
      AND remaining_balance > 0;
      
    p_loan_deduction := p_loan_deduction + v_auto_loan_deduction;

    -- ============================================================
    -- READ DEFAULT BPJS SETTINGS
    -- ============================================================
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

    -- ============================================================
    -- READ PER-EMPLOYEE BPJS OVERRIDES
    -- ============================================================
    SELECT bpjs_config, ptkp_status, encrypted_npwp, tax_config 
    INTO v_bpjs_config, v_ptkp_status, v_npwp, v_tax_config
    FROM employees
    WHERE id = p_employee_id;

    IF v_bpjs_config IS NOT NULL THEN
        v_kesehatan_enabled := COALESCE((v_bpjs_config #>> '{kesehatan,enabled}')::BOOLEAN, TRUE);
        v_jht_enabled       := COALESCE((v_bpjs_config #>> '{jht,enabled}')::BOOLEAN, TRUE);
        v_jp_enabled        := COALESCE((v_bpjs_config #>> '{jp,enabled}')::BOOLEAN, TRUE);
        v_jkk_enabled       := COALESCE((v_bpjs_config #>> '{jkk,enabled}')::BOOLEAN, TRUE);
        v_jkm_enabled       := COALESCE((v_bpjs_config #>> '{jkm,enabled}')::BOOLEAN, TRUE);

        -- Rates overrides
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

    -- ============================================================
    -- CALCULATE BPJS (PEKERJA) — with custom nominal overrides
    -- ============================================================
    IF v_kesehatan_enabled THEN
        IF v_bpjs_config #>> '{kesehatan,employee_nominal}' IS NOT NULL THEN
            v_bpjs_kesehatan := (v_bpjs_config #>> '{kesehatan,employee_nominal}')::DECIMAL;
        ELSE
            v_bpjs_kesehatan := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_emp_rate;
        END IF;
    ELSE
        v_bpjs_kesehatan := 0;
    END IF;

    IF v_jht_enabled THEN
        IF v_bpjs_config #>> '{jht,employee_nominal}' IS NOT NULL THEN
            v_bpjs_jht := (v_bpjs_config #>> '{jht,employee_nominal}')::DECIMAL;
        ELSE
            v_bpjs_jht := v_gross_salary * v_jht_emp_rate;
        END IF;
    ELSE
        v_bpjs_jht := 0;
    END IF;

    IF v_jp_enabled THEN
        IF v_bpjs_config #>> '{jp,employee_nominal}' IS NOT NULL THEN
            v_bpjs_jp := (v_bpjs_config #>> '{jp,employee_nominal}')::DECIMAL;
        ELSE
            v_bpjs_jp := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_emp_rate;
        END IF;
    ELSE
        v_bpjs_jp := 0;
    END IF;

    -- ============================================================
    -- CALCULATE BPJS (PERUSAHAAN) — with custom nominal overrides
    -- ============================================================
    IF v_kesehatan_enabled THEN
        IF v_bpjs_config #>> '{kesehatan,company_nominal}' IS NOT NULL THEN
            v_bpjs_ks_company := (v_bpjs_config #>> '{kesehatan,company_nominal}')::DECIMAL;
        ELSE
            v_bpjs_ks_company := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_comp_rate;
        END IF;
    ELSE
        v_bpjs_ks_company := 0;
    END IF;

    IF v_jht_enabled THEN
        IF v_bpjs_config #>> '{jht,company_nominal}' IS NOT NULL THEN
            v_bpjs_jht_company := (v_bpjs_config #>> '{jht,company_nominal}')::DECIMAL;
        ELSE
            v_bpjs_jht_company := v_gross_salary * v_jht_comp_rate;
        END IF;
    ELSE
        v_bpjs_jht_company := 0;
    END IF;

    IF v_jp_enabled THEN
        IF v_bpjs_config #>> '{jp,company_nominal}' IS NOT NULL THEN
            v_bpjs_jp_company := (v_bpjs_config #>> '{jp,company_nominal}')::DECIMAL;
        ELSE
            v_bpjs_jp_company := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_comp_rate;
        END IF;
    ELSE
        v_bpjs_jp_company := 0;
    END IF;

    IF v_jkk_enabled THEN
        IF v_bpjs_config #>> '{jkk,company_nominal}' IS NOT NULL THEN
            v_bpjs_jkk := (v_bpjs_config #>> '{jkk,company_nominal}')::DECIMAL;
        ELSE
            v_bpjs_jkk := v_gross_salary * (v_jkk_rate / 100);
        END IF;
    ELSE
        v_bpjs_jkk := 0;
    END IF;

    IF v_jkm_enabled THEN
        IF v_bpjs_config #>> '{jkm,company_nominal}' IS NOT NULL THEN
            v_bpjs_jkm := (v_bpjs_config #>> '{jkm,company_nominal}')::DECIMAL;
        ELSE
            v_bpjs_jkm := v_gross_salary * v_jkm_comp_rate;
        END IF;
    ELSE
        v_bpjs_jkm := 0;
    END IF;

    -- ============================================================
    -- CALCULATE PPh 21 (TAX)
    -- ============================================================
    IF v_npwp IS NOT NULL AND length(decrypt_sensitive(v_npwp)) > 0 THEN
        v_has_npwp := TRUE;
    END IF;

    -- Tentukan Kategori TER berdasarkan PTKP Status
    IF v_ptkp_status IN ('TK0', 'TK1', 'K0') THEN
        v_ter_category := 'A';
    ELSIF v_ptkp_status IN ('TK2', 'TK3', 'K1', 'K2') THEN
        v_ter_category := 'B';
    ELSIF v_ptkp_status IN ('K3') THEN
        v_ter_category := 'C';
    ELSE
        v_ter_category := 'A';
    END IF;

    -- Ambil Overrides Pajak
    IF v_tax_config IS NOT NULL THEN
        v_tax_override_type := COALESCE(v_tax_config->>'override_type', 'none');
        v_tax_override_rate := (v_tax_config->>'override_rate')::NUMERIC;
        v_tax_override_nominal := (v_tax_config->>'override_nominal')::NUMERIC;
    END IF;

    IF v_tax_override_type = 'free' THEN
        v_pph21 := 0;
        v_pph21_desc := 'Bebas Pajak (Override 0%)';
    ELSIF v_tax_override_type = 'nominal' AND v_tax_override_nominal IS NOT NULL THEN
        v_pph21 := v_tax_override_nominal;
        v_pph21_desc := 'Kustom Nominal: ' || v_pph21::TEXT;
    ELSIF v_tax_override_type = 'rate' AND v_tax_override_rate IS NOT NULL THEN
        v_pph21 := v_gross_salary * v_tax_override_rate;
        v_pph21_desc := 'Kustom Tarif: ' || (v_tax_override_rate * 100)::TEXT || '%';
    ELSE
        -- Kalkulasi standard TER
        v_ter_rate := calculate_ter_rate(v_ter_category, v_gross_salary);
        IF NOT v_has_npwp THEN
            v_ter_rate := v_ter_rate * 1.20; -- Tanpa NPWP dikenakan tarif 20% lebih tinggi
            v_pph21_desc := 'TER Kat. ' || v_ter_category || ' (' || (v_ter_rate*100)::TEXT || '% - Tanpa NPWP)';
        ELSE
            v_pph21_desc := 'TER Kat. ' || v_ter_category || ' (' || (v_ter_rate*100)::TEXT || '%)';
        END IF;
        v_pph21 := v_gross_salary * v_ter_rate;
    END IF;

    -- ============================================================
    -- TOTAL DEDUCTIONS & NET SALARY
    -- ============================================================
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
        allowances, overtime_pay, overtime_hours, thr_amount, bonus_amount,
        gross_salary,
        deductions, pph21_amount, pph21_ter_category, pph21_description, 
        bpjs_kesehatan, bpjs_jht, bpjs_jp,
        loan_deduction, other_deductions, total_deductions,
        net_salary,
        company_cost,
        bpjs_kesehatan_company, bpjs_jht_company, bpjs_jp_company, bpjs_jkk, bpjs_jkm,
        status
    ) VALUES (
        p_payroll_period_id, p_employee_id,
        p_base_salary, p_daily_wage, p_total_days_worked,
        p_allowances, p_overtime_pay, v_overtime_hours, p_thr_amount, p_bonus_amount,
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
        v_pph21, v_ter_category, v_pph21_desc,
        v_bpjs_kesehatan, v_bpjs_jht, v_bpjs_jp,
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
        overtime_hours = EXCLUDED.overtime_hours,
        thr_amount = EXCLUDED.thr_amount,
        bonus_amount = EXCLUDED.bonus_amount,
        gross_salary = EXCLUDED.gross_salary,
        deductions = EXCLUDED.deductions,
        pph21_amount = EXCLUDED.pph21_amount,
        pph21_ter_category = EXCLUDED.pph21_ter_category,
        pph21_description = EXCLUDED.pph21_description,
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

-- +goose Down
-- ============================================================
-- Rollback
-- ============================================================
ALTER TABLE employees DROP COLUMN IF EXISTS tax_config;
DROP FUNCTION IF EXISTS calculate_ter_rate(VARCHAR, NUMERIC);
