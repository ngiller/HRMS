# Database Schema Design — HRMS Application

**Teknologi:** PostgreSQL 16+ | **Migration Tool:** Goose | **Generator:** sqlc

---

## 1. Entity Relationship Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ORGANIZATION                                    │
│  companies ──1:N──> departments ──1:N──> positions ──N:1──> position_grades  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              EMPLOYEES (Master)                              │
│  employees ──── (encrypted: NIK, NPWP, bank_account, address_ktp)           │
│     │                                                                        │
│     ├──1:N──> employee_histories                                             │
│     ├──1:N──> employee_emergency_contacts                                    │
│     ├──1:N──> employee_salary_histories                                      │
│     ├──1:N──> employee_documents                                             │
│     └──1:N──> notification_preferences                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
          ┌─────────────────────────┼─────────────────────────┐
          ▼                         ▼                         ▼
┌──────────────────┐   ┌────────────────────┐   ┌─────────────────────┐
│   ATTENDANCE      │   │     LEAVE           │   │     PAYROLL         │
│ attendance_locations│   │ leave_types         │   │ payroll_periods     │
│ work_schedules    │   │ leave_balances      │   │ payroll_items       │
│ attendance_records│   │ leave_requests      │   │                     │
│ manual_attendance │   │                    │   │                     │
└──────────────────┘   └────────────────────┘   └─────────────────────┘
                                                    │
          ┌──────────────┼────────────────┐          │
          ▼              ▼                ▼          ▼
┌──────────────┐ ┌────────────┐ ┌──────────────┐ ┌──────────────┐
│ FINANCIAL     │ │  KPI       │ │  HR          │ │  NOTIF       │
│ reimbursements│ │ kpi_templates│ │ reprimands   │ │ notifications│
│ overtime_reqs │ │ kpi_indicators││ announcements│ │ user_sessions│
│ loans         │ │ kpi_reviews│ │ company_holidays││ login_attempts│
│ loan_installments│ │           │ │              │ │ activity_logs│
└──────────────┘ └────────────┘ └──────────────┘ └──────────────┘
```

---

## 2. Migration Details

| No | File | Tabel Dibuat | Keterangan |
|:--:|------|-------------|------------|
| 1 | `00001_extension_and_encryption.sql` | — | pgcrypto, ENUM types (29 enums), encrypt/decrypt functions |
| 2 | `00002_companies.sql` | companies | Settings perusahaan, HR config, approval flow config |
| 3 | `00003_organization.sql` | position_grades, departments, positions | Struktur organisasi dengan self-referencing department |
| 4 | `00004_schedules_and_locations.sql` | work_schedules, attendance_locations | Schedule 5/6 hari, GPS locations, seed default schedules |
| 5 | `00005_employees.sql` | employees, employee_histories, employee_emergency_contacts, employee_salary_histories | **Table utama** dengan encrypted columns + history |
| 6 | `00006_attendance_records.sql` | attendance_records, manual_attendance_requests, attendance_summary | Check-in/out, GPS, face detection, materialized view |
| 7 | `00007_leave_management.sql` | leave_types, leave_balances, leave_requests | 10 jenis cuti, approval trail JSON, seed leave types |
| 8 | `00008_reimbursements_overtime.sql` | reimbursements, overtime_requests, overtime_calculation | View perhitungan lembur 1.5x-4x |
| 9 | `00009_loans.sql` | loans, loan_installments | Pinjaman dengan cicilan, payroll deduction |
| 10 | `00010_kpi_performance.sql` | kpi_templates, kpi_indicators, kpi_reviews | Auto-calculate score & category, 3 tahap review |
| 11 | `00011_reprimands_announcements_holidays.sql` | reprimands, announcements, announcement_reads, company_holidays | SP1-3, escalation, seed 26 holidays 2026 |
| 12 | `00012_employee_documents.sql` | employee_documents | Dokumen dengan verifikasi & expiry |
| 13 | `00013_payroll.sql` | payroll_periods, payroll_items, payslip_view | BPJS, PPh 21, THR, tunjangan, company cost |
| 14 | `00014_activity_logs_notifications_users.sql` | activity_logs, notifications, notification_preferences, user_sessions, password_reset_tokens, login_attempts | Audit trail, notification, auth |
| 15 | `00015_final_indexes_and_triggers.sql` | — | Indexes, triggers, payroll function, reporting views |
| 16 | `00016_employee_auth_and_department_schedules.sql` | roles, auth columns on employees, work_schedule_id on departments | **Role-based access control, employee login (password_hash, role_id), department-level work schedule assignment** |
| 17 | `00017_employee_salary_components.sql` | employee_salary_components, employee_salary_component_histories | **Per-employee salary component master data (allowances & deductions) with auto-change history tracking via trigger** |
| 18 | `00018_flexible_overtime_rates.sql` | position_grade_overtime_rates, employee_overtime_rates, updated overtime_calculation view | **3-level flexible overtime rate configuration (company default → position grade → per employee), updated calculation view** |
| 19 | `00019_flexible_bpjs_config.sql` | bpjs_config on employees, updated calculate_employee_payroll() | **Flexible BPJS configuration — semua rate & ceiling bisa dikonfigurasi via hr_settings, plus override per employee via bpjs_config JSONB** |
| 20 | `00020_shift_change_requests.sql` | shift_change_requests, 2 new ENUMs | **Shift Change Request — employee dapat mengajukan ganti shift individu atau tukar shift dengan approval workflow (pending → partner_pending → approved/rejected)** |

---

## 3. Autentikasi & Role-Based Access Control

### 3.1 Login & Autentikasi

Setiap **karyawan adalah user aplikasi** — tidak ada tabel `users` terpisah.

| Field di `employees` | Tipe | Keterangan |
|----------------------|------|------------|
| `email` | VARCHAR(255) UNIQUE | Digunakan sebagai username / login ID |
| `password_hash` | TEXT | Hash bcrypt/argon2 dari password |
| `role_id` | UUID (FK → roles) | Menentukan hak akses karyawan |
| `approval_line_id` | UUID (FK → employees) | Penanggung jawab approval untuk karyawan ini (NULL = puncak rantai approval) |
| `last_login_at` | TIMESTAMPTZ | Timestamp terakhir login |
| `is_locked` | BOOLEAN | Terkunci karena gagal login berulang |
| `locked_until` | TIMESTAMPTZ | Sampai kapan akun terkunci |

### 3.2 Role & Permissions

**7 Role Bawaan (System Role):**
| Role | Slug |
|------|------|
| Super Admin | `super_admin` |
| HR Manager | `hr_manager` |
| HR Staff | `hr_staff` |
| Finance | `finance` |
| Manager | `manager` |
| Karyawan | `employee` |
| Direktur | `director` |

Setiap role memiliki `permissions JSONB` yang mendefinisikan akses CRUD per modul:
```json
{
  "employee": {"create": true, "read": true, "update": true, "delete": true},
  "payroll": {"create": true, "read": true},
  ...
}
```

### 3.3 Department-Level Work Schedule

`departments.work_schedule_id` → FK ke `work_schedules(id)`
- Setiap departemen memiliki jadwal kerja default
- Karyawan mewarisi jadwal departemen kecuali memiliki override individual (`employees.work_schedule_id`)

**Hierarki resolusi:** `Individual override` → `Departemen default` → `Global default`

---

## 4. Enkripsi Data Sensitif

### 4.1 Data yang Dienkripsi

| Kolom | Tipe | Fungsi Enkripsi |
|-------|------|-----------------|
| `employees.encrypted_nik` | BYTEA | `encrypt_sensitive()` |
| `employees.encrypted_npwp` | BYTEA | `encrypt_sensitive()` |
| `employees.encrypted_bank_account` | BYTEA | `encrypt_sensitive()` |
| `employees.encrypted_bank_name` | BYTEA | `encrypt_sensitive()` |
| `employees.encrypted_address_ktp` | BYTEA | `encrypt_sensitive()` |

### 4.2 Metode Enkripsi

- **Algoritma:** AES-256 via pgp_sym_encrypt (OpenPGP)
- **Key Management:** ENCRYPTION_KEY disimpan di environment variable (bukan di database)
- **Fungsi:** `encrypt_sensitive(data)` dan `decrypt_sensitive(data)` — SECURITY DEFINER
- **Akses:** Hanya bisa di-decrypt oleh koneksi database yang memiliki `app.encryption_key` yang benar

> **Catatan:** Untuk keamanan lebih tinggi (AES-256-GCM murni), enkripsi/dekripsi bisa dilakukan di layer aplikasi Go sebelum data masuk database. Pendekatan hybrid memungkinkan: Go melakukan AES-256-GCM, sementara PostgreSQL tetap menyimpan ciphertext dalam BYTEA.

### 4.3 Key Rotation

Key rotation dilakukan dengan:
1. Set ENCRYPTION_KEY baru di environment variable
2. Jalankan query `UPDATE employees SET encrypted_nik = encrypt_sensitive(decrypt_sensitive(encrypted_nik), new_key)`
3. Hapus key lama

---

## 4. Fitur Database Penting

### 4.1 BPJS Configuration

Semua parameter BPJS (persentase iuran & ceiling) dikonfigurasi via `companies.hr_settings`:

```json
{
  "bpjs": {
    "kesehatan": {"employee_rate": 0.01, "company_rate": 0.04, "ceiling": 12000000},
    "jht":       {"employee_rate": 0.02, "company_rate": 0.037, "ceiling": null},
    "jp":        {"employee_rate": 0.01, "company_rate": 0.02, "ceiling": 10000000},
    "jkm":       {"company_rate": 0.003, "ceiling": null}
  }
}
```

Setiap karyawan bisa override via `employees.bpjs_config` (JSONB):
- `enabled: false` — nonaktifkan komponen untuk karyawan tertentu
- `employee_rate`, `company_rate` — override rate spesifik

Hierarki: `bpjs_config per employee` → `hr_settings.bpjs` → `Default pemerintah`

### 4.2 Audit Trail
- Semua tabel menggunakan `created_at`, `updated_at`, `deleted_at` (soft delete)
- `activity_logs` mencatat semua perubahan dengan `old_values` dan `new_values` dalam JSONB
- Trigger `audit_employees` otomatis mencatat perubahan ke `employees`
- IP address dan user_agent tercatat untuk setiap aksi

### 4.3 Soft Delete
- Semua tabel utama menggunakan `deleted_at TIMESTAMPTZ`
- Query harus menyertakan `WHERE deleted_at IS NULL`
- Data tidak pernah benar-benar dihapus (kecuali untuk GDPR/data privacy request)

### 4.4 Approval Trail
- Semua request (cuti, reimbursement, lembur, pinjaman) menggunakan `approval_trail JSONB`
- Format: `[{level, approver_id, status, note, date}]`
- fleksibel untuk approval bertingkat tanpa perlu join table approval

### 4.5 Auto Timestamps
- `trigger_set_updated_at()` otomatis mengisi `updated_at` pada setiap UPDATE
- `DEFAULT NOW()` untuk `created_at`

### 4.6 Reporting Views
- `v_employee_headcount` — Headcount per departemen
- `v_attendance_monthly` — Rekap absensi bulanan
- `v_leave_balance_summary` — Sisa cuti per karyawan
- `payslip_view` — Slip gaji dengan join employee & department
- `attendance_summary` — Materialized view untuk dashboard absensi

---

## 5. ENUM Types (29 Total)

| ENUM | Values |
|------|--------|
| `employment_status` | tetap, kontrak, percobaan, harian |
| `gender_type` | laki_laki, perempuan |
| `religion_type` | islam, kristen, katolik, hindu, buddha, konghucu, lainnya |
| `marital_status` | lajang, menikah, cerai_hidup, cerai_mati |
| `ptkp_status` | TK0-TK3, K0-K3, KIT0-KIT3 |
| `tax_method` | gross, gross_up, nett |
| `attendance_status` | hadir, terlambat, izin, sakit, tanpa_keterangan, cuti, libur |
| `leave_status` | pending, approved, rejected, cancelled |
| `loan_status` | pending, approved, active, completed, rejected, defaulted |
| `payroll_status` | draft, completed, approved, paid |
| `kpi_review_status` | draft, self_review, manager_review, hr_review, completed |
| `reprimand_type` | verbal, sp1, sp2, sp3 |
| `reprimand_status` | issued, acknowledged, expired |
| `announcement_type` | general, important, emergency |
| `holiday_type` | national, joint, company |
| `doc_status` | pending, verified, rejected |
| `doc_type` | ktp, kk, ijazah, sertifikat, kontrak, npwp, bpjs, photo, other |
| `loan_payment_method` | payroll_deduction, manual_transfer |
| `notification_type` | approval_request, approved, rejected, announcement, reminder, system |
| `work_schedule_type` | 5_day, 6_day, shift |
| `reimbursement_type` | medical, travel, training, supplies, other |
| `overtime_type` | weekday, weekend, holiday |
| `loan_type` | regular, emergency, education |

---

## 6. Index Strategy

### 6.1 Primary Lookup Indexes
- `employee_id` pada semua tabel yang berelasi dengan employees
- `department_id` pada employees dan positions
- `status` untuk filter cepat (pending, approved, dll.)

### 6.2 Date Range Indexes
- `attendance_records(employee_id, date)`
- `leave_requests(start_date, end_date)`
- `payroll_periods(year, month)`
- `activity_logs(created_at)`

### 6.3 Composite Indexes
- `attendance_records(employee_id, date_trunc('month', date), status)` — Dashboard absensi
- `payroll_items(employee_id, payroll_period_id)` — Slip gaji per karyawan
- `notifications(user_id, is_read, created_at)` — Notifikasi unread

### 6.4 Special Indexes
- Full-text search index on `employees(full_name)` for search
- Partial indexes for filtered queries (e.g., `WHERE is_active = TRUE`)

---

## 7. Payroll Calculation Logic (PL/pgSQL)

Fungsi `calculate_employee_payroll()` menangani:

```
BASE SALARY + ALLOWANCES + OVERTIME + THR + BONUS
  = GROSS SALARY
  - PPh 21 (via application using TER table)
  - BPJS Kesehatan (rate configurable, default 1% of min(gaji, ceiling Rp12jt))
  - BPJS JHT (rate configurable, default 2% of gaji)
  - BPJS JP (rate configurable, default 1% of min(gaji, ceiling))
  - Loan deduction
  - Other deductions
  = NET SALARY (Take Home Pay)
  
COMPANY COST:
  + BPJS Kesehatan (rate configurable, default 4%)
  + BPJS JHT (rate configurable, default 3.7%)
  + BPJS JP (rate configurable, default 2%)
  + BPJS JKK (risk-based: 0.24-1.74%)
  + BPJS JKM (rate configurable, default 0.3%)
```

---

## 8. Performance Considerations

### 8.1 Materialized View
- `attendance_summary` — Refresh periodik untuk dashboard, bukan query real-time

### 8.2 Partitioning (Future)
- `activity_logs` bisa dipartisi per bulan untuk data > 1 tahun
- `attendance_records` bisa dipartisi per tahun

### 8.3 Connection Pooling
- Gunakan PgBouncer untuk connection pooling
- Minimum 20 connections, maximum 100

### 8.4 Estimated Database Size
- 500 employees → ~500MB (termasuk dokumen & activity logs)
- 2000 employees → ~2GB
- Growth: ~50MB/bulan untuk activity logs

---

## 9. Cara Menjalankan Migration

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Jalankan migration
goose -dir database/migrations postgres "postgres://user:pass@localhost:5432/hrms?sslmode=disable" up

# Rollback 1 migration
goose -dir database/migrations postgres "postgres://user:pass@localhost:5432/hrms?sslmode=disable" down

# Check status
goose -dir database/migrations postgres "postgres://user:pass@localhost:5432/hrms?sslmode=disable" status
```

---

## 10. File Structure

```
database/
├── migrations/
│   ├── 00001_extension_and_encryption.sql    # pgcrypto, ENUMs, encrypt helpers
│   ├── 00002_companies.sql                    # Company settings
│   ├── 00003_organization.sql                 # Departments, positions, grades
│   ├── 00004_schedules_and_locations.sql      # Work schedules, GPS locations
│   ├── 00005_employees.sql                    # Employees (with encrypted data)
│   ├── 00006_attendance_records.sql           # Check-in/out, GPS, face detection
│   ├── 00007_leave_management.sql             # Leave types, requests, balances
│   ├── 00008_reimbursements_overtime.sql      # Reimbursements, overtime
│   ├── 00009_loans.sql                        # Loans & installments
│   ├── 00010_kpi_performance.sql              # KPI templates, reviews
│   ├── 00011_reprimands_announcements_holidays.sql  # SP, announcements, holidays
│   ├── 00012_employee_documents.sql           # Document management
│   ├── 00013_payroll.sql                      # Payroll periods & items
│   ├── 00014_activity_logs_notifications_users.sql  # Audit trail, notifications
│   ├── 00015_final_indexes_and_triggers.sql   # Indexes, functions, views
│   ├── 00016_employee_auth_and_department_schedules.sql  # Roles, auth, department schedules
│   ├── 00017_employee_salary_components.sql              # Salary component master data
│   ├── 00018_flexible_overtime_rates.sql                 # 3-level overtime rate config
│   ├── 00019_flexible_bpjs_config.sql                    # Flexible BPJS rate config
│   └── 00020_shift_change_requests.sql                  # Shift change requests with approval
├── seeds/
│   └── (optional additional seed data)
└── schema-overview.md                        # This document
```

---

## 11. Notes for sqlc Integration

sqlc akan membaca file SQL migration untuk generate Go code:

1. Buat `sqlc.yaml` di root project
2. Tentukan direktori query (e.g., `internal/repository/queries/`)
3. Tiap modul memiliki file query sendiri:
   - `employee.sql` — CRUD employees
   - `attendance.sql` — Absensi queries
   - `payroll.sql` — Payroll queries
   - dll.
4. sqlc akan generate type-safe Go functions dari query SQL

> **PENTING:** Fungsi `encrypt_sensitive` dan `decrypt_sensitive` harus dipanggil di level query SQL, bukan di aplikasi. Atau, alternatifnya: enkripsi di aplikasi Go dengan AES-256-GCM sebelum data dikirim ke database.
