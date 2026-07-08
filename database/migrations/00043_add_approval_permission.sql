-- Migration 00043: Add approval module permission to roles
-- Description: Add approval module with read + approve actions for roles that need to approve requests
-- Modules: approval

-- +goose Up
-- +goose StatementBegin

-- Disable audit trigger to avoid 'unrecognized configuration parameter app.current_user_id'
ALTER TABLE roles DISABLE TRIGGER audit_roles;

-- super_admin: full access to approval workflow
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{approval}',
    '{"read": true, "approve": true}'::jsonb
)
WHERE slug = 'super_admin';

-- hr_manager: can view and approve
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{approval}',
    '{"read": true, "approve": true}'::jsonb
)
WHERE slug = 'hr_manager';

-- hr_staff: can view pending approvals
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{approval}',
    '{"read": true, "approve": true}'::jsonb
)
WHERE slug = 'hr_staff';

-- finance: can view and approve (for reimbursement and loan approvals)
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{approval}',
    '{"read": true, "approve": true}'::jsonb
)
WHERE slug = 'finance';

-- director: can view and approve
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{approval}',
    '{"read": true, "approve": true}'::jsonb
)
WHERE slug = 'director';

-- department_head: can view and approve (approves their department's requests)
UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{approval}',
    '{"read": true, "approve": true}'::jsonb
)
WHERE slug = 'department_head';

-- Re-enable audit trigger
ALTER TABLE roles ENABLE TRIGGER audit_roles;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove approval permissions
UPDATE roles SET permissions = permissions #- '{approval}'
WHERE slug IN ('super_admin', 'hr_manager', 'hr_staff', 'finance', 'director', 'department_head');
-- +goose StatementEnd
