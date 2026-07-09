# HRMS — Human Resource Management System

Aplikasi manajemen HR berbasis web dengan backend Go (Fiber) dan frontend SvelteKit.

---

## 📦 Architecture

```
┌─────────────────────────────────────┐
│         Single Binary (prod)        │
│  ┌─────────────┐  ┌──────────────┐ │
│  │ Go/Fiber    │  │ SvelteKit    │ │
│  │ API Server  │◄─┤ Static Build │ │
│  │ :8080       │  │ ./public/    │ │
│  └──────┬──────┘  └──────────────┘ │
└─────────┼──────────────────────────┘
          │
    ┌─────▼──────┐
    │ PostgreSQL │
    │ :5432      │
    └────────────┘
```

- **Backend**: Go + Fiber v2 — REST API + SPA file server
- **Frontend**: SvelteKit 5 (adapter-static) — SPA with client-side routing
- **Database**: PostgreSQL 16
- **Development**: Docker Compose with Air hot reload + Vite HMR
- **Production**: Single Docker image containing both backend binary and built frontend

---

## 🚀 Quick Start

### Development Mode (with HMR)

```bash
# Start all services (backend API :8080 + frontend HMR :5177)
docker compose --profile dev up
```

### Production Mode (single binary)

```bash
# 1. Build frontend + copy to backend/public/
make frontend-build

# 2. Start backend (serves frontend too)
docker compose up
# → API & Frontend: http://localhost:8080
```

### Manual Build

```bash
# Build frontend
make frontend-build

# Build & run Go backend
cd backend && go build -o hrms-server . && ./hrms-server
```

---

## 🛠️ Development

### Prerequisites

- Go 1.26+
- Node.js 22+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker)

### Available Commands

| Command | Description |
|---------|-------------|
| `make dev-backend` | Start Go backend with Air hot reload |
| `make dev-frontend` | Start SvelteKit dev server (port 5173) |
| `make build` | Build all Go binaries |
| `make vet` | Run `go vet` |
| `make test` | Run unit tests |
| `make frontend-build` | Build frontend & copy to `backend/public/` |
| `make docker-up` | Start all Docker services |
| `make ci-check` | Full CI pipeline (vet + build + test + frontend) |

---

## 🔒 Security Headers

Semua security headers dikelola oleh `SecurityHeadersMiddleware` di `backend/internal/middleware/security.go`:

| Header | Value | Description |
|--------|-------|-------------|
| `X-Content-Type-Options` | `nosniff` | Mencegah MIME sniffing |
| `X-Frame-Options` | `DENY` | Mencegah clickjacking |
| `X-XSS-Protection` | `1; mode=block` | Mencegah XSS |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Mengontrol referrer info |
| `Permissions-Policy` | `geolocation=(self), camera=(self), ...` | Mengontrol API permissions |
| `Content-Security-Policy` | Custom per env | Mencegah XSS & injection |
| `X-DNS-Prefetch-Control` | `off` | Mencegah DNS prefetching |
| `X-Download-Options` | `noopen` | Mencegah download otomatis |
| `X-Permitted-Cross-Domain-Policies` | `none` | Mencegah Flash cross-domain |
| `Origin-Agent-Cluster` | `?1` | Isolasi process per origin |

> **Note**: Fiber `helmet.New()` tidak digunakan karena defaultnya mengirim `Cross-Origin-Embedder-Policy: require-corp` dan `Cross-Origin-Resource-Policy: same-origin` yang memblokir upload gambar.
> `SecurityHeadersMiddleware` mencakup semua header kritis tanpa efek samping tersebut.

---

## 📁 CI/CD

GitHub Actions workflow di `.github/workflows/ci.yml` menjalankan:

1. **Backend** — `go build`, `go vet`, unit test, integration test
2. **Frontend** — `svelte-check`, `npm run build`
3. **Docker** — Build production image (frontend + backend in single image)

---

## 📚 Documentation

- [Technical Specification](TECHNICAL-SPEC.md)
- [API Routes](TECHNICAL-SPEC.md#api-routes)
- [Database Schema](database/schema-overview.md)
- [Product Requirements](PRD-HRMS-Application.md)
- [Project Status](PROJECT-STATUS.md)
- [Roadmap](ROADMAP.md)
