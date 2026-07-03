-- +goose Up
-- ============================================================
-- Migration 00023: Add missing module permissions to super_admin
-- ============================================================
-- Modul-modul yang ditambahkan setelah seed awal di 00016
-- belum tercantum di permissions role super_admin.
-- Gunakan JSONB concatenation (||) agar tidak menimpa
-- permissions existing.
-- ============================================================

ALTER TABLE roles DISABLE TRIGGER audit_roles;

UPDATE roles SET permissions = permissions || '{
    "position_grade": {"create": true, "read": true, "update": true, "delete": true},
    "position": {"create": true, "read": true, "update": true, "delete": true},
    "work_schedule": {"create": true, "read": true, "update": true, "delete": true},
    "attendance_location": {"create": true, "read": true, "update": true, "delete": true},
    "shift_change": {"create": true, "read": true, "update": true, "delete": true}
}'::jsonb
WHERE slug = 'super_admin';

ALTER TABLE roles ENABLE TRIGGER audit_roles;

-- +goose Down
ALTER TABLE roles DISABLE TRIGGER audit_roles;
UPDATE roles SET permissions = permissions - 'position_grade' - 'position' - 'work_schedule' - 'attendance_location' - 'shift_change'
WHERE slug = 'super_admin';
ALTER TABLE roles ENABLE TRIGGER audit_roles;
