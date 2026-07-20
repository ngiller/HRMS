-- +goose Up
-- ============================================================
-- Migration 00050: Fix Employee KPI Permissions
-- ============================================================
-- Karyawan (employee) seharusnya TIDAK memiliki akses ke modul KPI.
-- Modul KPI adalah untuk HR / Manager / Direktur, bukan untuk karyawan biasa.
--
-- Latar belakang:
-- Migration 00016 seed role 'employee' dengan kpi: {create: true, read: true}
-- yang tidak sesuai standar. Karyawan tidak boleh akses management KPI.
--
-- Note: File ini awalnya 00048, diganti karena conflict dengan
-- 00048_face_descriptor.sql yang sudah ada di repo.
-- ============================================================

-- Disable audit trigger to avoid 'unrecognized configuration parameter app.current_user_id'
ALTER TABLE roles DISABLE TRIGGER audit_roles;

-- Hapus kpi module dari role 'employee' (#- operator aman walau key tidak ada)
UPDATE roles
SET permissions = permissions #- '{kpi}'
WHERE slug = 'employee';

ALTER TABLE roles ENABLE TRIGGER audit_roles;

-- +goose Down
-- ============================================================
-- Rollback: Kembalikan kpi permissions untuk role employee
-- ============================================================

ALTER TABLE roles DISABLE TRIGGER audit_roles;

UPDATE roles
SET permissions = jsonb_set(
    COALESCE(permissions, '{}'::jsonb),
    '{kpi}',
    '{"create": true, "read": true}'::jsonb
)
WHERE slug = 'employee';

ALTER TABLE roles ENABLE TRIGGER audit_roles;
