-- +goose Up
-- ============================================================
-- Migration 00046: Employee Mutations & Promotions
-- ============================================================
-- Mencatat riwayat mutasi/promosi/demosi/transfer karyawan
-- dengan approval workflow integration.
-- ============================================================

CREATE TABLE employee_mutations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    mutation_type   VARCHAR(50) NOT NULL,
    -- mutation_type: promotion, demotion, transfer, position_change, status_change, salary_change

    -- Nilai sebelum mutasi
    old_department_id       UUID REFERENCES departments(id),
    old_position_id         UUID REFERENCES positions(id),
    old_position_grade_id   UUID REFERENCES position_grades(id),
    old_employment_status   VARCHAR(50),
    old_base_salary         DECIMAL(15,2),

    -- Nilai setelah mutasi
    new_department_id       UUID REFERENCES departments(id),
    new_position_id         UUID REFERENCES positions(id),
    new_position_grade_id   UUID REFERENCES position_grades(id),
    new_employment_status   VARCHAR(50),
    new_base_salary         DECIMAL(15,2),

    -- Alasan & dokumen pendukung
    reason              TEXT NOT NULL,
    document_url        TEXT,
    effective_date      DATE NOT NULL,
    notes               TEXT,

    -- Status & approval
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending / approved / rejected / cancelled
    approved_by         UUID REFERENCES employees(id),
    approved_at         TIMESTAMPTZ,
    rejection_reason    TEXT,
    approval_trail      JSONB DEFAULT '[]'::jsonb,

    -- Audit
    created_by          UUID NOT NULL REFERENCES employees(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_employee_mutations_employee ON employee_mutations(employee_id, deleted_at);
CREATE INDEX idx_employee_mutations_status ON employee_mutations(status, deleted_at);
CREATE INDEX idx_employee_mutations_effective ON employee_mutations(effective_date);

CREATE TRIGGER set_employee_mutations_updated_at
    BEFORE UPDATE ON employee_mutations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS set_employee_mutations_updated_at ON employee_mutations;
DROP TABLE IF EXISTS employee_mutations;
