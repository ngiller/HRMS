-- +goose Up
-- ============================================================
-- Migration 00034: Add attendance create/update permissions to roles
-- ============================================================
ALTER TABLE roles DISABLE TRIGGER audit_roles;

UPDATE roles 
SET permissions = permissions || '{
    "attendance": {"create": true, "read": true, "update": true}
}'::jsonb
WHERE slug IN ('employee', 'manager', 'director', 'finance');

ALTER TABLE roles ENABLE TRIGGER audit_roles;

-- +goose Down
ALTER TABLE roles DISABLE TRIGGER audit_roles;

UPDATE roles 
SET permissions = permissions || '{
    "attendance": {"read": true}
}'::jsonb
WHERE slug IN ('employee', 'manager', 'director', 'finance');

ALTER TABLE roles ENABLE TRIGGER audit_roles;
