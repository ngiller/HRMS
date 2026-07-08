# HRMS Application — Project Status

> **File ini menyimpan riwayat progress project.**  
> Dibaca oleh AI agent untuk melanjutkan pekerjaan dari sesi sebelumnya.  
> Update file ini setiap kali ada progress signifikan.

**Last Updated:** 8 Juli 2026 (PWA Mobile ✅, Security Hardening ✅, Manual Attendance ✅, Resign ✅)
**Updated By:** AI Agent (Buffy)  
**Project:** HRMS (Human Resource Management System)  
**Stack:** Go (Fiber) + PostgreSQL 16 + SvelteKit (SPA)

---

## 📋 Ringkasan Progress Keseluruhan

| Fase | Status | % |
|------|--------|---|
| **PRD** | ✅ Selesai | 100% |
| **Database Schema** | ✅ Selesai | 100% |
| **Schema Overview** | ✅ Selesai | 100% |
| **Technical Spec** | ✅ Selesai | 100% |
| **Security Hardening** | ✅ Selesai | 100% |
| **UI Mockups** | ✅ Selesai | 100% |
| **Docker Setup (Infrastructure)** | ✅ **Selesai** | 100% |
| **Unit Tests (Go)** | ✅ **Selesai** | 100% |
| **Postman Collection** | ✅ **Selesai** | 100% |
| **TypeScript (0 errors)** | ✅ **Selesai** | 100% |
| **Backend Implementation (Go/Fiber)** | ✅ `go build` + `go vet` + `go test` — **Semua Lulus** | **100%** |
| **Frontend Implementation (SvelteKit)** | ✅ `svelte-check` 0 errors + `npm run build` sukses | **100%** |
| **Chart.js Integration** | ✅ **Selesai** | 100% |
| **Dashboard M3** | ✅ **Selesai** | 100% |
| **AG Grid (all tables)** | ✅ **Selesai** | 100% |
| **Dark Mode** | ✅ **Selesai** | 100% |
| **Dashboard Manager/HR API** | ✅ **Selesai** | 100% |
| **Shift Change Request** | ✅ **Selesai** | 100% |
| **Permission Fix (shift_change)** | ✅ **Selesai** | 100% |
| **Salary Components CRUD** | ✅ **Selesai** | 100% |
| **Payroll Engine — Backend** | ✅ **Selesai** | 100% |
| **Payroll Engine — Frontend** | ✅ **Selesai (Payslip redesign — two-column, full width, detailed breakdown)** | 100% |
| **BPJS Config (Flexible)** | ✅ **DB siap** | 100% |
| **Overtime Rates (Flexible)** | ✅ **DB siap** | 100% |
| **PPh 21 TER** | ✅ **Backend siap** | 100% |
| **THR** | ✅ **Backend siap** | 100% |
| **Payslip** | ✅ **Selesai (Full width, two-column, PDF download, page navigation — no modal)** | 100% |
| **Attendance Record (Check-in/out)** | ✅ **Backend + Frontend** | 100% |
| **Schedule Templates** | ✅ **Backend + Frontend** | 100% |
| **Employee Schedule (Level 2 & 3)** | ✅ **Backend + Frontend** | 100% |
| **Salary Consolidation (Detail Karyawan)** | ✅ **Semua gaji di satu tempat** | 100% |
| **Gaji Privacy (Permission-based)** | ✅ **Kolom gaji hanya untuk user berhak** | 100% |
| **Department RBAC Hardening** | ✅ **Edit/Create/Delete diproteksi permission** | 100% |
| **Overtime (Lembur) — Full Module** | ✅ **Backend CRUD + approval + calculation + Frontend** | 100% |
| **Reimbursement — Full Module** | ✅ **Backend CRUD + approval + pay + Frontend + Receipt Upload** | 100% |
| **Leave Management (Cuti)** | ✅ **Backend CRUD + approval/cancel + balance auto-update + Frontend AG Grid** | **100%** |
| **Dokumen Karyawan (M7)** | ✅ **Backend + Frontend** — 5 endpoints, filter, verify/reject | **100%** |
| **Pengumuman (M7)** | ✅ **Backend + Frontend** — 6 endpoints, pinned, read tracking | **100%** |
| **Hari Libur (M7)** | ✅ **Backend + Frontend** — 6 endpoints, filter tahun, calendar UI | **100%** |
| **AG Grid Bug Fix (18 halaman)** | ✅ **gridApi?.destroy() di semua load() — fix tabel kosong setelah search/delete** | **100%** |
| **Gaji Privacy — Backend** | ✅ **Handler employee: sembunyikan base_salary untuk role employee** | **100%** |
| **Gaji Privacy — Frontend** | ✅ **permissions.ts fallback deny, karyawan column hide, EmployeeDetail wrapped** | **100%** |
| **Salary Component RBAC** | ✅ **Routes.go pindah module employee → payroll** | **100%** |
| **EmployeeDetail UI Redesign** | ✅ **Gradient header, KPI cards, floating form, striped table, empty/loading states** | **100%** |
| **Workflow Analysis & Missing Features** | ✅ **Dokumentasi lengkap + Mermaid diagrams** | **100%** |
| **Migration 00037 — Salary Range (position_grades)** | ✅ **min_salary/max_salary columns + seed data by level** | **100%** |
| **Salary Range Display di Frontend** | ✅ **Info card saat pilih posisi + column di golongan-jabatan** | **100%** |
| **Attendance Export Excel** | ✅ **API endpoint + Tombol Export di halaman absensi** | **100%** |
| **Playwright E2E Tests** | ✅ **6 tests (login + navigation + console error check)** | **100%** |
| **ConfirmModal a11y Fix** | ✅ **Escape key handler + svelte-ignore (0 warnings)** | **100%** |
| **Seed Data Enhancement** | ✅ **Shift change FK fix + position grades fix** | **100%** |
| **Frontend UI/UX Polish** | ✅ **Form pengumuman full-width, detail view redesign, no-modal policy** | **100%** |
| **AG Grid State & Memory Fix** | ✅ **Fix empty tables after search, proper grid destroy/re-init di semua module** | **100%** |
| **Backend Database Integrity** | ✅ **Implementasi Transaction (tx.Begin) di seluruh mutasi data (auth, payroll, announcement)** | **100%** |
| **Daily Working Journal (M13b)** | ✅ **Migration 00040 + Backend + Frontend — AG Grid, dept/date filter, color-coded status** | **100%** |
| **Reprimand Management (M13)** | ✅ **Backend + Frontend — SP1-3, acknowledge, transaction create** | **100%** |
| **Laporan & Analytics (M14)** | ✅ **Backend 6 reports + Frontend 5 tabs** | **100%** |
| **Notifikasi & Audit Trail (M15)** | ✅ **Backend CRUD + Frontend 2 halaman + Sidebar** | **100%** |
| **Frontend Implementation (SvelteKit)** | ✅ `svelte-check` 0 errors + `npm run build` sukses | **100%** |
| **GitHub Actions CI Workflow** | ✅ **Go build/vet/test + svelte-check + integration test** | **100%** |
| **Optimistic Locking Integration Tests** | ✅ **6 test scenarios — double-approve prevention** | **100%** |
| **Transaction Safety (All Write Ops)** | ✅ **WithUserContext di 18+ repository files** | **100%** |
| **Makefile** | ✅ **build, vet, test, migrate, backup, db, docker commands** | **100%** |
| **Database Backup Script** | ✅ **Dump custom + SQL + S3 upload + retention policy** | **100%** |
| **SMTP Email Config** | ✅ **SMTP_HOST/PORT/USER/PASSWORD/FROM/FROM_NAME di .env & config.go** | **100%** |
| **JWT Secret Hardening** | ✅ **Random fallback secret + warning log jika tidak di-set** | **100%** |
| **Approval Workflow Integration** | ✅ **Leave, Loan, Overtime, Reimbursement, Shift Change via ApprovalWorkflowService** | **100%** |
| **THR Calculation Endpoint** | ✅ **GET /api/payroll/periods/:id/calculate-thr** | **100%** |
| **Loan Cancel Endpoint** | ✅ **PUT /api/loans/:id/cancel** | **100%** |
| **Manual Attendance Request** | ✅ **Backend CRUD + approval workflow + Frontend + API Client** | **100%** |
| **Resign & Exit Management** | ✅ **Backend CRUD + approval workflow + exit clearance + Frontend** | **100%** |
| **PWA Service Worker** | ✅ **5 caching strategies, offline fallback, install prompt, navigation cache** | **100%** |
| **PWA Bottom Tab Navigation** | ✅ **BottomTabBar dengan 5 tab + slide-out menu drawer (mobile)** | **100%** |
| **PWA Swipe Actions** | ✅ **SwipeActions.svelte — swipe approve/reject** | **100%** |
| **PWA Manifest & Icons** | ✅ **manifest.json, SVG icons, shortcuts, install prompt** | **100%** |
| **Security Headers (CSP)** | ✅ **X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, CSP, HSTS** | **100%** |
| **File Upload Validation** | ✅ **MIME type + extension + size validation** | **100%** |
| **Company Settings UI** | ✅ **Form BPJS + Profil Perusahaan + Workflow Persetujuan** | **100%** |

---

## 📄 File yang Dibuat

### Desain
| File | Deskripsi | Status |
|------|-----------|--------|
| `PRD-HRMS-Application.md` | PRD — 14 sections | ✅ Final |
| `TECHNICAL-SPEC.md` | Technical Spec — 20 sections | ✅ Final |
| `database/schema-overview.md` | ER diagram, index strategy, payroll logic | ✅ Final |
| `database/migrations/00001-00041.sql` | 41 migration files | ✅ Final |
| `ui-mockups/index.html` | 11 UI wireframes | ✅ Final |
| `postman_collection.json` | Postman Collection — 60+ API endpoints, request body, test scripts | ✅ **Baru** |
| `PROJECT-STATUS.md` | Progress tracker | ✅ Active |
| `ROADMAP.md` | Roadmap & timeline | ✅ Active |
| `MILESTONE.md` | Milestone checklist | ✅ Active |
| `Makefile` | Development task runner — build, test, migrate, backup, docker | ✅ **Baru** |
| `.github/workflows/ci.yml` | GitHub Actions CI — Go + SvelteKit + PostgreSQL 16 | ✅ **Baru** |
| `scripts/backup_db.sh` | Database backup — custom + SQL + S3 + retention | ✅ **Baru** |

### Backend (Go/Fiber)
| File | Deskripsi | Status |
|------|-----------|--------|
| `backend/.env` | Database & JWT config | ✅ Done |
| `backend/main.go` | Entry point: Fiber server, routes, middleware + pass encryptionKey | ✅ Done |
| `backend/internal/config/config.go` | Environment config loader + EncryptionKey field | ✅ Done |
| `backend/internal/database/database.go` | pgx connection pool + AfterConnect set app.encryption_key | ✅ Done |
| `backend/internal/models/auth.go` | Auth models (login, forgot/reset password) | ✅ Done |
| `backend/internal/models/employee.go` | Employee models (detail, summary, dashboard, history) | ✅ Done |
| `backend/internal/repository/auth_repo.go` | Auth queries (login, password reset, login attempts, refresh token) | ✅ Done |
| `backend/internal/repository/employee_repo.go` | Employee queries (list, detail, dashboard stats, history, CRUD) | ✅ Done |
| `backend/internal/service/auth_service.go` | Auth logic (bcrypt, JWT, forgot/reset, refresh, logout) | ✅ Done |
| `backend/internal/service/employee_service.go` | Employee logic (list, detail, dashboard, create, update, delete, history, export Excel, import Excel, upload photo) | ✅ Done |
| `backend/internal/handlers/auth_handler.go` | Auth HTTP handlers (login, forgot/reset, me, refresh, logout) | ✅ Done |
| `backend/internal/handlers/employee_handler.go` | Employee HTTP handlers (list, get, dashboard, create, update, delete, history, export, import, photo upload, updateWorkSchedule) | ✅ Done |
| `backend/internal/middleware/auth.go` | JWT Bearer middleware + CORS | ✅ Done |
| `backend/internal/middleware/rbac.go` | RBAC middleware — role-based permission check per module+action | ✅ **Done** |
| `backend/internal/middleware/ratelimit.go` | Rate limiter — tiered in-memory (critical 5/min, high 10/min, low 100/min) | ✅ **Done** |
| `backend/internal/repository/rbac_repo.go` | RBAC repository — query permissions dari tabel roles | ✅ **Done** |
| `backend/internal/routes/routes.go` | Route definitions (dipisah dari main.go) — 40+ routes | ✅ **Baru** |
| `backend/internal/models/department.go` | Department models (struct, summary, request/response) | ✅ **Baru** |
| `backend/internal/repository/department_repo.go` | Department CRUD queries (list, get, create, update, delete) | ✅ **Baru** |
| `backend/internal/service/department_service.go` | Department business logic (validasi, CRUD) | ✅ **Baru** |
| `backend/internal/handlers/department_handler.go` | Department HTTP handlers (7 endpoints — + UpdateWorkSchedule) | ✅ **Baru** |
| `backend/internal/models/role.go` | Role models + permission template (17 modul) | ✅ **Baru** |
| `backend/internal/repository/role_repo.go` | Role CRUD queries + JSONB permission handling | ✅ **Baru** |
| `backend/internal/service/role_service.go` | Role logic (validasi, system role protection) | ✅ **Baru** |
| `backend/internal/handlers/role_handler.go` | Role HTTP handlers (6 endpoints) | ✅ **Baru** |
| `backend/cmd/seed/main.go` | Seed script (admin + 8 test employees) + pass encryptionKey | ✅ Done |
| `backend/.air.toml` | Air hot reload config — auto rebuild saat file .go berubah | ✅ Done |
| `backend/internal/config/config.go` | URL-encode password di DatabaseURL() — support special chars (/) | ✅ **Fix** |
| `.env` | Environment variables + DB_PASSWORD=/ | ✅ **Updated** |
| `docker-compose.yml` | 4 services (db, migrate, api, web) + password encoding di migrate command | ✅ **Updated** |
| `backend/internal/models/position_grade.go` | Position Grade model | ✅ **Baru** |
| `backend/internal/repository/position_grade_repo.go` | Position Grade CRUD queries | ✅ **Baru** |
| `backend/internal/service/position_grade_service.go` | Position Grade business logic | ✅ **Baru** |
| `backend/internal/handlers/position_grade_handler.go` | Position Grade HTTP handlers (6 endpoints) | ✅ **Baru** |
| `backend/internal/models/position.go` | Position model | ✅ **Baru** |
| `backend/internal/repository/position_repo.go` | Position CRUD queries | ✅ **Baru** |
| `backend/internal/service/position_service.go` | Position business logic | ✅ **Baru** |
| `backend/internal/handlers/position_handler.go` | Position HTTP handlers (6 endpoints) | ✅ **Baru** |
| `backend/internal/models/work_schedule.go` | Work Schedule model | ✅ **Baru** |
| `backend/internal/repository/work_schedule_repo.go` | Work Schedule CRUD queries | ✅ **Baru** |
| `backend/internal/service/work_schedule_service.go` | Work Schedule business logic | ✅ **Baru** |
| `backend/internal/handlers/work_schedule_handler.go` | Work Schedule HTTP handlers (6 endpoints) | ✅ **Baru** |
| `backend/internal/models/attendance_location.go` | Attendance Location model | ✅ **Baru** |
| `backend/internal/repository/attendance_location_repo.go` | Attendance Location CRUD queries | ✅ **Baru** |
| `backend/internal/service/attendance_location_service.go` | Attendance Location business logic | ✅ **Baru** |
| `backend/internal/handlers/attendance_location_handler.go` | Attendance Location HTTP handlers (6 endpoints) | ✅ **Baru** |
| `backend/internal/models/organization.go` | Organization tree node model | ✅ **Baru** |
| `backend/internal/repository/organization_repo.go` | Organization tree queries (depts, positions, employees) | ✅ **Baru** |
| `backend/internal/service/organization_service.go` | Organization tree assembly (dept → position → employee) | ✅ **Baru** |
| `backend/internal/handlers/organization_handler.go` | Organization tree HTTP handler | ✅ **Baru** |
| `backend/internal/handlers/response.go` | Response helpers: SuccessResponse, ErrorResponse, SuccessResponseWithMeta, PaginationMeta | ✅ **Baru** |
| `backend/internal/models/shift_change_request.go` | Shift Change Request model — struct, summary, create/update/status requests, list response | ✅ **Baru** |
| `backend/internal/repository/shift_change_repo.go` | Shift Change CRUD queries — list (paginated + filter status/employee), get, create (individual/swap), update status, confirm swap, cancel, duplicate check | ✅ **Baru** |
| `backend/internal/service/shift_change_service.go` | Shift Change logic — validasi tipe/tanggal/jadwal/alasan, duplicate pending check, status transition (create → pending/partner_pending → approve/reject/cancel/confirm-swap) | ✅ **Baru** |
| `backend/internal/handlers/shift_change_handler.go` | Shift Change HTTP handlers — 7 endpoints: list, get, create, approve, reject, confirm-swap, cancel | ✅ **Baru** |
| `backend/internal/models/salary_component.go` | Salary Component model — struct, summary, create/update/delete request/response | ✅ **Baru** |
| `backend/internal/repository/salary_component_repo.go` | Salary Component queries — CRUD + WithUserContext (set app.current_user_id untuk trigger) | ✅ **Baru** |
| `backend/internal/service/salary_component_service.go` | Salary Component logic — validasi (component_name, type, amount >= 0), CRUD | ✅ **Baru** |
| `backend/internal/handlers/salary_component_handler.go` | Salary Component HTTP handlers — 4 endpoints (list, create, update, delete) nested di /employees/:id/salary-components | ✅ **Baru** |
| `backend/internal/models/payroll.go` | Payroll models — PayrollPeriod, PayrollItem, CreatePeriodRequest, CalculateRequest, PayslipResponse, PPh21Calculation | ✅ **Baru** |
| `backend/internal/repository/payroll_repo.go` | Payroll CRUD queries — periods, items, panggil `calculate_employee_payroll()`, get active salary components, get base salary, PPh 21 TER, THR | ✅ **Baru** |
| `backend/internal/service/payroll_service.go` | Payroll logic — period management (create/approve/pay), calculate payroll, PPh 21 TER (Go), THR, payslip generation | ✅ **Baru** |
| `backend/internal/handlers/payroll_handler.go` | Payroll HTTP handlers — 8 endpoints: list/create/get period, calculate, list items, approve, pay, get payslip | ✅ **Baru** |
| `backend/internal/models/attendance_record.go` | Attendance Record model — check-in/out, today status, history, report | ✅ **Baru** |
| `backend/internal/repository/attendance_record_repo.go` | Attendance Record queries — check-in, check-out, today status, my history, report | ✅ **Baru** |
| `backend/internal/service/attendance_record_service.go` | Attendance Record logic — validasi lokasi GPS, duplicate check-in, work schedule resolution | ✅ **Baru** |
| `backend/internal/handlers/attendance_record_handler.go` | Attendance Record HTTP handlers — 5 endpoints: today status, check-in, check-out, my-history, report | ✅ **Baru** |
| `backend/internal/models/schedule.go` | Schedule models — ScheduleTemplate (level 1), EmployeeSchedule (level 2 & 3) | ✅ **Baru** |
| `backend/internal/repository/schedule_repo.go` | Schedule queries — CRUD templates, CRUD employee schedules, resolve schedule | ✅ **Baru** |
| `backend/internal/service/schedule_service.go` | Schedule logic — validasi, resolve hierarchy (template → dept → employee) | ✅ **Baru** |
| `backend/internal/handlers/schedule_handler.go` | Schedule HTTP handlers — 11 endpoints: templates CRUD, employee schedules CRUD, resolve | ✅ **Baru** |
| `backend/internal/models/overtime_request.go` | Overtime Request model — struct, summary, create/status requests, calculation | ✅ **Baru** |
| `backend/internal/repository/overtime_repo.go` | Overtime CRUD queries — list (paginated + filter), get, create, approve/reject/cancel, get calculation | ✅ **Baru** |
| `backend/internal/service/overtime_service.go` | Overtime logic — validasi (tanggal, jam, tipe, alasan), status flow (create → pending → approve/reject/cancel), get calculation | ✅ **Baru** |
| `backend/internal/handlers/overtime_handler.go` | Overtime HTTP handlers — 7 endpoints: list, get, create, approve, reject, cancel, get calculation | ✅ **Baru** |
| `backend/internal/models/reimbursement.go` | Reimbursement model — struct, summary, create/status/pay requests | ✅ **Baru** |
| `backend/internal/repository/reimbursement_repo.go` | Reimbursement CRUD queries — list (paginated + filter), get, create, approve/reject/pay/cancel | ✅ **Baru** |
| `backend/internal/service/reimbursement_service.go` | Reimbursement logic — validasi (tipe, jumlah, deskripsi), status flow (create → pending → approve/reject/pay/cancel), upload receipt | ✅ **Baru** |
| `backend/internal/handlers/reimbursement_handler.go` | Reimbursement HTTP handlers — 7 endpoints: list, get, create, approve, reject, pay, cancel + upload receipt | ✅ **Baru** |
| `backend/internal/models/leave.go` | Leave models — LeaveType, LeaveBalance, LeaveRequest + request/response structs | ✅ **Baru** |
| `backend/internal/repository/leave_repo.go` | Leave CRUD queries — list, get, create, approve/reject/cancel, types, balances, auto-update balance on approve | ✅ **Baru** |
| `backend/internal/service/leave_service.go` | Leave logic — validasi, status flow (pending → approved/rejected/cancelled) | ✅ **Baru** |
| `backend/internal/handlers/leave_handler.go` | Leave HTTP handlers — 7 endpoints: list, get, create, approve, reject, cancel, types, my-balances, all-balances | ✅ **Baru** |

### Frontend (SvelteKit)
| File | Deskripsi | Status |
|------|-----------|--------|
| `frontend/src/lib/config.js` | API configuration (backend URL) | ✅ Done |
| `frontend/src/lib/api.js` | API client (fetch + JWT auth + auto-refresh + error handling + salary components + shift change) | ✅ Done |
| `frontend/src/lib/permissions.ts` | Frontend RBAC helper — `hasPermission()`, `getAccessibleMenus()`, `SIDEBAR_MENUS` | ✅ **Baru** |
| `frontend/src/routes/login/+page.svelte` | Login page — terhubung ke backend API | ✅ Done |
| `frontend/src/routes/forgot-password/+page.svelte` | Forgot password — input email + success state | ✅ Done |
| `frontend/src/routes/reset-password/+page.svelte` | Reset password — token from URL, confirm password | ✅ Done |
| `frontend/src/routes/(app)/+layout.svelte` | App layout — sidebar, topbar, user dropdown | ✅ **Route group** |
| `frontend/src/routes/(app)/dashboard/+page.svelte` | Dashboard — stat cards, charts, approvals | ✅ **Route: /dashboard** |
| `frontend/src/routes/(app)/[...rest]/+page.svelte` | Catch-all — sidebar menu placeholder | ✅ **Route group** |
| `frontend/src/routes/(app)/karyawan/+page.svelte` | Daftar Karyawan — AG Grid, search, pagination, inline create/edit form, salary components fields (NIK, NPWP, Bank), export/import | ✅ **Route: /karyawan** |
| `frontend/src/routes/(app)/karyawan/[id]/+page.svelte` | Detail Karyawan — info pribadi, pekerjaan, alamat, history, **salary components CRUD + summary cards**, photo upload, schedule override | ✅ **Route: /karyawan/[id]** |
| `frontend/src/routes/(app)/departemen/+page.svelte` | Daftar Departemen — AG Grid, search, stats bar, inline CRUD, toggle is_active | ✅ **Route: /departemen** |
| `frontend/src/routes/(app)/pengaturan/roles/+page.svelte` | Role Management — AG Grid, stats bar, inline form dengan permission editor | ✅ **Route: /pengaturan/roles** |
| `frontend/src/routes/(app)/golongan-jabatan/+page.svelte` | Halaman Golongan Jabatan — AG Grid, CRUD, search, pagination | ✅ **Baru** |
| `frontend/src/routes/(app)/posisi-jabatan/+page.svelte` | Halaman Posisi Jabatan — AG Grid, CRUD, dropdown departemen & golongan | ✅ **Baru** |
| `frontend/src/routes/(app)/jadwal-kerja/+page.svelte` | Halaman Jadwal Kerja — AG Grid, CRUD, input time per hari, detail view | ✅ **Baru** |
| `frontend/src/routes/(app)/lokasi-absensi/+page.svelte` | Halaman Lokasi Absensi — AG Grid, CRUD, koordinat GPS & radius | ✅ **Baru** |
| `frontend/src/routes/(app)/struktur-organisasi/+page.svelte` | Halaman Struktur Organisasi — collapsible tree view with search | ✅ **Baru** |
| `frontend/src/routes/(app)/permintaan-shift/+page.svelte` | Halaman Permintaan Shift — AG Grid + filter status, form create (individual/swap), detail view, approve/reject/cancel/confirm-swap | ✅ **Baru** |
| `frontend/src/routes/(app)/absensi/+page.svelte` | Halaman Absensi — check-in/out, today status, riwayat absensi, report | ✅ **Baru** |
| `frontend/src/routes/(app)/jadwal-templates/+page.svelte` | Halaman Template Jadwal — CRUD template jadwal (level 1) | ✅ **Baru** |
| `frontend/src/routes/(app)/jadwal-karyawan/+page.svelte` | Halaman Jadwal Karyawan — employee schedule (level 2 & 3), resolve hierarki | ✅ **Baru** |
| `frontend/src/routes/(app)/penggajian/+page.svelte` | Halaman Periode Penggajian — daftar periode, status badges, create/calculate/approve/pay | ✅ **Baru** |
| `frontend/src/routes/(app)/penggajian/[id]/+page.svelte` | Halaman Detail Payroll — tabel payroll items per karyawan, totals, filter | ✅ **Baru** |
| `frontend/src/routes/(app)/penggajian/payslip/[periodId]/[employeeId]/+page.svelte` | Halaman Slip Gaji — detail payslip, income/deduction breakdown, PDF download | ✅ **Baru** |
| `frontend/src/routes/(app)/penggajian/slip-saya/+page.svelte` | Halaman Slip Gaji Saya — daftar payslip self-service, navigasi ke detail payslip | ✅ **Baru** |
| `frontend/src/lib/components/EmployeeDetail.svelte` | Komponen Detail Karyawan — info pribadi, pekerjaan, alamat, history, salary components CRUD, photo upload, schedule override | ✅ **Baru** |
| `frontend/src/routes/(app)/lembur/+page.svelte` | Halaman Lembur — AG Grid, form ajukan (tanggal, jam, tipe, alasan), detail view, approve/reject/cancel | ✅ **Baru** |
| `frontend/src/routes/(app)/reimbursement/+page.svelte` | Halaman Reimbursement — AG Grid, form ajukan (tipe, jumlah, deskripsi, foto receipt), detail view, approve/reject/pay/cancel, receipt image gallery | ✅ **Baru** |
| `backend/internal/models/document.go` | Document models — struct, summary, create request | ✅ **Baru** |
| `backend/internal/repository/document_repo.go` | Document CRUD queries — list, get, create, verify/reject, delete | ✅ **Baru** |
| `backend/internal/service/document_service.go` | Document logic — validasi, verify, reject | ✅ **Baru** |
| `backend/internal/handlers/document_handler.go` | Document HTTP handlers — 5 endpoints | ✅ **Baru** |
| `backend/internal/models/announcement.go` | Announcement models — struct, summary, create/update request | ✅ **Baru** |
| `backend/internal/repository/announcement_repo.go` | Announcement CRUD queries — list, get, create, update, delete, mark read | ✅ **Baru** |
| `backend/internal/service/announcement_service.go` | Announcement logic — validasi, CRUD, mark read | ✅ **Baru** |
| `backend/internal/handlers/announcement_handler.go` | Announcement HTTP handlers — 6 endpoints | ✅ **Baru** |
| `backend/internal/models/holiday.go` | Holiday models — struct, create/update request, year response | ✅ **Baru** |
| `backend/internal/repository/holiday_repo.go` | Holiday CRUD queries — list (filter year/type), get, create, update, delete, get by year | ✅ **Baru** |
| `backend/internal/service/holiday_service.go` | Holiday logic — validasi, CRUD, get by year | ✅ **Baru** |
| `backend/internal/handlers/holiday_handler.go` | Holiday HTTP handlers — 6 endpoints | ✅ **Baru** |
| `frontend/src/routes/(app)/dokumen/+page.svelte` | Halaman Dokumen Karyawan — AG Grid, form tambah, detail view, verify/reject | ✅ **Baru** |
| `frontend/src/routes/(app)/pengumuman/+page.svelte` | Halaman Pengumuman — AG Grid, filter tipe, create/edit form, detail + read tracking | ✅ **Baru** |
| `frontend/src/routes/(app)/hari-libur/+page.svelte` | Halaman Hari Libur — AG Grid, calendar-like, filter tahun, grouped by month, create/edit | ✅ **Baru** |

---

## 🔌 API Endpoints — Sudah Berfungsi (60+ endpoints)

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| `GET` | `/api/health` | ❌ | Health check |
| `POST` | `/api/auth/login` | ❌ | Login (return JWT + user data) |
| `POST` | `/api/auth/forgot-password` | ❌ | Kirim reset token |
| `POST` | `/api/auth/reset-password` | ❌ | Reset password dengan token |
| `POST` | `/api/auth/refresh` | ❌ | Refresh token (token rotation) |
| `POST` | `/api/auth/logout` | ✅ | Logout (invalidasi session) |
| `GET` | `/api/auth/me` | ✅ | Data user yang login |
| `GET` | `/api/employees` | ✅ | Daftar karyawan (paginated, filter dept + status) |
| `GET` | `/api/employees/export` | ✅ | Export seluruh karyawan ke Excel (.xlsx) |
| `POST` | `/api/employees/import` | ✅ | Import karyawan dari file Excel |
| `POST` | `/api/employees` | ✅ | Tambah karyawan (bcrypt + validasi) |
| `GET` | `/api/employees/:id` | ✅ | Detail karyawan (include photo_url) |
| `POST` | `/api/employees/:id/photo` | ✅ | Upload foto karyawan (JPEG/PNG, max 2MB) |
| `GET` | `/api/employees/:id/salary-components` | ✅ | Daftar komponen gaji per karyawan (paginated) |
| `POST` | `/api/employees/:id/salary-components` | ✅ | Tambah komponen gaji (allowance/deduction) |
| `PUT` | `/api/employees/:id/salary-components/:componentId` | ✅ | Update komponen gaji |
| `DELETE` | `/api/employees/:id/salary-components/:componentId` | ✅ | Hapus komponen gaji (soft delete) |
| `PUT` | `/api/employees/:id` | ✅ | Update karyawan (dynamic SET) |
| `PUT` | `/api/employees/:id/restore` | ✅ | Aktifkan kembali karyawan (restore soft delete) |
| `GET` | `/api/employees/:id/history` | ✅ | Riwayat perubahan karyawan |
| `GET` | `/api/employees` | ✅ | **+ param `include_deleted=true`** — tampilkan data nonaktif |
| `DELETE` | `/api/employees/:id` | ✅ | Nonaktifkan karyawan (soft delete) |
| `GET` | `/api/dashboard` | ✅ | Statistik dashboard |
| `GET` | `/api/payroll/periods` | ✅ | Daftar periode penggajian (paginated) |
| `POST` | `/api/payroll/periods` | ✅ | Buat periode penggajian baru |
| `GET` | `/api/payroll/periods/:id` | ✅ | Detail periode penggajian |
| `POST` | `/api/payroll/periods/:id/calculate` | ✅ | Hitung payroll untuk periode tertentu |
| `GET` | `/api/payroll/periods/:id/items` | ✅ | Daftar payroll items per periode |
| `PUT` | `/api/payroll/periods/:id/approve` | ✅ | Setujui periode penggajian |
| `PUT` | `/api/payroll/periods/:id/pay` | ✅ | Bayar periode penggajian |
| `GET` | `/api/payroll/payslips/:periodId/:employeeId` | ✅ | Slip gaji per karyawan |
| `GET` | `/api/payroll/my-payslips` | ✅ | Daftar slip gaji saya |
| `GET` | `/api/payroll/my-payslips/:periodId` | ✅ | Slip gaji saya per periode |
| `GET` | `/api/organization/tree` | ✅ | Struktur organisasi tree |
| `GET` | `/api/departments` | ✅ | Daftar departemen (paginated) |
| `GET` | `/api/departments/all` | ✅ | Semua departemen (for dropdown) |
| `GET` | `/api/departments/:id` | ✅ | Detail departemen |
| `POST` | `/api/departments` | ✅ | Tambah departemen |
| `PUT` | `/api/departments/:id` | ✅ | Update departemen |
| `PUT` | `/api/departments/:id/work-schedule` | ✅ | Assign jadwal kerja ke departemen |
| `PUT` | `/api/employees/:id/work-schedule` | ✅ | Override jadwal kerja individu (set atau reset ke NULL untuk ikuti departemen) |
| `DELETE` | `/api/departments/:id` | ✅ | Hapus departemen (soft delete) |
| `GET` | `/api/roles` | ✅ | Daftar role (paginated) |
| `GET` | `/api/roles/:id` | ✅ | Detail role + permissions |
| `GET` | `/api/roles/permissions/template` | ✅ | Template permission (17 modul) |
| `POST` | `/api/roles` | ✅ | Tambah role |
| `PUT` | `/api/roles/:id` | ✅ | Update role |
| `DELETE` | `/api/roles/:id` | ✅ | Hapus role (kustom saja) |
| `GET` | `/api/position-grades` | ✅ | Daftar golongan jabatan (paginated) |
| `GET` | `/api/position-grades/all` | ✅ | Semua golongan jabatan (for dropdown) |
| `GET` | `/api/position-grades/:id` | ✅ | Detail golongan jabatan |
| `POST` | `/api/position-grades` | ✅ | Tambah golongan jabatan |
| `PUT` | `/api/position-grades/:id` | ✅ | Update golongan jabatan |
| `DELETE` | `/api/position-grades/:id` | ✅ | Hapus golongan jabatan |
| `GET` | `/api/positions` | ✅ | Daftar posisi jabatan (paginated) |
| `GET` | `/api/positions/all` | ✅ | Semua posisi jabatan (for dropdown) |
| `GET` | `/api/positions/:id` | ✅ | Detail posisi jabatan |
| `POST` | `/api/positions` | ✅ | Tambah posisi jabatan |
| `PUT` | `/api/positions/:id` | ✅ | Update posisi jabatan |
| `DELETE` | `/api/positions/:id` | ✅ | Hapus posisi jabatan |
| `GET` | `/api/work-schedules` | ✅ | Daftar jadwal kerja (paginated) |
| `GET` | `/api/work-schedules/all` | ✅ | Semua jadwal kerja (for dropdown) |
| `GET` | `/api/work-schedules/:id` | ✅ | Detail jadwal kerja |
| `POST` | `/api/work-schedules` | ✅ | Tambah jadwal kerja |
| `PUT` | `/api/work-schedules/:id` | ✅ | Update jadwal kerja |
| `DELETE` | `/api/work-schedules/:id` | ✅ | Hapus jadwal kerja |
| `GET` | `/api/attendance-locations` | ✅ | Daftar lokasi absensi (paginated) |
| `GET` | `/api/attendance-locations/all` | ✅ | Semua lokasi absensi (for dropdown) |
| `GET` | `/api/attendance-locations/:id` | ✅ | Detail lokasi absensi |
| `POST` | `/api/attendance-locations` | ✅ | Tambah lokasi absensi |
| `PUT` | `/api/attendance-locations/:id` | ✅ | Update lokasi absensi |
| `DELETE` | `/api/attendance-locations/:id` | ✅ | Hapus lokasi absensi |
| `GET` | `/api/shift-change-requests` | ✅ | Daftar permintaan shift (paginated, filter status + employee) |
| `GET` | `/api/shift-change-requests/:id` | ✅ | Detail permintaan shift |
| `POST` | `/api/shift-change-requests` | ✅ | Buat permintaan shift baru (individual/swap) |
| `PUT` | `/api/shift-change-requests/:id/approve` | ✅ | Setujui permintaan shift |
| `PUT` | `/api/shift-change-requests/:id/reject` | ✅ | Tolak permintaan shift |
| `PUT` | `/api/shift-change-requests/:id/confirm-swap` | ✅ | Konfirmasi partner swap |
| `PUT` | `/api/shift-change-requests/:id/cancel` | ✅ | Batalkan permintaan shift |
| `GET` | `/api/attendance/today` | ✅ | Status absensi hari ini (check-in/out) |
| `POST` | `/api/attendance/check-in` | ✅ | Check-in absensi (validasi lokasi GPS) |
| `PUT` | `/api/attendance/check-out` | ✅ | Check-out absensi |
| `GET` | `/api/attendance/my-history` | ✅ | Riwayat absensi saya (self-service) |
| `GET` | `/api/attendance/report` | ✅ | Laporan absensi (HR/Manager) |
| `GET` | `/api/schedule-templates` | ✅ | Daftar template jadwal (level 1) |
| `GET` | `/api/schedule-templates/all` | ✅ | Semua template jadwal (for dropdown) |
| `GET` | `/api/schedule-templates/:id` | ✅ | Detail template jadwal |
| `POST` | `/api/schedule-templates` | ✅ | Tambah template jadwal |
| `PUT` | `/api/schedule-templates/:id` | ✅ | Update template jadwal |
| `DELETE` | `/api/schedule-templates/:id` | ✅ | Hapus template jadwal |
| `GET` | `/api/employee-schedules` | ✅ | Daftar jadwal karyawan (level 2 & 3) |
| `GET` | `/api/employee-schedules/resolve` | ✅ | Resolve jadwal berdasarkan hierarki (template → dept → employee) |
| `GET` | `/api/employee-schedules/:id` | ✅ | Detail jadwal karyawan |
| `POST` | `/api/employee-schedules` | ✅ | Tambah jadwal karyawan |
| `PUT` | `/api/employee-schedules/:id` | ✅ | Update jadwal karyawan |
| `DELETE` | `/api/employee-schedules/:id` | ✅ | Hapus jadwal karyawan |
| `GET` | `/api/overtime-requests` | ✅ | Daftar permintaan lembur (paginated, filter status + employee) |
| `GET` | `/api/overtime-requests/:id` | ✅ | Detail permintaan lembur |
| `POST` | `/api/overtime-requests` | ✅ | Buat permintaan lembur baru |
| `GET` | `/api/overtime-requests/:id/calculation` | ✅ | Perhitungan lembur (berdasarkan rate fleksibel) |
| `PUT` | `/api/overtime-requests/:id/approve` | ✅ | Setujui permintaan lembur |
| `PUT` | `/api/overtime-requests/:id/reject` | ✅ | Tolak permintaan lembur |
| `PUT` | `/api/overtime-requests/:id/cancel` | ✅ | Batalkan permintaan lembur |
| `GET` | `/api/reimbursements` | ✅ | Daftar reimbursement (paginated, filter status + type) |
| `GET` | `/api/reimbursements/:id` | ✅ | Detail reimbursement + receipt images |
| `POST` | `/api/reimbursements` | ✅ | Buat reimbursement baru |
| `POST` | `/api/reimbursements/upload` | ✅ | Upload foto bukti reimbursement (JPEG/PNG, max 5MB) |
| `PUT` | `/api/reimbursements/:id/approve` | ✅ | Setujui reimbursement |
| `PUT` | `/api/reimbursements/:id/reject` | ✅ | Tolak reimbursement |
| `PUT` | `/api/reimbursements/:id/pay` | ✅ | Bayar reimbursement |
| `PUT` | `/api/reimbursements/:id/cancel` | ✅ | Batalkan reimbursement |

### API Endpoints Baru (M7 — Dokumen & Pengumuman)

| Method | Endpoint | RBAC | Deskripsi |
|--------|----------|------|-----------|
| `GET` | `/api/documents` | `document:read` | Daftar dokumen (filter status/employee/type) |
| `GET` | `/api/documents/:id` | `document:read` | Detail dokumen |
| `POST` | `/api/documents` | `document:create` | Tambah dokumen |
| `PUT` | `/api/documents/:id/verify` | `document:update` | Verifikasi dokumen |
| `PUT` | `/api/documents/:id/reject` | `document:update` | Tolak dokumen |
| `DELETE` | `/api/documents/:id` | `document:delete` | Hapus dokumen |
| `GET` | `/api/announcements` | `announcement:read` | Daftar pengumuman (filter type) |
| `GET` | `/api/announcements/:id` | `announcement:read` | Detail pengumuman + read count |
| `POST` | `/api/announcements` | `announcement:create` | Buat pengumuman |
| `PUT` | `/api/announcements/:id` | `announcement:update` | Update pengumuman |
| `DELETE` | `/api/announcements/:id` | `announcement:delete` | Hapus pengumuman |
| `POST` | `/api/announcements/:id/read` | — | Tandai sudah dibaca |
| `GET` | `/api/holidays` | `announcement:read` | Daftar hari libur (filter tahun/type) |
| `GET` | `/api/holidays/year/:year` | — | Hari libur per tahun |
| `GET` | `/api/holidays/:id` | `announcement:read` | Detail hari libur |
| `POST` | `/api/holidays` | `announcement:create` | Tambah hari libur |
| `PUT` | `/api/holidays/:id` | `announcement:update` | Update hari libur |
| `DELETE` | `/api/holidays/:id` | `announcement:delete` | Hapus hari libur |

### Akun Test
| Role | Slug | Email | Password | Akses |
|------|------|-------|----------|-------|
| **Super Admin** | `super_admin` | `admin@company.com` | `admin123` | Semua modul (full access) |
| **HR Manager** | `hr_manager` | `dewi@company.com` | `password123` | Kelola karyawan, payroll (CRUD), cuti, reimburs, absensi |
| **Finance** | `finance` | `siti@company.com` | `password123` | Kelola payroll, pinjaman, reimbursement approve |
| **Karyawan** | `employee` | `budi@company.com` | `password123` | Self-service: lihat data sendiri, payslip, absensi |
| **Karyawan** | `employee` | `andi@company.com` | `password123` | Sama (self-service) |
| **Karyawan** | `employee` | `rudi@company.com` | `password123` | Sama |
| **Finance** | `finance` | `hendra@company.com` | `password123` | Sama seperti Siti (Finance) |

> **Test RBAC**: 
> - `admin@company.com` — full access semua menu (Super Admin)
> - `dewi@company.com` — lihat menu **Payroll, Karyawan, Absensi, Cuti**, dll (HR Manager)
> - `siti@company.com` — lihat menu **Payroll & Penggajian**, Pinjaman (Finance)
> - `budi@company.com` — menu terbatas (self-service), **tidak ada menu Payroll**

---

## 🗺️ Arsitektur Terkini

```
Frontend (SvelteKit SPA) — localhost:5173
  └── config.js → API_BASE_URL = http://localhost:8080
        └── POST /api/auth/login ──→ Backend (Go / Fiber) — localhost:8080
              └── Middleware: JWT Auth, CORS, Helmet, Logger, Recover, RequestID
              └── Middleware: RBAC (role-based), Rate Limiter (tiered in-memory)
              └── Service: Auth, Employee, Department, Role, SalaryComponent, etc.
              └── Repository: pgx/v5 parameterized queries
              └── Response Helpers: SuccessResponse/ErrorResponse (standar)
              └── Database: PostgreSQL 16 (hrms)
                    ├── 22 migrations (44+ tables, fungsi PL/pgSQL)
                    ├── 7 roles with permissions
                    ├── 9 employees (1 admin + 8 test)
                    ├── 5 attendance locations seed
                    └── Payroll: calculate_employee_payroll(), payslip_view, flexible BPJS & overtime
```

---

## ▶️ Cara Menjalankan

```bash
# Backend (dengan hot reload)
cd backend
air

# Backend (tanpa hot reload)
cd backend
go run .

# Frontend
cd frontend
npm run dev

# Buka http://localhost:5173
# Login: admin@company.com / admin123
```

---

## ✅ Status Build Terkini (2 Juli 2026 — Migration 00037 + Salary Range + Export Excel ✅)

| Area | Hasil | Detail |
|------|-------|--------|
| **Backend (`go build ./...`)** | ✅ **Bersih** | Kompilasi sukses, 0 error |
| **Backend (`go vet ./...`)** | ✅ **Bersih** | 0 warning/vet issue |
| **Backend (`go test ./...`)** | ✅ **Lulus** | Semua test pass |
| **Frontend (`svelte-check`)** | ✅ **0 errors, 0 warnings** | `svelte-check` — 0 errors, 0 warnings ✅ |
| **Frontend (`npm run build`)** | ✅ **Sukses** | Build production berhasil. |

## ✅ Sesi Ini — 2 Juli 2026 (Migration 00037 + Salary Range + Export Excel + E2E Tests)

### 🐛 AG Grid Bug Fix (18 Halaman)

**Bug:** Setelah search/delete/pagination, tabel AG Grid kosong karena `{#if isLoading}` menghancurkan DOM element grid, tapi `gridApi` masih mereferensi grid yang sudah hancur.

**Fix:** Tambahkan `gridApi?.destroy(); gridApi = null;` di awal setiap fungsi `load()`:

| Sesi | Halaman |
|------|---------|
| **Sesi 1 (6)** | golongan-jabatan, posisi-jabatan, departemen, jadwal-kerja, jadwal-templates, lokasi-absensi |
| **Sesi 2 (7)** | lembur, reimbursement, cuti, permintaan-shift, pengumuman, dokumen, hari-libur |
| **Sesi 3 (5)** | pengaturan/roles, jadwal-karyawan, karyawan, penggajian/[id], penggajian |

### 🔒 Gaji Privacy — Backend & Frontend

| File | Perubahan |
|------|-----------|
| `backend/internal/handlers/employee_handler.go` | `ListEmployees`: set `BaseSalary=0` untuk role employee. `GetEmployee`: set `BaseSalary=nil, DailyWage=nil` jika bukan dirinya |
| `frontend/src/lib/permissions.ts` | Fallback `return true` → `return false` (deny by default) |
| `frontend/.../karyawan/+page.svelte` | Kondisi `hide` kolom gaji: `!hasPermission('payroll','read')` |
| `frontend/.../EmployeeDetail.svelte` | Section "Pengaturan Gaji" wrapped dengan `{#if hasPermission('payroll','read')}` |
| `backend/internal/routes/routes.go` | RBAC salary components pindah module `employee` → `payroll` |

### 🎨 EmployeeDetail.svelte UI Redesign

Section "Pengaturan Gaji" di-redesign profesional:
- Gradient header dengan icon badge
- Salary breakdown card (Gaji Pokok + Upah Harian) dengan input fields
- KPI Summary Cards: Total Tunjangan (emerald), Total Potongan (red), Gaji Bersih (blue)
- Inline form dengan floating label + border-2
- Tabel dengan gradient header, striped rows, hover-reveal actions, status dot
- Mobile cards dengan colored type badges
- Empty state + Loading state lebih baik
- Fix: `w-4.5` → `w-5` (Tailwind tidak punya .5 scale)
- Fix: Ternary `<span>` inside `{expression}` → `{#if}/{:else}` blocks
- Fix: Missing `</p>` closing tag

### 📋 Workflow Analysis & Missing Features

- Analisis lengkap PRD vs Implementasi (13 fitur ❌, 2 🚧 parsial, 20+ ✅)
- 8 workflow diagram Mermaid: Autentikasi, Master Data, Absensi, Cuti, Payroll, Dokumen, Siklus Hidup, Arsitektur, RBAC Tree
- Prioritas rekomendasi fitur yang belum ada

---

### 🔥 Hot Reload (Air)

Backend menggunakan [Air](https://github.com/air-verse/air) v1.65.2 untuk auto rebuild.

| Setting | Value |
|---------|-------|
| File dipantau | `.go`, `.env`, `.mod`, `.sum` |
| Direktori | `.` (root), `internal/`, `cmd/` |
| Build output | `./tmp/hrms-server` |
| Shutdown | `SIGINT` (graceful) + 500ms kill delay |

Setiap perubahan kode akan: build ulang → kirim SIGINT → jalankan binary baru.

```bash
cd backend
air
```

---

## ✅ Sesi Ini — 2 Juli 2026

### 🗄️ Migration 00037 — Position Grades Salary Range
- Tambah kolom `min_salary` dan `max_salary` (DECIMAL nullable) ke tabel `position_grades`
- Seed data: setiap level grade dapat range gaji realistis (Staff Rp4-7jt → President Director Rp75-150jt)
- Backend: model, repository, handler di-update untuk include min_salary/max_salary
- Frontend: column AG Grid + form fields di halaman golongan-jabatan

### 📊 Attendance Export Excel
- Backend: `ExportReport()` di attendance_record_service.go — generate Excel via excelize
- Backend: Handler `GET /api/attendance/report/export` dengan RBAC
- Frontend: Tombol "Export Excel" di halaman absensi → download .xlsx

### 💰 Salary Range di Form Karyawan
- Load position_grades saat mount
- Saat pilih posisi, tampilkan info card rentang gaji (format Rp)\n- Card hilang otomatis jika grade tidak punya data gaji

### 🎯 Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | ✅ 0 errors |
| `go vet ./...` | ✅ 0 issues |
| `svelte-check` | ✅ **0 errors, 0 warnings** |
| `npx playwright test` | ✅ **6 passed** |

---

## ✅ Payroll Engine — Sudah Selesai

### 🎯 Scope: Penggajian Fleksibel — Harian, Pokok, Tunjangan, Potongan

Sistem penggajian dirancang **sangat fleksibel** untuk memenuhi berbagai kebutuhan perusahaan Indonesia:

### ✅ Database Schema (4 Migrations)

| Migration | Isi |
|-----------|-----|
| `00013_payroll.sql` | `payroll_periods`, `payroll_items` (base_salary, daily_wage, allowances JSONB, deductions JSONB, overtime_pay, thr_amount, bpjs, loan, pph21, net_salary, company_cost), `payslip_view` |
| `00017_employee_salary_components.sql` | `employee_salary_components` (master data komponen per karyawan + history), trigger auto-log perubahan |
| `00018_flexible_overtime_rates.sql` | 3-level hierarchy: company default → position grade override → employee override |
| `00019_flexible_bpjs_config.sql` | Configurable BPJS rates via `companies.hr_settings`, per-employee override via `employees.bpjs_config` |

### ✅ Salary Components CRUD

**Backend (4 file):**
- `models/salary_component.go` — struct, summary, request/response
- `repository/salary_component_repo.go` — CRUD queries + `WithUserContext` untuk trigger audit
- `service/salary_component_service.go` — validasi (component_name, type allowance/deduction, amount >= 0)
- `handlers/salary_component_handler.go` — 4 endpoints (list, create, update, delete)

**Frontend:**
- `api.js` — `salaryComponents` export (list, create, update, remove)
- `/karyawan/detail` (EmployeeDetail.svelte) — Komponen Gaji: inline form, summary cards (Tunjangan/Potongan/Bersih), tabel + mobile cards

### ✅ Payroll Engine Backend — Sudah Dibuat

| File | Fungsi | Baris |
|------|--------|-------|
| `models/payroll.go` | PayrollPeriod, PayrollItem, CreatePeriodRequest, CalculateRequest, Payslip response, PPh21Calculation | 156 |
| `repository/payroll_repo.go` | CRUD payroll_periods, payroll_items, get latest base salary, get active salary components, call `calculate_employee_payroll()` function | 494 |
| `service/payroll_service.go` | Period management (create, approve, pay), calculate payroll (panggil PL/pgSQL function), PPh 21 TER di Go, THR calculation, payslip generation | 305 |
| `handlers/payroll_handler.go` | HTTP handlers: 8 endpoints (list/get/create period, calculate, list items, approve, pay, get payslip) | 194 |
| **Total** | **4 file — Payroll engine backend lengkap** | **1.149** |

### ✅ Payroll Engine Frontend — Sudah Dibuat

| Halaman | Route | Fitur |
|---------|-------|-------|
| Periode Penggajian | `/penggajian` | Daftar periode (bulan/tahun), status (draft/calculated/approved/paid), tombol Calculate/Approve/Pay |
| Detail Payroll | `/penggajian/[id]` | Tabel payroll items per karyawan, gross/net/deduction totals, filter |
| Slip Gaji (Payslip) | `/penggajian/payslip/[periodId]/[employeeId]` | Detail slip gaji per karyawan, income/deduction breakdown |
| Payslip Self-service | `/penggajian/slip-saya` | ✅ **Selesai** |

### 🧮 Alur Perhitungan Gaji

```
Input dari Master Data:
  ├── Gaji Pokok (base_salary dari employee_salary_histories)
  ├── Upah Harian (daily_wage, untuk karyawan harian)
  ├── Tunjangan Aktif (allowances dari employee_salary_components)
  ├── Potongan Aktif (deductions dari employee_salary_components)
  ├── Lembur (overtime_pay dari overtime_calculation view)
  ├── Pinjaman (loan_deduction dari loan_installments due)
  └── THR (jika bulan Juni / sesuai company settings)

Perhitungan Otomatis (PL/pgSQL function):
  Gross Salary = base_salary + total_tunjangan + lembur + THR + bonus
  BPJS (Pekerja) = Kesehatan (1%) + JHT (2%) + JP (1%) — ceiling aware
  PPh 21 = HITUNG di Go dengan TER (Tarif Efektif Rata-rata)
  Total Potongan = BPJS + PPh 21 + Pinjaman + deductions_lain
  Net Salary (Take Home Pay) = Gross - Total Potongan
  Company Cost = BPJS Perusahaan (Kesehatan 4% + JHT 3.7% + JP 2% + JKK + JKM)
```

### 💰 Fitur Fleksibel yang Didukung

| Fitur | Implementasi |
|-------|-------------|
| **Gaji Harian** | Kolom `daily_wage` + `total_days_worked` di payroll_items. Untuk karyawan dengan status `harian` |
| **Gaji Pokok** | `base_salary` dari `employee_salary_histories` (riwayat perubahan gaji tercatat) |
| **Tunjangan** | JSONB `allowances` — bisa berbagai macam (transport, makan, komunikasi, jabatan, dll) dari `employee_salary_components` |
| **Potongan** | JSONB `deductions` — pinjaman, kasbon, denda, iuran, dari `employee_salary_components` |
| **Prorate** | `total_days_worked` — untuk karyawan baru/keluar di tengah bulan |
| **Lembur** | Flexible rate 3-level (company → position grade → employee), perhitungan via `overtime_calculation` view |
| **BPJS** | Configurable rate + ceiling di `companies.hr_settings`, override per karyawan di `bpjs_config` |
| **PPh 21 TER** | Perhitungan di Go dengan TER category A/B/C, recap Desember |
| **THR** | Full (>=12 bulan) atau proporsional, dihitung dari gaji pokok + tunjangan tetap |

---

## ✅ Yang Telah Diselesaikan (Ringkasan)

### ✅ Payroll Engine — Backend & Frontend
- 4 file backend (1.149 baris): models, repository, service, handler — 8 endpoints
- 3 halaman frontend: daftar periode, detail payroll, payslip
- PPh 21 TER (Go), THR, BPJS configurable, overtime flexible rates

### ✅ Attendance Record — Check-in/out
- 4 file backend: check-in/out dengan validasi lokasi GPS, today status, my history, report
- 1 halaman frontend: `/absensi`

### ✅ Schedule Templates (Level 1)
- 4 file backend: template jadwal CRUD
- 1 halaman frontend: `/jadwal-templates`

### ✅ Employee Schedule (Level 2 & 3)
- Sama schedule_handler.go — CRUD employee schedules + resolve hierarki
- 1 halaman frontend: `/jadwal-karyawan`

### ✅ Response Format Standardization
- `handlers/response.go` — helper functions, semua handler di-standardisasi

### ✅ Database Trigger & Function Audit
- Fix audit_trigger_function DELETE bug, tambah audit 9 tabel, migration 00021

### ✅ Enkripsi Data Sensitif (NIK, NPWP, Bank)
- AES-256 via pgcrypto, encrypt_sensitive() di INSERT/UPDATE, decrypt_sensitive() di SELECT

### ✅ Docker Compose Setup
- 4 services: db, migrate, api, web + password encoding fix

### ✅ Shift Change Request
- 7 endpoints backend, full frontend page dengan approve/reject/cancel/confirm-swap

### ✅ AG Grid Semua Halaman
- Semua halaman pakai ag-grid-community, custom cell renderers, dark theme

### ✅ Dark Mode
- 200+ baris CSS overlay, toggle di topbar, localStorage persistence

### ✅ Overtime & Reimbursement — Seed Data + Frontend Test (25 Juni 2026)

**Seed data di database:**
- 5 overtime requests (Budi approved, Siti pending, Andi pending, Dewi approved, Rudi pending)
- 5 reimbursement requests (Budi approved, Siti pending, Andi approved, Dewi pending, Hendra paid)
- Permissions: `overtime` & `reimbursement` modules added to super_admin, hr_manager, hr_staff, finance, employee

**Frontend test via browser:**
- `/lembur` — AG Grid menampilkan 5 data, detail view, approve/reject/cancel buttons ✅
- `/reimbursement` — AG Grid menampilkan 5 data, form ajukan dengan drag & drop receipt upload ✅

**Bug fixes di seed:**
- `approvalEntry` JSON — pake `adminUUID` langsung (bukan SQL subquery di JSON)
- Reimbursement `paid` status — ganti ke `approved` + `paid_at`/`paid_by` (enum `leave_status` gak punya `paid`)
- Audit trigger disable/enable buat `overtime_requests` & `reimbursements`
- Reimbursement seed refactored ke parameterized queries

### ✅ TypeScript Clean — 0 Errors

Semua TypeScript strict null check issues telah diperbaiki:

| File | Fix |
|------|-----|
| `EmployeeDetail.svelte` | Type assertion `as EmployeeDetail` setelah reassignment dari API response (`any`) |
| `jadwal-templates/+page.svelte` | `class:hidden` → `{#if showDetail && detailItem}` untuk type narrowing |
| `jadwal-kerja/+page.svelte` | `class:hidden` → `{#if}` + fix bug Edit button (null dereference) |
| `lokasi-absensi/+page.svelte` | Install `@types/leaflet` untuk module declaration |

---

## ✅ Sesi Ini — 2 Juli 2026 (Migration 00037 + Salary Range + Export Excel + E2E Tests) (Final — Leave & Overtime Bug Fix + API Verification)

### 🐛 Fix Leave Approve/Reject — 3 Perubahan di leave_repo.go

| Perubahan | Sebelum | Sesudah |
|-----------|---------|---------|
| **rejected_at** | Tidak di-set | `CASE WHEN status='rejected' THEN NOW() ELSE rejected_at END` |
| **rejected_by** | Tidak di-set | `CASE WHEN status='rejected' THEN $5::uuid ELSE rejected_by END` |
| **Balance error** | `_ = err` (ditelan) | `fmt.Printf("[WARN] ...")` (di-log) |
| **Balance param** | Ada `$1` (id) tidak terpakai | Renumber $1-$4, id dihapus |

### 🐛 Fix Overtime GET 404 — Missing Columns (Migration 00033)

**Root cause:** `GetOvertimeRequestByID` di repo mereferensi `otr.rejection_reason`, `otr.cancelled_at`, `otr.cancelled_by` yang tidak ada di tabel `overtime_requests` (migration 00008). Sama seperti bug `leave_requests` sebelumnya.

**Fix:** Migration `00033_add_overtime_missing_columns.sql`:
- `rejection_reason TEXT DEFAULT ''`
- `rejected_at TIMESTAMPTZ`
- `rejected_by UUID REFERENCES employees(id)`
- `cancelled_by UUID`
- `cancelled_at TIMESTAMPTZ`

Plus `UpdateOvertimeStatus` di overtime_repo.go: tambah `rejected_at`/`rejected_by` CASE/WHEN pada reject.

### 🐛 Fix Audit Trigger — Zero UUID Foreign Key Violation

**Root cause:** `app.current_user_id` default `00000000-0000-0000-0000-000000000000` di-set saat koneksi database, tapi trigger audit mencoba INSERT UUID ini ke `activity_logs(user_id)` yang memiliki FK `REFERENCES employees(id)`.

**Fix:** Fungsi `audit_trigger_function()` di-update: jika `user_id_val = '00000000-...'` maka set ke `NULL`.

### 🐛 Fix JSON `\"` Escaping di 4 Repository Files

`leave_repo.go`, `overtime_repo.go`, `reimbursement_repo.go`, `shift_change_repo.go`:
- `\"` dalam raw string Go → `"` (valid JSON untuk `::jsonb`)
- `"NOW()"` literal → `time.Now().Format(time.RFC3339)`

### ✅ API Verification Results

| API | Endpoint | Result |
|-----|----------|--------|
| **Cuti Approve** | `PUT /api/leaves/:id/approve` | ✅ **Berhasil** |
| **Cuti Reject** | `PUT /api/leaves/:id/reject` | ✅ **Berhasil** (dengan rejection_reason) |
| **Overtime GET** | `GET /api/overtime-requests/:id` | ✅ **Berhasil** (setelah migration 00033) |
| **Overtime Approve** | `PUT /api/overtime-requests/:id/approve` | ✅ **Berhasil** |
| **Reimbursement Approve** | `PUT /api/reimbursements/:id/approve` | ✅ **Berhasil** |
| **Reimbursement Pay** | `PUT /api/reimbursements/:id/pay` | ✅ **Berhasil** |

### 📊 Build Status

| Area | Hasil |
|------|-------|
| **Backend `go build ./...`** | ✅ **Bersih** |
| **Backend `go vet ./...`** | ✅ **Bersih** |
| **Backend `go test ./...`** | ✅ **Lulus** |
| **Database Migrations** | ✅ 33 migrations applied (00001-00033) |

---

## ✅ Sesi Ini — 3 Juli 2026 (Final — Migration 41, Seed Daily Journal, A11y Fixes, CI Clean ✅)

### 🗄️ Migration 38-41 — Berhasil di-apply

| Migration | Status |
|-----------|--------|
| `00038_cleanup_payroll_status_enum.sql` | ✅ **Fix: handle payslip_view + DEFAULT values** |
| `00039_transaction_safety_documentation.sql` | ✅ Dokumentasi |
| `00040_daily_journal.sql` | ✅ Tabel `daily_journals` + enum `journal_status` |
| `00041_add_new_module_permissions.sql` | ✅ Permission `reprimand`, `daily_journal`, `report` |

### 🐛 Migration 00038 Fix — Dependency Handling

**Masalah:** Migration gagal karena `payslip_view` + DEFAULT values (`'draft'::payroll_status`) mereferensi tipe `payroll_status`.

**Fix:**
- `DROP VIEW IF EXISTS payslip_view CASCADE` diawal Up/Down
- `ALTER COLUMN status DROP DEFAULT` sebelum alter type
- `ALTER COLUMN status SET DEFAULT 'draft'::payroll_status` setelah recreate type
- `CREATE OR REPLACE VIEW payslip_view` di akhir Up/Down
- Semua migration lain dicek — **tidak ada masalah serupa** ✅

### 🐛 Seed Fix — Shift Change FK Constraint

**Masalah:** `requested_schedule_id` seed menggunakan `schedule_template_days.id`, tapi FK mereferensi `work_schedules(id)`.

**Fix:** Ganti source ID dari `schedule_template_days` → `work_schedules`

### 📝 Seed Daily Journal — 8 Entries

| Employee | Entri | Status |
|----------|-------|--------|
| EMP-001 Budi | 3 entri | 1 submitted, 2 acknowledged |
| EMP-002 Siti | 2 entri | 1 submitted, 1 acknowledged |
| EMP-004 Dewi | 1 entri | submitted |
| EMP-005 Rudi | 1 entri | draft |
| EMP-007 Rina | 1 entri | submitted |

### ♿ A11y Fixes

| Halaman | Fix |
|---------|-----|
| Notifikasi (+layout) | Tambah `id` + `name` ke menu search inputs (desktop & mobile) |
| Jurnal Harian | Tambah `name` + `aria-label` ke filter form fields |
| Notifikasi (pagination) | Tambah `aria-label` ke prev/next buttons (sebelumnya) |

### ✅ Build Status (Final)

| Check | Result |
|-------|--------|
| `go build ./...` | ✅ 0 errors |
| `go vet ./...` | ✅ 0 issues |
| `go test ./...` | ✅ All pass |
| `svelte-check` | ✅ **0 errors, 0 warnings** |
| `npm run build` | ✅ Sukses |
| `playwright test` | ✅ **6/6 passed** |
| `make ci-check` | ✅ **CI check passed** |

### ✅ Full UI Verification

| Halaman | Status |
|---------|--------|
| Login | ✅ |
| Dashboard | ✅ Charts & metrics |
| Karyawan | ✅ |
| Absensi | ✅ |
| Penggajian | ✅ Payroll periods |
| Jurnal Harian | ✅ 8 data seed muncul |
| Notifikasi | ✅ |
| Laporan | ✅ |
| Audit Trail | ✅ |
| Departemen | ✅ |

---
*Update file ini saat ada progress baru dengan menambahkan entry di atas.*
