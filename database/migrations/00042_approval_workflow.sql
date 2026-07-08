-- ================================================================
-- Migration 00042: Approval Workflow Engine
-- Membuat tabel konfigurasi workflow multi-level approval
-- Seed default workflows untuk leave, overtime, reimbursement, loan, shift_change
-- ================================================================

-- +goose Up
-- +goose StatementBegin

-- ─── 1. Approval Workflows ─────────────────────────────────────
-- Mendefinisikan workflow per entity type
CREATE TABLE IF NOT EXISTS approval_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,      -- 'leave', 'overtime', 'reimbursement', 'loan', 'shift_change'
    name VARCHAR(255) NOT NULL,            -- Nama workflow (e.g., "Cuti Tahunan")
    description TEXT,                       -- Deskripsi
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(entity_type, name)
);

-- ─── 2. Approval Workflow Steps ────────────────────────────────
-- Mendefinisikan langkah-langkah approval dalam workflow
CREATE TABLE IF NOT EXISTS approval_workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES approval_workflows(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,           -- Urutan step (1, 2, 3, ...)
    approver_type VARCHAR(50) NOT NULL,    -- 'approval_line', 'hr_manager', 'finance', 'director', 'department_head', 'specific_role'
    approver_role_id UUID REFERENCES roles(id), -- Jika approver_type = 'specific_role'
    -- Kondisi: kapan step ini berlaku (optional)
    condition_field VARCHAR(50),           -- Field yang dievaluasi: 'total_days', 'amount', dll
    condition_operator VARCHAR(10),        -- '>', '>=', '<', '<=', '=='
    condition_value DECIMAL(20,2),         -- Nilai pembanding
    -- Eskalasi jika tidak direspon dalam waktu tertentu
    escalation_hours INTEGER DEFAULT 48,
    escalation_step_id UUID REFERENCES approval_workflow_steps(id), -- Ke step mana eskalasi
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_condition CHECK (
        (condition_field IS NULL AND condition_operator IS NULL AND condition_value IS NULL) OR
        (condition_field IS NOT NULL AND condition_operator IS NOT NULL AND condition_value IS NOT NULL)
    )
);

-- Index untuk performance
CREATE INDEX idx_aw_entity_type ON approval_workflows(entity_type, is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_aws_workflow_id ON approval_workflow_steps(workflow_id);
CREATE INDEX idx_aws_step_order ON approval_workflow_steps(workflow_id, step_order);

-- ─── 3. Seed Default Workflows ─────────────────────────────────

-- Helper function to create a workflow step
-- (akan dipanggil di DO block)

DO $$
DECLARE
    v_leave_id UUID;
    v_overtime_id UUID;
    v_reimbursement_id UUID;
    v_shift_change_id UUID;
    v_loan_id UUID;
    v_hr_manager_id UUID;
    v_finance_id UUID;
    v_director_id UUID;
BEGIN
    -- Get role IDs
    SELECT id INTO v_hr_manager_id FROM roles WHERE slug = 'hr_manager' LIMIT 1;
    SELECT id INTO v_finance_id FROM roles WHERE slug = 'finance' LIMIT 1;
    SELECT id INTO v_director_id FROM roles WHERE slug = 'director';
    IF NOT FOUND THEN
        -- Jika tidak ada role director, cari super_admin sebagai fallback
        SELECT id INTO v_director_id FROM roles WHERE slug = 'super_admin' LIMIT 1;
    END IF;

    -- ==================== LEAVE WORKFLOW ====================
    INSERT INTO approval_workflows (entity_type, name, description)
    VALUES ('leave', 'Cuti Tahunan & Izin', 'Workflow approval untuk cuti dan izin: Level 1 = Approval Line, Level 2 = HR Manager (jika > 3 hari), Level 3 = Direktur (jika > 7 hari)')
    RETURNING id INTO v_leave_id;

    -- Step 1: Approval Line (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_leave_id, 1, 'approval_line', 48);

    -- Step 2: HR Manager (if total_days > 3)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, condition_field, condition_operator, condition_value, escalation_hours)
    VALUES (v_leave_id, 2, 'hr_manager', 'total_days', '>', 3, 48);

    -- Step 3: Director (if total_days > 7)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, condition_field, condition_operator, condition_value, escalation_hours)
    VALUES (v_leave_id, 3, 'director', 'total_days', '>', 7, 24);

    -- ==================== OVERTIME WORKFLOW ====================
    INSERT INTO approval_workflows (entity_type, name, description)
    VALUES ('overtime', 'Lembur', 'Workflow approval untuk lembur: Level 1 = Approval Line, Level 2 = HR Manager')
    RETURNING id INTO v_overtime_id;

    -- Step 1: Approval Line (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_overtime_id, 1, 'approval_line', 24);

    -- Step 2: HR Manager (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_overtime_id, 2, 'hr_manager', 48);

    -- ==================== REIMBURSEMENT WORKFLOW ====================
    INSERT INTO approval_workflows (entity_type, name, description)
    VALUES ('reimbursement', 'Reimbursement', 'Workflow approval untuk reimbursement: Level 1 = Approval Line, Level 2 = HR Manager (jika > 1jt), Level 3 = Direktur (jika > 5jt)')
    RETURNING id INTO v_reimbursement_id;

    -- Step 1: Approval Line (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_reimbursement_id, 1, 'approval_line', 48);

    -- Step 2: HR Manager (if amount > 1000000)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, condition_field, condition_operator, condition_value, escalation_hours)
    VALUES (v_reimbursement_id, 2, 'hr_manager', 'amount', '>', 1000000, 48);

    -- Step 3: Director (if amount > 5000000)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, condition_field, condition_operator, condition_value, escalation_hours)
    VALUES (v_reimbursement_id, 3, 'director', 'amount', '>', 5000000, 24);

    -- ==================== SHIFT CHANGE WORKFLOW ====================
    INSERT INTO approval_workflows (entity_type, name, description)
    VALUES ('shift_change', 'Perubahan Shift', 'Workflow approval untuk perubahan shift: Level 1 = Approval Line')
    RETURNING id INTO v_shift_change_id;

    -- Step 1: Approval Line (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_shift_change_id, 1, 'approval_line', 24);

    -- ==================== LOAN WORKFLOW ====================
    INSERT INTO approval_workflows (entity_type, name, description)
    VALUES ('loan', 'Pinjaman Karyawan', 'Workflow approval untuk pinjaman: Level 1 = Approval Line, Level 2 = HR Manager, Level 3 = Finance, Level 4 = Direktur (jika > 10jt)')
    RETURNING id INTO v_loan_id;

    -- Step 1: Approval Line (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_loan_id, 1, 'approval_line', 48);

    -- Step 2: HR Manager (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_loan_id, 2, 'hr_manager', 48);

    -- Step 3: Finance (always)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, escalation_hours)
    VALUES (v_loan_id, 3, 'finance', 48);

    -- Step 4: Director (if amount > 10000000)
    INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, condition_field, condition_operator, condition_value, escalation_hours)
    VALUES (v_loan_id, 4, 'director', 'amount', '>', 10000000, 24);

END $$;

-- ─── 4. Approval Request Tracking ─────────────────────────────
-- Tabel untuk melacak status approval multi-level
CREATE TABLE IF NOT EXISTS approval_request_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,      -- 'leave', 'overtime', etc.
    entity_id UUID NOT NULL,               -- ID dari entity (leave_requests.id, etc.)
    workflow_id UUID NOT NULL REFERENCES approval_workflows(id),
    current_step INTEGER DEFAULT 1,        -- Step yang sedang aktif (0 = all done, -1 = rejected)
    total_steps INTEGER NOT NULL,          -- Total steps yang diperlukan (setelah evaluasi kondisi)
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(entity_type, entity_id)
);

CREATE INDEX idx_art_status ON approval_request_tracking(status);
CREATE INDEX idx_art_entity ON approval_request_tracking(entity_type, entity_id);
CREATE INDEX idx_art_current_step ON approval_request_tracking(current_step);

-- ─── 5. Add approval_config to HRSettings comment ─────────────
-- NOTE: approval_config is stored in companies.approval_config JSONB
-- which was added in migration 00002_companies.sql

COMMENT ON TABLE approval_workflows IS 'Konfigurasi workflow approval multi-level per entity type';
COMMENT ON TABLE approval_workflow_steps IS 'Step-step dalam workflow approval dengan kondisi dan eskalasi';
COMMENT ON TABLE approval_request_tracking IS 'Tracking status approval multi-level untuk setiap request';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS approval_request_tracking CASCADE;
DROP TABLE IF EXISTS approval_workflow_steps CASCADE;
DROP TABLE IF EXISTS approval_workflows CASCADE;
-- +goose StatementEnd
