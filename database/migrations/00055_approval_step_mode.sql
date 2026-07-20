-- ================================================================
-- Migration 00054: Approval Step Mode (Parallel / Any Approver)
-- Menambahkan step_mode pada approval_workflow_steps untuk
-- mendukung mode approval paralel (siapa saja bisa approve)
-- ================================================================

-- +goose Up
-- +goose StatementBegin

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='approval_workflow_steps' AND column_name='step_mode'
    ) THEN
        ALTER TABLE approval_workflow_steps ADD COLUMN step_mode VARCHAR(20) NOT NULL DEFAULT 'single';
    END IF;
END $$;

COMMENT ON COLUMN approval_workflow_steps.step_mode IS
  'Mode approval: single (default, 1 approver spesifik) atau any (salah satu dari approver yang cocok bisa approve)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE approval_workflow_steps DROP COLUMN IF EXISTS step_mode;

-- +goose StatementEnd
