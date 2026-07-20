-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees
	ADD COLUMN IF NOT EXISTS is_pregnant BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN employees.is_pregnant IS 'Menandakan apakah karyawan sedang hamil (untuk validasi cuti melahirkan)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees
	DROP COLUMN IF EXISTS is_pregnant;
-- +goose StatementEnd
