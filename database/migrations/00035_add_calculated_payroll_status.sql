-- +goose NO TRANSACTION
-- +goose Up
-- Add 'calculated' to payroll_status enum for payroll_items calculation workflow
ALTER TYPE payroll_status ADD VALUE IF NOT EXISTS 'calculated';

-- +goose Down
-- Cannot remove enum values in PostgreSQL, this is a no-op down
-- To downgrade, you would need to recreate the type, which is complex.
-- Instead, keep the value but it won't be used when downgrading other schema changes.
SELECT 1;
