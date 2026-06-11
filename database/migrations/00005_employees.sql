-- +goose Up
-- ============================================================
-- Migration 00005: Employees (Master Data Karyawan)
-- ============================================================
-- Data sensitif dienkripsi menggunakan pgp_sym_encrypt.
-- Kunci enkripsi disimpan di environment variable.
-- ============================================================

CREATE TABLE employees (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id             VARCHAR(50) UNIQUE NOT NULL,  -- NIP / NIK Perusahaan
    full_name               VARCHAR(255) NOT NULL,

    -- Data Pribadi (terenkripsi)
    encrypted_nik           BYTEA,                         -- NIK KTP
    encrypted_npwp          BYTEA,                         -- NPWP
    encrypted_bank_account  BYTEA,                         -- Nomor Rekening
    encrypted_bank_name     BYTEA,                         -- Nama Bank
    birth_place             VARCHAR(100),
    birth_date              DATE,
    gender                  gender_type NOT NULL,
    religion                religion_type,
    marital_status          marital_status DEFAULT 'lajang',
    dependent_count         INTEGER DEFAULT 0,             -- Jumlah tanggungan
    encrypted_address_ktp   BYTEA,                         -- Alamat KTP
    address_domicile        TEXT,                          -- Alamat domisili (tidak dienkripsi)
    phone                   VARCHAR(50),
    email                   VARCHAR(255),
    blood_type              VARCHAR(5),

    -- Data Pekerjaan
    join_date               DATE NOT NULL,
    end_date                DATE,                          -- Untuk kontrak / resign
    employment_status       employment_status NOT NULL DEFAULT 'percobaan',
    nik_perusahaan          VARCHAR(50),                   -- NIK Karyawan (internal)

    -- Relasi Organisasi
    department_id           UUID REFERENCES departments(id) ON DELETE SET NULL,
    position_id             UUID REFERENCES positions(id) ON DELETE SET NULL,
    position_grade_id       UUID REFERENCES position_grades(id) ON DELETE SET NULL,
    approval_line_id        UUID REFERENCES employees(id) ON DELETE SET NULL,

    -- Pajak & Keuangan
    ptkp_status             ptkp_status DEFAULT 'TK0',
    tax_method              tax_method DEFAULT 'gross',

    -- Absensi
    work_schedule_id        UUID REFERENCES work_schedules(id) ON DELETE SET NULL,
    is_remote               BOOLEAN DEFAULT FALSE,         -- Boleh absen tanpa GPS
    face_embedding          JSONB,                         -- Vector untuk face recognition (Float32Array)

    -- Status
    photo_url               TEXT,
    is_active               BOOLEAN DEFAULT TRUE,
    resigned_at             TIMESTAMPTZ,
    resign_reason           TEXT,

    -- Audit Trail
    created_by              UUID REFERENCES employees(id) ON DELETE SET NULL,
    updated_by              UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE TRIGGER set_employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Employee History (Riwayat perubahan data penting)
-- ============================================================
CREATE TABLE employee_histories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    change_type     VARCHAR(50) NOT NULL,      -- promotion, mutation, salary_change, status_change, department_change
    old_value       JSONB,
    new_value       JSONB,
    reason          TEXT,
    changed_by      UUID REFERENCES employees(id) ON DELETE SET NULL,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Employee Emergency Contacts
-- ============================================================
CREATE TABLE employee_emergency_contacts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    relationship    VARCHAR(100) NOT NULL,
    phone           VARCHAR(50) NOT NULL,
    address         TEXT,
    is_primary      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_employee_emergency_contacts_updated_at
    BEFORE UPDATE ON employee_emergency_contacts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Employee Salary History (Riwayat perubahan gaji)
-- ============================================================
CREATE TABLE employee_salary_histories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    base_salary     DECIMAL(15,2) NOT NULL,
    daily_wage      DECIMAL(15,2),
    effective_date  DATE NOT NULL,
    reason          TEXT,
    changed_by      UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_employees_employee_id ON employees(employee_id);
CREATE INDEX idx_employees_department_id ON employees(department_id);
CREATE INDEX idx_employees_position_id ON employees(position_id);
CREATE INDEX idx_employees_approval_line_id ON employees(approval_line_id);
CREATE INDEX idx_employees_status ON employees(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_employees_join_date ON employees(join_date);
CREATE INDEX idx_employee_histories_employee_id ON employee_histories(employee_id);
CREATE INDEX idx_employee_salary_histories_employee_id ON employee_salary_histories(employee_id);

-- +goose Down
DROP TABLE IF EXISTS employee_salary_histories;
DROP TABLE IF EXISTS employee_emergency_contacts;
DROP TABLE IF EXISTS employee_histories;
DROP TABLE IF EXISTS employees;
