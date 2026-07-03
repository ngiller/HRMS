# GitHub Actions — Repository Secrets

> File ini mendokumentasikan environment variables yang perlu di-set sebagai **Repository Secrets** di GitHub Actions agar CI workflow bisa jalan.

## Cara Set Secrets

1. Buka repository di GitHub
2. **Settings → Secrets and variables → Actions**
3. Klik **"New repository secret"**
4. Tambahkan satu per satu variabel di bawah

---

## Required Secrets (Production)

Untuk deployment / production pipeline (jika ada), set secrets ini:

| Secret Name | Deskripsi | Contoh Value |
|-------------|-----------|-------------|
| `DB_HOST` | Hostname database PostgreSQL | `db.example.com` |
| `DB_PORT` | Port database | `5432` |
| `DB_USER` | Username database | `magnum` |
| `DB_PASSWORD` | Password database | `your-secure-password` |
| `DB_NAME` | Nama database | `hrms` |
| `DB_SSLMODE` | SSL mode koneksi | `require` |
| `JWT_SECRET` | Secret key untuk JWT token (min 32 chars) | `your-256-bit-secret-key-change-in-production` |
| `JWT_ACCESS_EXPIRY` | Expiry access token | `15m` |
| `JWT_REFRESH_EXPIRY` | Expiry refresh token | `168h` |
| `ENCRYPTION_KEY` | AES-256 key untuk enkripsi data sensitif di DB | `a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6` |
| `FRONTEND_URL` | URL frontend untuk CORS | `https://hrms.example.com` |
| `SERVER_HOST` | Host backend server | `0.0.0.0` |
| `SERVER_PORT` | Port backend server | `8080` |
| `UPLOAD_DIR` | Direktori upload file | `./uploads` |

---

## CI Workflow Secrets

CI workflow (`ci.yml`) sudah menggunakan **hardcoded test values** yang aman untuk di-commit:

| Variable | CI Value | Notes |
|----------|----------|-------|
| `GO_VERSION` | `1.26` | Bisa diubah jika versi Go berubah |
| `NODE_VERSION` | `22` | Bisa diubah jika versi Node berubah |
| `DB_HOST` | `localhost` | Dari service container PostgreSQL |
| `DB_PORT` | `5432` | Default PostgreSQL port |
| `DB_USER` | `magnum` | Hanya untuk test database |
| `DB_PASSWORD` | `Pass@w0rd` | Hanya untuk test database di CI |
| `DB_NAME` | `hrms_test` | Database terpisah untuk test |
| `DB_SSLMODE` | `disable` | Koneksi lokal di container, no SSL |
| `JWT_SECRET` | `ci-test-secret-key` | Bukan rahasia sesungguhnya |
| `JWT_ACCESS_EXPIRY` | `15m` | Standar |
| `JWT_REFRESH_EXPIRY` | `168h` | 7 hari |
| `ENCRYPTION_KEY` | `ci-encryption-key-for-testing-32bytes` | Hanya untuk test |

> **Tidak perlu set secrets untuk CI workflow saat ini** — semua test values sudah aman di-hardcode.

---

## Production Deployment Pipeline (Future)

Jika nanti menambahkan production deployment (misal deploy ke VPS / Docker registry), secrets berikut akan diperlukan:

### Docker Registry
| Secret | Deskripsi |
|--------|-----------|
| `DOCKER_USERNAME` | Username Docker Hub / GitHub Container Registry |
| `DOCKER_PASSWORD` | Token / Password registry |

### Deployment SSH (jika VPS)
| Secret | Deskripsi |
|--------|-----------|
| `SSH_HOST` | Alamat VPS |
| `SSH_USERNAME` | SSH user |
| `SSH_KEY` | Private SSH key (PEM format) |
| `SSH_PORT` | SSH port (default 22) |

### Environment (Production)
| Secret | Deskripsi |
|--------|-----------|
| `PROD_DB_URL` | Production database URL |
| `PROD_JWT_SECRET` | Production JWT secret |
| `PROD_ENCRYPTION_KEY` | Production encryption key |
| `PROD_FRONTEND_URL` | Production frontend URL |
| `SLACK_WEBHOOK` | (Optional) Notifikasi deploy via Slack |

---

## Security Notes

1. **JWT_SECRET** — Gunakan minimal 32 karakter. Generate dengan: `openssl rand -hex 32`
2. **ENCRYPTION_KEY** — Harus **32 hex characters** (16 bytes untuk AES-256). Generate dengan: `openssl rand -hex 16`
3. **DB_PASSWORD** — Gunakan password kuat (min 12 karakter, kombinasi huruf/angka/simbol)
4. **Jangan commit .env file** — File `.env` sudah di `.gitignore`
5. **Rotasi secrets** — Ganti secrets secara berkala (terutama JWT_SECRET dan ENCRYPTION_KEY)

---

## Quick Setup Checklist

- [ ] Go to Settings → Secrets → Actions
- [ ] Add `DOCKER_USERNAME` (jika deploy)
- [ ] Add `DOCKER_PASSWORD` (jika deploy)
- [ ] Add `SSH_HOST`, `SSH_USERNAME`, `SSH_KEY` (jika deploy VPS)
- [ ] Update `ci.yml` `GO_VERSION` jika versi Go berubah
- [ ] Update `ci.yml` `NODE_VERSION` jika versi Node berubah
- [ ] Run workflow manually untuk verifikasi
