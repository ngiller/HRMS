-- +goose Up
-- ============================================================
-- Migration 00001: Enable pgcrypto & Create Encryption Helpers
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Encryption Configuration
-- Data sensitif dienkripsi menggunakan pgp_sym_encrypt dengan
-- AES256. Kunci enkripsi disimpan di environment variable
-- (ENCRYPTION_KEY), bukan di database.
--
-- Untuk keamanan lebih tinggi (AES-256-GCM), enkripsi dapat
-- dilakukan di layer aplikasi Go sebelum disimpan ke database.
-- ============================================================

-- Fungsi untuk mengenkripsi data sensitif
-- Key diambil dari current_setting('app.encryption_key')
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION encrypt_sensitive(data TEXT)
RETURNS BYTEA AS $$
BEGIN
    IF data IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN pgp_sym_encrypt(
        data,
        current_setting('app.encryption_key'),
        'cipher-algo=aes256, compress-algo=2'
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
-- +goose StatementEnd

-- Fungsi untuk mendekripsi data sensitif
-- Key diambil dari current_setting('app.encryption_key')
-- Jika key salah atau data corrupt, return NULL (dengan warning) agar query tidak gagal.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION decrypt_sensitive(encrypted_data BYTEA)
RETURNS TEXT AS $$
BEGIN
    IF encrypted_data IS NULL THEN
        RETURN NULL;
    END IF;
    BEGIN
        RETURN pgp_sym_decrypt(
            encrypted_data,
            current_setting('app.encryption_key'),
            'cipher-algo=aes256, compress-algo=2'
        );
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING 'decrypt_sensitive failed: %, SQLSTATE: %', SQLERRM, SQLSTATE;
        RETURN NULL;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
-- +goose StatementEnd

-- ============================================================
-- Generic updated_at trigger function
-- Digunakan oleh semua tabel. HARUS ada sebelum migrasi tabel.
-- ============================================================
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- ============================================================
-- ENUM Types
-- ============================================================

CREATE TYPE employment_status AS ENUM (
    'tetap',
    'kontrak',
    'percobaan',
    'harian'
);

CREATE TYPE gender_type AS ENUM (
    'laki_laki',
    'perempuan'
);

CREATE TYPE religion_type AS ENUM (
    'islam',
    'kristen',
    'katolik',
    'hindu',
    'buddha',
    'konghucu',
    'lainnya'
);

CREATE TYPE marital_status AS ENUM (
    'lajang',
    'menikah',
    'cerai_hidup',
    'cerai_mati'
);

CREATE TYPE ptkp_status AS ENUM (
    'TK0', 'TK1', 'TK2', 'TK3',
    'K0', 'K1', 'K2', 'K3',
    'KIT0', 'KIT1', 'KIT2', 'KIT3'
);

CREATE TYPE tax_method AS ENUM (
    'gross',
    'gross_up',
    'nett'
);

CREATE TYPE attendance_status AS ENUM (
    'hadir',
    'terlambat',
    'izin',
    'sakit',
    'tanpa_keterangan',
    'cuti',
    'libur'
);

CREATE TYPE leave_status AS ENUM (
    'pending',
    'approved',
    'rejected',
    'cancelled'
);

CREATE TYPE loan_status AS ENUM (
    'pending',
    'approved',
    'active',
    'completed',
    'rejected',
    'defaulted'
);

CREATE TYPE payroll_status AS ENUM (
    'draft',
    'completed',
    'approved',
    'paid'
);

CREATE TYPE kpi_review_status AS ENUM (
    'draft',
    'self_review',
    'manager_review',
    'hr_review',
    'completed'
);

CREATE TYPE reprimand_type AS ENUM (
    'verbal',
    'sp1',
    'sp2',
    'sp3'
);

CREATE TYPE reprimand_status AS ENUM (
    'issued',
    'acknowledged',
    'expired'
);

CREATE TYPE announcement_type AS ENUM (
    'general',
    'important',
    'emergency'
);

CREATE TYPE holiday_type AS ENUM (
    'national',
    'joint',
    'company'
);

CREATE TYPE doc_status AS ENUM (
    'pending',
    'verified',
    'rejected'
);

CREATE TYPE doc_type AS ENUM (
    'ktp',
    'kk',
    'ijazah',
    'sertifikat',
    'kontrak',
    'npwp',
    'bpjs',
    'photo',
    'other'
);

CREATE TYPE loan_payment_method AS ENUM (
    'payroll_deduction',
    'manual_transfer'
);

CREATE TYPE notification_type AS ENUM (
    'approval_request',
    'approved',
    'rejected',
    'announcement',
    'reminder',
    'system'
);

CREATE TYPE work_schedule_type AS ENUM (
    '5_day',
    '6_day',
    'shift'
);

CREATE TYPE reimbursement_type AS ENUM (
    'medical',
    'travel',
    'training',
    'supplies',
    'other'
);

CREATE TYPE overtime_type AS ENUM (
    'weekday',
    'weekend',
    'holiday'
);

CREATE TYPE loan_type AS ENUM (
    'regular',
    'emergency',
    'education'
);

-- +goose Down
DROP TYPE IF EXISTS loan_type;
DROP TYPE IF EXISTS overtime_type;
DROP TYPE IF EXISTS reimbursement_type;
DROP TYPE IF EXISTS work_schedule_type;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS loan_payment_method;
DROP TYPE IF EXISTS doc_type;
DROP TYPE IF EXISTS doc_status;
DROP TYPE IF EXISTS holiday_type;
DROP TYPE IF EXISTS announcement_type;
DROP TYPE IF EXISTS reprimand_status;
DROP TYPE IF EXISTS reprimand_type;
DROP TYPE IF EXISTS kpi_review_status;
DROP TYPE IF EXISTS payroll_status;
DROP TYPE IF EXISTS loan_status;
DROP TYPE IF EXISTS leave_status;
DROP TYPE IF EXISTS attendance_status;
DROP TYPE IF EXISTS tax_method;
DROP TYPE IF EXISTS ptkp_status;
DROP TYPE IF EXISTS marital_status;
DROP TYPE IF EXISTS religion_type;
DROP TYPE IF EXISTS gender_type;
DROP TYPE IF EXISTS employment_status;

DROP FUNCTION IF EXISTS decrypt_sensitive;
DROP FUNCTION IF EXISTS encrypt_sensitive;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS pgcrypto;
