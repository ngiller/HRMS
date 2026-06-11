-- +goose Up
-- ============================================================
-- Migration 00012: Employee Documents
-- ============================================================

CREATE TABLE employee_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    doc_type        doc_type NOT NULL,
    
    -- File metadata
    file_name       VARCHAR(255) NOT NULL,
    file_url        TEXT NOT NULL,                           -- Encrypted storage path (S3/MinIO)
    file_size       INTEGER,                                 -- In bytes
    mime_type       VARCHAR(100),
    
    -- Description
    title           VARCHAR(255),
    description     TEXT,
    
    -- Verification
    status          doc_status NOT NULL DEFAULT 'pending',
    verified_by     UUID REFERENCES employees(id) ON DELETE SET NULL,
    verified_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- Expiry
    expiry_date     DATE,                                    -- Untuk KTP, kontrak, dll
    is_required     BOOLEAN DEFAULT FALSE,
    
    -- For face recognition
    is_face_reference BOOLEAN DEFAULT FALSE,                 -- Foto referensi untuk face detection
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_employee_documents_updated_at
    BEFORE UPDATE ON employee_documents
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Indexes
CREATE INDEX idx_employee_documents_employee_id ON employee_documents(employee_id);
CREATE INDEX idx_employee_documents_type ON employee_documents(doc_type);
CREATE INDEX idx_employee_documents_status ON employee_documents(status);
CREATE INDEX idx_employee_documents_expiry ON employee_documents(expiry_date) WHERE expiry_date IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS employee_documents;
