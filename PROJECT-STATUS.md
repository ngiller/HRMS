# HRMS Application — Project Status

> **File ini menyimpan riwayat progress project.**  
> Dibaca oleh AI agent untuk melanjutkan pekerjaan dari sesi sebelumnya.  
> Update file ini setiap kali ada progress signifikan.

**Last Updated:** 11 Juni 2026  
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
| **Backend Implementation (Go/Fiber)** | 🚀 **Dimulai** | 30% |
| **Frontend Implementation (SvelteKit)** | 🚀 **Dimulai** | 25% |

---

## 📄 File yang Dibuat

### Desain
| File | Deskripsi | Status |
|------|-----------|--------|
| `PRD-HRMS-Application.md` | PRD — 14 sections | ✅ Final |
| `TECHNICAL-SPEC.md` | Technical Spec — 20 sections | ✅ Final |
| `database/schema-overview.md` | ER diagram, index strategy, payroll logic | ✅ Final |
| `database/migrations/00001-00020.sql` | 20 migration files | ✅ Final |
| `ui-mockups/index.html` | 11 UI wireframes | ✅ Final |
| `PROJECT-STATUS.md` | Progress tracker | ✅ Active |

### Backend (Go/Fiber)
| File | Deskripsi | Status |
|------|-----------|--------|
| `backend/.env` | Database & JWT config | ✅ Done |
| `backend/main.go` | Entry point: Fiber server, routes, middleware | ✅ Done |
| `backend/internal/config/config.go` | Environment config loader | ✅ Done |
| `backend/internal/database/database.go` | pgx connection pool | ✅ Done |
| `backend/internal/models/auth.go` | Auth models (login, forgot/reset password) | ✅ Done |
| `backend/internal/models/employee.go` | Employee models (detail, summary, dashboard) | ✅ Done |
| `backend/internal/repository/auth_repo.go` | Auth queries (login, password reset, login attempts) | ✅ Done |
| `backend/internal/repository/employee_repo.go` | Employee queries (list, detail, dashboard stats) | ✅ Done |
| `backend/internal/service/auth_service.go` | Auth logic (bcrypt, JWT, forgot/reset) | ✅ Done |
| `backend/internal/service/employee_service.go` | Employee logic (list, detail, dashboard) | ✅ Done |
| `backend/internal/handlers/auth_handler.go` | Auth HTTP handlers (login, forgot/reset, me) | ✅ Done |
| `backend/internal/handlers/employee_handler.go` | Employee HTTP handlers (list, get, dashboard) | ✅ Done |
| `backend/internal/middleware/auth.go` | JWT Bearer middleware + CORS | ✅ Done |
| `backend/cmd/seed/main.go` | Seed script (admin + 8 test employees) | ✅ Done |
| `backend/.air.toml` | Air hot reload config — auto rebuild saat file .go berubah | ✅ Done |

### Frontend (SvelteKit)
| File | Deskripsi | Status |
|------|-----------|--------|
| `frontend/src/lib/config.js` | API configuration (backend URL) | ✅ Done |
| `frontend/src/lib/api.js` | API client (fetch + JWT auth + error handling) | ✅ Done |
| `frontend/src/routes/login/+page.svelte` | Login page — terhubung ke backend API | ✅ Done |
| `frontend/src/routes/forgot-password/+page.svelte` | Forgot password — input email + success state | ✅ Done |
| `frontend/src/routes/reset-password/+page.svelte` | Reset password — token from URL, confirm password | ✅ Done |
| `frontend/src/routes/dashboard/+page.svelte` | Dashboard — stat cards, charts, approvals | ✅ Done |
| `frontend/src/routes/dashboard/+layout.svelte` | Dashboard layout — sidebar, topbar, user dropdown | ✅ Done |

---

## 🔌 API Endpoints — Sudah Berfungsi

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| `GET` | `/api/health` | ❌ | Health check |
| `POST` | `/api/auth/login` | ❌ | Login (return JWT + user data) |
| `POST` | `/api/auth/forgot-password` | ❌ | Kirim reset token |
| `POST` | `/api/auth/reset-password` | ❌ | Reset password dengan token |
| `GET` | `/api/auth/me` | ✅ | Data user yang login |
| `GET` | `/api/employees` | ✅ | Daftar karyawan (paginated) |
| `GET` | `/api/employees/:id` | ✅ | Detail karyawan |
| `GET` | `/api/dashboard` | ✅ | Statistik dashboard |

### Akun Test
| Role | Email | Password |
|------|-------|----------|
| Super Admin | `admin@company.com` | `admin123` |
| Employee | `budi@company.com`, `siti@company.com`, dll | `password123` |

---

## 🗺️ Arsitektur Terkini

```
Frontend (SvelteKit SPA) — localhost:5173
  └── config.js → API_BASE_URL = http://localhost:8080
        └── POST /api/auth/login ──→ Backend (Go / Fiber) — localhost:8080
              └── Middleware: JWT, CORS, Helmet, Logger
              └── Service: Auth (bcrypt + JWT)
              └── Repository: pgx/v5 parameterized queries
              └── Database: PostgreSQL 16 (hrms)
                    ├── 20 migrations (44 tables)
                    ├── 7 roles with permissions
                    ├── 9 employees (1 admin + 8 test)
                    └── Seed data ready
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

## 🚧 Rencana Selanjutnya

1. **Employee CRUD** — form create/edit karyawan
2. **Attendance Check-in** — halaman absensi dengan GPS
3. **Docker Compose** — PostgreSQL + backend + frontend + Air dalam satu command

---

*Update file ini saat ada progress baru dengan menambahkan entry di atas.*
