-- Migration 00041: Add RBAC permissions for new modules
-- Description: Add reprimand, daily_journal, and report module permissions to roles
-- Modules: reprimand, daily_journal, report

-- +goose Up
-- +goose StatementBegin

-- Disable audit trigger to avoid 'unrecognized configuration parameter app.current_user_id'
ALTER TABLE roles DISABLE TRIGGER audit_roles;

-- Add permissions to super_admin role
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{reprimand}',
    '{"read": true, "create": true, "update": true, "delete": true}'::jsonb
)
WHERE slug = 'super_admin';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{daily_journal}',
    '{"read": true, "create": true, "update": true, "delete": true}'::jsonb
)
WHERE slug = 'super_admin';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{report}',
    '{"read": true, "create": true, "update": true, "delete": true}'::jsonb
)
WHERE slug = 'super_admin';

-- Add to hr_manager role: reprimand (create/update), daily_journal (read/update), report (read)
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{reprimand}',
    '{"read": true, "create": true, "update": true}'::jsonb
)
WHERE slug = 'hr_manager';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{daily_journal}',
    '{"read": true, "create": true, "update": true}'::jsonb
)
WHERE slug = 'hr_manager';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{report}',
    '{"read": true}'::jsonb
)
WHERE slug = 'hr_manager';

-- Add to employee role: daily_journal (create their own), reprimand (read their own)
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{daily_journal}',
    '{"read": true, "create": true}'::jsonb
)
WHERE slug = 'employee';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{reprimand}',
    '{"read": true}'::jsonb
)
WHERE slug = 'employee';

-- Add to hr_staff role same as hr_manager
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{reprimand}',
    '{"read": true, "create": true, "update": true}'::jsonb
)
WHERE slug = 'hr_staff';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{daily_journal}',
    '{"read": true, "update": true}'::jsonb
)
WHERE slug = 'hr_staff';

UPDATE roles
SET permissions = jsonb_set(
    permissions,
    '{report}',
    '{"read": true}'::jsonb
)
WHERE slug = 'hr_staff';

-- Add to finance role: report (read)
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{report}',
    '{"read": true}'::jsonb
)
WHERE slug = 'finance';

-- Re-enable audit trigger
ALTER TABLE roles ENABLE TRIGGER audit_roles;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove reprimand permissions
UPDATE roles SET permissions = permissions #- '{reprimand}' WHERE slug IN ('super_admin', 'hr_manager', 'hr_staff', 'employee');
-- Remove daily_journal permissions
UPDATE roles SET permissions = permissions #- '{daily_journal}' WHERE slug IN ('super_admin', 'hr_manager', 'hr_staff', 'employee');
-- Remove report permissions
UPDATE roles SET permissions = permissions #- '{report}' WHERE slug IN ('super_admin', 'hr_manager', 'hr_staff', 'finance');
-- +goose StatementEnd
