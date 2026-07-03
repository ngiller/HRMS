-- +goose Up
-- ============================================================
-- Migration 00019: Flexible BPJS Configuration
-- ============================================================
-- Sebelumnya: BPJS rates (1%, 2%, 4%, dll) hardcoded di
-- function calculate_employee_payroll().
--
-- Sesudah: Semua rate dan ceiling bisa dikonfigurasi via
-- companies.hr_settings, plus override per karyawan.
-- ============================================================

-- ============================================================
-- 1. Tambah kolom bpjs_config di employees
-- ============================================================
-- Format JSONB:
-- {
--   "kesehatan": {"enabled": true|false, "employee_rate": 0.01, "company_rate": 0.04},
--   "jht":       {"enabled": true|false, "employee_rate": 0.02, "company_rate": 0.037},
--   "jp":        {"enabled": true|false, "employee_rate": 0.01, "company_rate": 0.02},
--   "jkk":       {"enabled": true|false, "company_rate": null},
--   "jkm":       {"enabled": true|false, "company_rate": 0.003}
-- }
-- Jika sebuah komponen tidak disebutkan, pakai default dari
-- companies.hr_settings. Jika tidak ada override, pakai
-- nilai default pemerintah.
ALTER TABLE employees
    ADD COLUMN bpjs_config JSONB DEFAULT NULL;

COMMENT ON COLUMN employees.bpjs_config IS
'Per-employee BPJS overrides. Format JSONB:
{
  "kesehatan": {"enabled": false},           -- Nonaktifkan BPJS Kesehatan
  "jht": {"employee_rate": 0.01},            -- Override rate JHT pekerja (1%)
  "jp": {"company_rate": 0.015}              -- Override rate JP perusahaan (1.5%)
}
Setiap field opsional — hanya override yang disebutkan.' ;

-- ============================================================
-- 2. Update function calculate_employee_payroll()
-- ============================================================
-- Semua BPJS rates (persentase) sekarang dibaca dari
-- companies.hr_settings, bukan hardcoded.
-- Per-employee override via employees.bpjs_config.
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

    -- BPJS results
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

    -- === BPJS CONFIG variables ===
    -- Company-level config (dari hr_settings)
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

    -- Per-employee overrides
    v_bpjs_config              JSONB;
    v_kesehatan_enabled        BOOLEAN := TRUE;
    v_jht_enabled              BOOLEAN := TRUE;
    v_jp_enabled               BOOLEAN := TRUE;
    v_jkk_enabled              BOOLEAN := TRUE;
    v_jkm_enabled              BOOLEAN := TRUE;
BEGIN
    -- Hitung total tunjangan
    SELECT COALESCE(SUM((value->>'amount')::DECIMAL), 0)
    INTO v_total_allowances
    FROM jsonb_array_elements(p_allowances);

    -- Gross Salary
    v_gross_salary := p_base_salary + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;

    -- ============================================================
    -- BACA BPJS CONFIG DARI COMPANY SETTINGS
    -- ============================================================
    -- Struktur hr_settings:
    -- {
    --   "bpjs": {
    --     "kesehatan": {"employee_rate": 0.01, "company_rate": 0.04, "ceiling": 12000000},
    --     "jht":       {"employee_rate": 0.02, "company_rate": 0.037, "ceiling": null},
    --     "jp":        {"employee_rate": 0.01, "company_rate": 0.02, "ceiling": 10000000},
    --     "jkm":       {"company_rate": 0.003, "ceiling": null},
    --     "jkk":       {"company_rate": null}  -- Gunakan bpjs_jkk_rate dari kolom companies
    --   }
    -- }
    SELECT
        -- Kesehatan
        COALESCE((hr_settings #>> '{bpjs,kesehatan,employee_rate}')::DECIMAL, 0.01),
        COALESCE((hr_settings #>> '{bpjs,kesehatan,company_rate}')::DECIMAL, 0.04),
        COALESCE((hr_settings #>> '{bpjs,kesehatan,ceiling}')::DECIMAL, 12000000),
        -- JHT
        COALESCE((hr_settings #>> '{bpjs,jht,employee_rate}')::DECIMAL, 0.02),
        COALESCE((hr_settings #>> '{bpjs,jht,company_rate}')::DECIMAL, 0.037),
        -- JP
        COALESCE((hr_settings #>> '{bpjs,jp,employee_rate}')::DECIMAL, 0.01),
        COALESCE((hr_settings #>> '{bpjs,jp,company_rate}')::DECIMAL, 0.02),
        COALESCE((hr_settings #>> '{bpjs,jp,ceiling}')::DECIMAL, 10000000),
        -- JKM
        COALESCE((hr_settings #>> '{bpjs,jkm,company_rate}')::DECIMAL, 0.003),
        -- JKK (masih dari kolom companies.bpjs_jkk_rate)
        bpjs_jkk_rate
    INTO
        v_kesehatan_emp_rate, v_kesehatan_comp_rate, v_kesehatan_ceiling,
        v_jht_emp_rate, v_jht_comp_rate,
        v_jp_emp_rate, v_jp_comp_rate, v_jp_ceiling,
        v_jkm_comp_rate,
        v_jkk_rate
    FROM companies LIMIT 1;

    -- Fallback jika tidak ada company record
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
    -- BACA PER-EMPLOYEE BPJS OVERRIDE
    -- ============================================================
    SELECT bpjs_config INTO v_bpjs_config
    FROM employees
    WHERE id = p_employee_id;

    IF v_bpjs_config IS NOT NULL THEN
        -- Enable/disable per komponen
        v_kesehatan_enabled := COALESCE((v_bpjs_config #>> '{kesehatan,enabled}')::BOOLEAN, TRUE);
        v_jht_enabled       := COALESCE((v_bpjs_config #>> '{jht,enabled}')::BOOLEAN, TRUE);
        v_jp_enabled        := COALESCE((v_bpjs_config #>> '{jp,enabled}')::BOOLEAN, TRUE);
        v_jkk_enabled       := COALESCE((v_bpjs_config #>> '{jkk,enabled}')::BOOLEAN, TRUE);
        v_jkm_enabled       := COALESCE((v_bpjs_config #>> '{jkm,enabled}')::BOOLEAN, TRUE);

        -- Override rate per komponen (jika specified)
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
    -- HITUNG BPJS (PEKERJA)
    -- ============================================================
    -- BPJS Kesehatan (0% jika disabled)
    IF v_kesehatan_enabled THEN
        v_bpjs_kesehatan := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_emp_rate;
    ELSE
        v_bpjs_kesehatan := 0;
    END IF;

    -- BPJS JHT
    IF v_jht_enabled THEN
        v_bpjs_jht := v_gross_salary * v_jht_emp_rate;
    ELSE
        v_bpjs_jht := 0;
    END IF;

    -- BPJS JP
    IF v_jp_enabled THEN
        v_bpjs_jp := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_emp_rate;
    ELSE
        v_bpjs_jp := 0;
    END IF;

    -- ============================================================
    -- HITUNG BPJS (PERUSAHAAN)
    -- ============================================================
    -- BPJS Kesehatan (perusahaan)
    IF v_kesehatan_enabled THEN
        v_bpjs_ks_company := LEAST(v_gross_salary, v_kesehatan_ceiling) * v_kesehatan_comp_rate;
    ELSE
        v_bpjs_ks_company := 0;
    END IF;

    -- BPJS JHT (perusahaan)
    IF v_jht_enabled THEN
        v_bpjs_jht_company := v_gross_salary * v_jht_comp_rate;
    ELSE
        v_bpjs_jht_company := 0;
    END IF;

    -- BPJS JP (perusahaan)
    IF v_jp_enabled THEN
        v_bpjs_jp_company := LEAST(v_gross_salary, v_jp_ceiling) * v_jp_comp_rate;
    ELSE
        v_bpjs_jp_company := 0;
    END IF;

    -- BPJS JKK (perusahaan) — rate dari kolom companies.bpjs_jkk_rate
    IF v_jkk_enabled THEN
        v_bpjs_jkk := v_gross_salary * (v_jkk_rate / 100);
    ELSE
        v_bpjs_jkk := 0;
    END IF;

    -- BPJS JKM (perusahaan)
    IF v_jkm_enabled THEN
        v_bpjs_jkm := v_gross_salary * v_jkm_comp_rate;
    ELSE
        v_bpjs_jkm := 0;
    END IF;

    -- ============================================================
    -- PPh 21 (dihitung oleh aplikasi dengan TER)
    -- ============================================================
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
-- +goose StatementEnd

-- +goose Down
-- ============================================================
-- Rollback: Hapus kolom bpjs_config, restore function ke versi lama
-- ============================================================

ALTER TABLE employees DROP COLUMN IF EXISTS bpjs_config;

-- Restore function ke versi sebelumnya (hardcoded rates)
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
    v_jkk_rate DECIMAL(5,2);
    v_payroll_item_id UUID;
    v_bpjs_ceiling DECIMAL(15,2);
    v_bpjs_jp_ceiling DECIMAL(15,2) := 10000000;
BEGIN
    SELECT COALESCE(SUM((value->>'amount')::DECIMAL), 0)
    INTO v_total_allowances
    FROM jsonb_array_elements(p_allowances);

    v_gross_salary := p_base_salary + v_total_allowances + p_overtime_pay + p_thr_amount + p_bonus_amount;

    SELECT
        COALESCE(hr_settings->>'bpjs_kesehatan_ceiling', '12000000')::DECIMAL,
        COALESCE(hr_settings->>'bpjs_jp_ceiling', '10000000')::DECIMAL,
        bpjs_jkk_rate
    INTO v_bpjs_ceiling, v_bpjs_jp_ceiling, v_jkk_rate
    FROM companies LIMIT 1;

    IF v_bpjs_ceiling IS NULL THEN v_bpjs_ceiling := 12000000; END IF;
    IF v_bpjs_jp_ceiling IS NULL THEN v_bpjs_jp_ceiling := 10000000; END IF;
    IF v_jkk_rate IS NULL THEN v_jkk_rate := 0.54; END IF;

    v_bpjs_kesehatan := LEAST(v_gross_salary, v_bpjs_ceiling) * 0.01;
    v_bpjs_jht := v_gross_salary * 0.02;
    v_bpjs_jp := LEAST(v_gross_salary, v_bpjs_jp_ceiling) * 0.01;

    v_bpjs_ks_company := LEAST(v_gross_salary, v_bpjs_ceiling) * 0.04;
    v_bpjs_jht_company := v_gross_salary * 0.037;
    v_bpjs_jp_company := LEAST(v_gross_salary, v_bpjs_jp_ceiling) * 0.02;
    v_bpjs_jkk := v_gross_salary * (v_jkk_rate / 100);
    v_bpjs_jkm := v_gross_salary * 0.003;

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
        gross_salary = EXCLUDED.gross_salary,
        net_salary = EXCLUDED.net_salary,
        total_deductions = EXCLUDED.total_deductions,
        updated_at = NOW()
    RETURNING id INTO v_payroll_item_id;

    RETURN v_payroll_item_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
