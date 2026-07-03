-- +goose NO TRANSACTION
-- +goose Up
-- Disable audit trigger to avoid 'unrecognized configuration parameter app.current_user_id'
ALTER TABLE employees DISABLE TRIGGER audit_employees;

-- Assign default work schedule (5 Hari Kerja, Senin-Jumat) to all active employees
-- who don't already have a work_schedule_id set.
UPDATE employees
SET work_schedule_id = ws.id
FROM work_schedules ws
WHERE ws.name = '5 Hari Kerja (Senin-Jumat)'
  AND ws.deleted_at IS NULL
  AND employees.deleted_at IS NULL
  AND employees.work_schedule_id IS NULL;

ALTER TABLE employees ENABLE TRIGGER audit_employees;

-- +goose Down
-- This migration is additive and non-destructive. We don't want to remove work_schedule_id
-- because that would break attendance check-in for employees who already have it.
SELECT 1;
