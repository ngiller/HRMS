-- +goose Up
-- ============================================================
-- Migration 00009: Employee Loans
-- ============================================================

-- ============================================================
-- Loans (Pinjaman Karyawan)
-- ============================================================
CREATE TABLE loans (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    loan_type           loan_type NOT NULL DEFAULT 'regular',
    
    -- Loan Details
    amount              DECIMAL(15,2) NOT NULL,
    interest_rate       DECIMAL(5,2) NOT NULL DEFAULT 0.00,  -- Persentase bunga (0% = tanpa bunga)
    total_interest      DECIMAL(15,2) DEFAULT 0.00,
    total_amount        DECIMAL(15,2) NOT NULL,              -- amount + interest
    
    -- Payment Terms
    installment_count   INTEGER NOT NULL,                     -- Tenor (3-24 bulan)
    installment_amount  DECIMAL(15,2) NOT NULL,               -- Per bulan
    payment_method      loan_payment_method NOT NULL DEFAULT 'payroll_deduction',
    
    -- Remaining
    remaining_balance   DECIMAL(15,2) NOT NULL,
    
    -- Description
    purpose             TEXT NOT NULL,                         -- Tujuan pinjaman
    
    -- Approval
    approval_trail      JSONB DEFAULT '[]'::jsonb,
    status              loan_status NOT NULL DEFAULT 'pending',
    
    -- Disbursement
    disbursed_at        TIMESTAMPTZ,
    disbursed_by        UUID REFERENCES employees(id) ON DELETE SET NULL,
    
    -- Settlement
    settled_at          TIMESTAMPTZ,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE TRIGGER set_loans_updated_at
    BEFORE UPDATE ON loans
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Loan Installments (Cicilan Pinjaman)
-- ============================================================
CREATE TABLE loan_installments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    loan_id             UUID NOT NULL REFERENCES loans(id) ON DELETE CASCADE,
    installment_number  INTEGER NOT NULL,
    amount              DECIMAL(15,2) NOT NULL,
    due_date            DATE NOT NULL,
    
    -- Payment
    paid_date           DATE,
    paid_amount         DECIMAL(15,2),
    payment_source      VARCHAR(50),                          -- 'payroll', 'manual_transfer'
    payroll_period_id   UUID,                                 -- FK ke payroll_periods (jika potong gaji)
    
    -- Status
    status              VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, paid, skipped, overdue
    
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(loan_id, installment_number)
);

CREATE TRIGGER set_loan_installments_updated_at
    BEFORE UPDATE ON loan_installments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Indexes
CREATE INDEX idx_loans_employee_id ON loans(employee_id);
CREATE INDEX idx_loans_status ON loans(status);
CREATE INDEX idx_loans_payment_method ON loans(payment_method);
CREATE INDEX idx_loan_installments_loan_id ON loan_installments(loan_id);
CREATE INDEX idx_loan_installments_due_date ON loan_installments(due_date);
CREATE INDEX idx_loan_installments_status ON loan_installments(status);
CREATE INDEX idx_loan_installments_payroll_period ON loan_installments(payroll_period_id);

-- +goose Down
DROP TABLE IF EXISTS loan_installments;
DROP TABLE IF EXISTS loans;
