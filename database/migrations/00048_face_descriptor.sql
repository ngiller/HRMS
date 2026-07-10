-- +goose Up
-- ============================================================
-- Migration 00045: Face Descriptor for Face Recognition
-- ============================================================
-- Stores face embeddings (128-dim Float32Array as JSON) for
-- employee face verification during attendance check-in/out.
-- ============================================================

ALTER TABLE employees
    ADD COLUMN face_descriptor JSONB DEFAULT NULL,
    ADD COLUMN face_descriptor_updated_at TIMESTAMPTZ DEFAULT NULL;

-- +goose Down
ALTER TABLE employees
    DROP COLUMN IF EXISTS face_descriptor,
    DROP COLUMN IF EXISTS face_descriptor_updated_at;
