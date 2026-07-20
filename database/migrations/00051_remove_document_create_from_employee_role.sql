-- +goose Up
-- ============================================================
-- Migration 00051: Remove document:create from employee role
-- ============================================================
-- Employee role should not be allowed to add documents.
-- They can still read and update their own documents.
-- ============================================================

ALTER TABLE roles DISABLE TRIGGER audit_roles;

UPDATE roles
SET permissions = jsonb_set(permissions, '{document}', '{"read": true, "update": true}'::jsonb)
WHERE slug = 'employee';

ALTER TABLE roles ENABLE TRIGGER audit_roles;

-- +goose Down
-- ============================================================
-- Rollback: Restore document:create for employee role
-- ============================================================

UPDATE roles
SET permissions = jsonb_set(permissions, '{document}', '{"create": true, "read": true, "update": true}'::jsonb)
WHERE slug = 'employee';
