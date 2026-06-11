-- +goose Up
-- ============================================================
-- Migration 00008: Reimbursements & Overtime Requests
-- ============================================================

-- ============================================================
-- Reimbursements
-- ============================================================
CREATE TABLE reimbursements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type            reimbursement_type NOT NULL,
    amount          DECIMAL(15,2) NOT NULL,
    description     TEXT NOT NULL,
    
    -- Receipt / Bukti
    receipt_urls    TEXT[],                                 -- Array foto/faktur/invoice
    
    -- Approval
    approval_trail  JSONB DEFAULT '[]'::jsonb,
    status          leave_status NOT NULL DEFAULT 'pending',
    
    -- Payment
    payment_method  VARCHAR(50) DEFAULT 'payroll',          -- payroll / manual_transfer
    paid_at         TIMESTAMPTZ,
    paid_by         UUID REFERENCES employees(id) ON DELETE SET NULL,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_reimbursements_updated_at
    BEFORE UPDATE ON reimbursements
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Overtime Requests
-- ============================================================
CREATE TABLE overtime_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Time
    date            DATE NOT NULL,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    total_hours     DECIMAL(5,2) NOT NULL,                  -- e.g. 2.5 hours
    overtime_type   overtime_type NOT NULL DEFAULT 'weekday',
    
    -- Details
    reason          TEXT NOT NULL,
    is_mandatory    BOOLEAN DEFAULT FALSE,                  -- Lembur wajib / sukarela
    
    -- Approval
    approval_trail  JSONB DEFAULT '[]'::jsonb,
    status          leave_status NOT NULL DEFAULT 'pending',
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_overtime_requests_updated_at
    BEFORE UPDATE ON overtime_requests
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Overtime Calculation (View untuk perhitungan lembur)
-- ============================================================
CREATE OR REPLACE VIEW overtime_calculation AS
SELECT
    otr.id,
    otr.employee_id,
    otr.date,
    otr.total_hours,
    otr.overtime_type,
    esh.base_salary,
    ROUND(esh.base_salary / 173, 2) AS hourly_rate,
    CASE
        WHEN otr.overtime_type = 'weekday' THEN
            CASE
                WHEN otr.total_hours <= 1 THEN ROUND((esh.base_salary / 173) * 1.5 * otr.total_hours, 2)
                ELSE ROUND((esh.base_salary / 173) * 1.5 * 1 + (esh.base_salary / 173) * 2 * (otr.total_hours - 1), 2)
            END
        WHEN otr.overtime_type = 'weekend' THEN
            CASE
                WHEN otr.total_hours <= 7 THEN ROUND((esh.base_salary / 173) * 2 * otr.total_hours, 2)
                ELSE ROUND((esh.base_salary / 173) * 2 * 7 + (esh.base_salary / 173) * 3 * (otr.total_hours - 7), 2)
            END
        WHEN otr.overtime_type = 'holiday' THEN
            CASE
                WHEN otr.total_hours <= 7 THEN ROUND((esh.base_salary / 173) * 2 * otr.total_hours, 2)
                ELSE ROUND((esh.base_salary / 173) * 2 * 7 + (esh.base_salary / 173) * 3 * (otr.total_hours - 7), 2)
            END
    END AS overtime_pay
FROM overtime_requests otr
LEFT JOIN LATERAL (
    SELECT base_salary FROM employee_salary_histories
    WHERE employee_id = otr.employee_id
    AND effective_date <= otr.date
    ORDER BY effective_date DESC
    LIMIT 1
) esh ON TRUE
WHERE otr.status = 'approved';

-- Indexes
CREATE INDEX idx_reimbursements_employee_id ON reimbursements(employee_id);
CREATE INDEX idx_reimbursements_status ON reimbursements(status);
CREATE INDEX idx_overtime_requests_employee_id ON overtime_requests(employee_id);
CREATE INDEX idx_overtime_requests_date ON overtime_requests(date);
CREATE INDEX idx_overtime_requests_status ON overtime_requests(status);

-- +goose Down
DROP VIEW IF EXISTS overtime_calculation;
DROP TABLE IF EXISTS overtime_requests;
DROP TABLE IF EXISTS reimbursements;
