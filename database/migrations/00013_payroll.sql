-- +goose Up
-- ============================================================
-- Migration 00013: Payroll System
-- ============================================================

-- ============================================================
-- Payroll Periods
-- ============================================================
CREATE TABLE payroll_periods (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    month           INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year            INTEGER NOT NULL,
    period_name     VARCHAR(50) NOT NULL,                     -- e.g., "Januari 2026"
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    
    -- Status workflow
    status          payroll_status NOT NULL DEFAULT 'draft',
    
    -- Approval
    approved_by     UUID REFERENCES employees(id) ON DELETE SET NULL,
    approved_at     TIMESTAMPTZ,
    paid_by         UUID REFERENCES employees(id) ON DELETE SET NULL,
    paid_at         TIMESTAMPTZ,
    
    -- Summary
    total_employee  INTEGER DEFAULT 0,
    total_gross     DECIMAL(20,2) DEFAULT 0,
    total_deductions DECIMAL(20,2) DEFAULT 0,
    total_net       DECIMAL(20,2) DEFAULT 0,
    total_company_cost DECIMAL(20,2) DEFAULT 0,              -- BPJS perusahaan + tunjangan
    
    -- Timestamps
    created_by      UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    
    UNIQUE(month, year)
);

CREATE TRIGGER set_payroll_periods_updated_at
    BEFORE UPDATE ON payroll_periods
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Payroll Items (per karyawan per periode)
-- ============================================================
CREATE TABLE payroll_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payroll_period_id   UUID NOT NULL REFERENCES payroll_periods(id) ON DELETE CASCADE,
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    
    -- ===================
    -- INCOME (Pendapatan)
    -- ===================
    -- Gaji Pokok
    base_salary         DECIMAL(15,2) NOT NULL DEFAULT 0,
    daily_wage          DECIMAL(15,2) DEFAULT 0,             -- Untuk karyawan harian
    total_days_worked   INTEGER DEFAULT 0,                   -- Untuk prorate
    
    -- Tunjangan (flexible JSON)
    allowances          JSONB DEFAULT '[]'::jsonb,           -- [{name, amount}]
    
    -- Lembur
    overtime_pay        DECIMAL(15,2) DEFAULT 0,
    overtime_hours      DECIMAL(5,2) DEFAULT 0,
    
    -- THR
    thr_amount          DECIMAL(15,2) DEFAULT 0,
    
    -- Bonus / Insentif
    bonus_amount        DECIMAL(15,2) DEFAULT 0,
    
    -- Gross Salary (before deductions)
    gross_salary        DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- =====================
    -- DEDUCTIONS (Potongan)
    -- =====================
    deductions          JSONB DEFAULT '[]'::jsonb,           -- [{name, amount}]
    
    -- PPh 21
    pph21_amount        DECIMAL(15,2) DEFAULT 0,
    pph21_ter_category  VARCHAR(5),                          -- A, B, or C
    pph21_description   TEXT,                                -- Perhitungan detail
    
    -- BPJS (Pekerja)
    bpjs_kesehatan      DECIMAL(15,2) DEFAULT 0,             -- 1% dari gaji
    bpjs_jht            DECIMAL(15,2) DEFAULT 0,             -- 2% dari gaji
    bpjs_jp             DECIMAL(15,2) DEFAULT 0,             -- 1% dari gaji
    
    -- Pinjaman
    loan_deduction      DECIMAL(15,2) DEFAULT 0,
    
    -- Lain-lain
    other_deductions    DECIMAL(15,2) DEFAULT 0,
    
    -- Total Deductions
    total_deductions    DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- =====================
    -- NET SALARY (Take Home Pay)
    -- =====================
    net_salary          DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- =====================
    -- COMPANY CONTRIBUTIONS
    -- =====================
    company_cost        JSONB DEFAULT '[]'::jsonb,           -- [{name, amount}]
    -- BPJS Perusahaan
    bpjs_kesehatan_company  DECIMAL(15,2) DEFAULT 0,         -- 4% dari gaji
    bpjs_jht_company        DECIMAL(15,2) DEFAULT 0,         -- 3.7% dari gaji
    bpjs_jp_company         DECIMAL(15,2) DEFAULT 0,         -- 2% dari gaji
    bpjs_jkk               DECIMAL(15,2) DEFAULT 0,          -- 0.24%-1.74%
    bpjs_jkm               DECIMAL(15,2) DEFAULT 0,          -- 0.3%
    
    -- =====================
    -- THR CALCULATION
    -- =====================
    thr_eligibility     BOOLEAN DEFAULT FALSE,
    thr_proportional    DECIMAL(15,2) DEFAULT 0,             -- THR proporsional
    
    -- Status
    status              payroll_status NOT NULL DEFAULT 'draft',
    notes               TEXT,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(payroll_period_id, employee_id)
);

CREATE TRIGGER set_payroll_items_updated_at
    BEFORE UPDATE ON payroll_items
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Payslip View
-- ============================================================
CREATE OR REPLACE VIEW payslip_view AS
SELECT
    pi.id AS payroll_item_id,
    pp.month,
    pp.year,
    e.id AS employee_id,
    e.employee_id AS nip,
    e.full_name,
    d.name AS department_name,
    pos.name AS position_name,
    
    -- Income
    pi.base_salary,
    pi.daily_wage,
    pi.allowances,
    pi.overtime_pay,
    pi.thr_amount,
    pi.bonus_amount,
    pi.gross_salary,
    
    -- Deductions
    pi.pph21_amount,
    pi.bpjs_kesehatan,
    pi.bpjs_jht,
    pi.bpjs_jp,
    pi.loan_deduction,
    pi.total_deductions,
    
    -- Take Home
    pi.net_salary,
    
    -- Company
    pi.company_cost,
    pi.bpjs_kesehatan_company,
    pi.bpjs_jht_company,
    pi.bpjs_jp_company,
    pi.bpjs_jkk,
    pi.bpjs_jkm,
    
    -- Status
    pi.status,
    pp.status AS period_status
FROM payroll_items pi
JOIN payroll_periods pp ON pp.id = pi.payroll_period_id
JOIN employees e ON e.id = pi.employee_id
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN positions pos ON pos.id = e.position_id
WHERE pi.deleted_at IS NULL AND e.deleted_at IS NULL;

-- Indexes
CREATE INDEX idx_payroll_periods_date ON payroll_periods(year DESC, month DESC);
CREATE INDEX idx_payroll_periods_status ON payroll_periods(status);
CREATE INDEX idx_payroll_items_period_id ON payroll_items(payroll_period_id);
CREATE INDEX idx_payroll_items_employee_id ON payroll_items(employee_id);
CREATE INDEX idx_payroll_items_status ON payroll_items(status);

-- +goose Down
DROP VIEW IF EXISTS payslip_view;
DROP TABLE IF EXISTS payroll_items;
DROP TABLE IF EXISTS payroll_periods;
