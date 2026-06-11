-- +goose Up
-- ============================================================
-- Migration 00003: Organization Structure
-- ============================================================

-- ============================================================
-- Position Grades / Levels
-- ============================================================
CREATE TABLE position_grades (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,   -- Staff, Senior, Supervisor, Manager, GM, Direktur
    level       INTEGER NOT NULL UNIQUE, -- 1-10
    description TEXT,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_position_grades_updated_at
    BEFORE UPDATE ON position_grades
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Departments (Hierarchical - self-referencing)
-- ============================================================
CREATE TABLE departments (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50) UNIQUE NOT NULL,
    parent_id   UUID REFERENCES departments(id) ON DELETE SET NULL,
    head_id     UUID,                  -- FK ke employees (added after employees table)
    description TEXT,
    is_active   BOOLEAN DEFAULT TRUE,
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE TRIGGER set_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Positions
-- ============================================================
CREATE TABLE positions (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name              VARCHAR(255) NOT NULL,
    department_id     UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    grade_id          UUID REFERENCES position_grades(id) ON DELETE SET NULL,
    description       TEXT,
    is_active         BOOLEAN DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE TRIGGER set_positions_updated_at
    BEFORE UPDATE ON positions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Seed default position grades
-- ============================================================
INSERT INTO position_grades (name, level, description) VALUES
    ('Staff',         1, 'Karyawan pelaksana / staf'),
    ('Senior Staff',  2, 'Karyawan senior dengan pengalaman'),
    ('Supervisor',    3, 'Pengawas tim kecil'),
    ('Senior Supervisor', 4, 'Pengawas tim menengah'),
    ('Assistant Manager', 5, 'Asisten manajer departemen'),
    ('Manager',       6, 'Kepala departemen'),
    ('Senior Manager',7, 'Kepala departemen senior / multi-departemen'),
    ('General Manager',8, 'Kepala divisi / direktorat'),
    ('Director',      9, 'Direktur'),
    ('President Director', 10, 'Direktur Utama / CEO');

-- Indexes
CREATE INDEX idx_departments_parent_id ON departments(parent_id);
CREATE INDEX idx_departments_code ON departments(code);
CREATE INDEX idx_positions_department_id ON positions(department_id);
CREATE INDEX idx_positions_grade_id ON positions(grade_id);

-- +goose Down
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS position_grades;
