-- Migration 00049: Add approval_trail column to manual_attendance_requests
-- Required for approval workflow integration (same pattern as leave_requests, overtime_requests, etc.)

-- +goose Up
ALTER TABLE manual_attendance_requests
ADD COLUMN IF NOT EXISTS approval_trail jsonb DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE manual_attendance_requests
DROP COLUMN IF EXISTS approval_trail;
