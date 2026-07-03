-- +goose Up
-- ============================================================
-- Migration 00030: Add 'paid' to leave_status enum
-- ============================================================
-- Reimbursement PayReimbursement mengupdate status ke 'paid'.
-- Tapi leave_status enum tidak punya nilai 'paid'.
-- ALTER TYPE ... ADD VALUE tidak bisa dijalankan dalam transaction block.
-- ============================================================

ALTER TYPE leave_status ADD VALUE IF NOT EXISTS 'paid';

-- +goose Down
-- Removing values from enums is not supported in PostgreSQL.
-- A new type would need to be created and all columns migrated.
-- This migration is one-way; if you need to roll back, restore from backup.
