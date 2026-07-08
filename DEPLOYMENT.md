# 🚀 HRMS Application — Deployment Guide

**Stack:** Go (Fiber) + PostgreSQL 16 + SvelteKit (SPA) + Nginx  
**Versi:** 1.0.0  
**Tanggal:** 8 Juli 2026

---

## 📋 Prasyarat

| Requirement | Versi Minimal | Catatan |
|-------------|---------------|---------|
| Docker | 24+ | [Install Docker](https://docs.docker.com/engine/install/) |
| Docker Compose | 2.20+ | [Install Docker Compose](https://docs.docker.com/compose/install/) |
| Domain + SSL | — | Gunakan Let's Encrypt / Cloudflare |
| SMTP Server | — | Untuk notifikasi email (opsional) |

---

## 🏗️ Arsitektur Deployment

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Browser   │────▶│   Nginx      │────▶│  Go/Fiber    │
│   (PWA)     │     │  (Frontend)  │     │  (Backend)   │
└─────────────┘     │  Port 80/443  │     │  Port 8080   │
                    └──────────────┘     └──────┬───────┘
                                                │
                                        ┌───────▼───────┐
                                        │  PostgreSQL   │
                                        │  Port 5432    │
                                        └───────────────┘
```

### Komponen

| Service | Container | Port | Image |
|---------|-----------|------|-------|
| **Database** | `hrms-db` | 5432 (internal) | `postgres:16-alpine` |
| **Backend API** | `hrms-api` | 8080 | `hrms-api:latest` |
| **Frontend** | `hrms-web` | 80 | `hrms-web:latest` |
| **Migrate** | `hrms-migrate` | — (one-shot) | `hrms-migrate:latest` |

---

## ⚙️ 1. Konfigurasi Lingkungan

### 1.1 Clone Repository

```bash
git clone https://github.com/company/hrms.git
cd hrms
```

### 1.2 Buat File `.env`

```bash
cp .env.example .env
```

**Minimal config untuk production:**

```env
# ─── Database ───
DB_HOST=db
DB_PORT=5432
DB_USER=hrms_user
DB_PASSWORD=<GENERATE_STRONG_PASSWORD>
DB_NAME=hrms
DB_SSLMODE=require

# ─── JWT ───
JWT_SECRET=<GENERATE_64_CHAR_HEX>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# ─── Encryption (AES-256) ───
ENCRYPTION_KEY=<GENERATE_32_CHAR_HEX>

# ─── Server ───
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
FRONTEND_URL=https://hrms.example.com

# ─── Storage ───
UPLOAD_DIR=./uploads

# ─── SMTP (opsional) ───
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=<SMTP_PASSWORD>
SMTP_FROM=noreply@hrms.example.com
SMTP_FROM_NAME=HRMS System
```

**Generate secrets:**

```bash
# JWT Secret (64 hex chars = 32 bytes)
openssl rand -hex 32

# Encryption Key (32 hex chars = 16 bytes untuk AES-256)
openssl rand -hex 16

# DB Password (16 chars)
openssl rand -base64 16
```

---

## 🐳 2. Docker Build & Deploy

### 2.1 Build Images

```bash
# Build backend production image
docker build -t hrms-api:latest --target prod -f Dockerfile .

# Build frontend production image (Nginx)
docker build -t hrms-web:latest --target prod -f frontend/Dockerfile .
```

### 2.2 Deploy with Docker Compose

Buat file `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    container_name: hrms-db
    restart: unless-stopped
    env_file: .env
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Build a temporary migration image that includes Goose + SQL files
  migrate:
    build:
      context: .
      target: base
      dockerfile: Dockerfile
    container_name: hrms-migrate
    depends_on:
      db:
        condition: service_healthy
    env_file: .env
    environment:
      DB_HOST: db
      DB_PORT: "5432"
    volumes:
      - ./database/migrations:/app/database/migrations
    command: >
      sh -c '
        PASSWORD_ENCODED=$$(echo "${DB_PASSWORD}" | sed "s|/|%2F|g")
        goose -dir /app/database/migrations postgres "postgres://${DB_USER}:$${PASSWORD_ENCODED}@db:5432/${DB_NAME}?sslmode=${DB_SSLMODE}" up
      '

  api:
    image: hrms-api:latest
    container_name: hrms-api
    restart: unless-stopped
    depends_on:
      migrate:
        condition: service_completed_successfully
    env_file: .env
    environment:
      DB_HOST: db
      DB_PORT: "5432"
    volumes:
      - uploads:/app/uploads
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s

  web:
    image: hrms-web:latest
    container_name: hrms-web
    restart: unless-stopped
    ports:
      - "80:80"
      # Optional: SSL via reverse proxy (Nginx on host or Cloudflare Tunnel)
      # - "443:443"
    depends_on:
      api:
        condition: service_started
    environment:
      - PUBLIC_API_BASE_URL=https://api.hrms.example.com
    volumes:
      - ./nginx/ssl:/etc/nginx/ssl:ro

volumes:
  pgdata:
  uploads:
```

**Deploy:**

```bash
docker compose -f docker-compose.prod.yml up -d
```

### 2.3 Verifikasi

```bash
# Cek status semua container
docker compose -f docker-compose.prod.yml ps

# Cek health API
curl https://hrms.example.com/api/health

# Cek frontend
curl -I https://hrms.example.com
```

---

## 🌐 3. Nginx Configuration (Production)

Jika tidak menggunakan frontend container, deploy frontend build ke Nginx sendiri:

```nginx
# /etc/nginx/sites-available/hrms
server {
    listen 443 ssl http2;
    server_name hrms.example.com;

    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;

    # Root for SvelteKit static build
    root /var/www/hrms;
    index index.html;

    # Security headers
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "strict-origin-when-cross-origin";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    # Gzip
    gzip on;
    gzip_comp_level 5;
    gzip_min_length 256;
    gzip_types text/plain text/css application/json application/javascript text/xml image/svg+xml;
    gzip_vary on;

    # Cache immutable chunks
    location /_app/immutable/ {
        expires 1y;
        add_header Cache-Control "public, immutable, max-age=31536000";
    }

    location /_app/ {
        expires 30d;
        add_header Cache-Control "public, max-age=2592000";
    }

    # PWA assets
    location /service-worker.js {
        expires -1;
        add_header Cache-Control "no-cache, must-revalidate";
    }

    location /manifest.json {
        expires 7d;
        add_header Cache-Control "public, max-age=604800";
    }

    # Static assets
    location /icons/ {
        expires 30d;
        add_header Cache-Control "public, max-age=2592000";
    }

    # API proxy
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Uploads
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
    }

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }
}

server {
    listen 80;
    server_name hrms.example.com;
    return 301 https://$server_name$request_uri;
}
```

---

## 📦 4. GitHub Actions Deployment

Buat file `.github/workflows/deploy.yml`:

```yaml
name: Deploy

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Backend
        run: docker build -t hrms-api:latest --target prod -f Dockerfile .
        
      - name: Build Frontend
        run: docker build -t hrms-web:latest --target prod -f frontend/Dockerfile .
        
      - name: Save Images as Tarball
        run: |
          docker save hrms-api:latest hrms-web:latest | gzip > hrms-images.tar.gz

      - name: Copy Images to Server
        uses: appleboy/scp-action@v1
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USERNAME }}
          key: ${{ secrets.SSH_KEY }}
          source: "hrms-images.tar.gz"
          target: "/opt/hrms/"

      - name: Deploy to Server
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USERNAME }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd /opt/hrms
            docker load < hrms-images.tar.gz
            rm hrms-images.tar.gz
            docker compose -f docker-compose.prod.yml up -d --force-recreate
            docker system prune -f
```

---

## 🔐 5. Security Checklist

- [ ] Semua secrets di `.env` tidak pernah di-commit
- [ ] JWT_SECRET minimal 64 karakter hex
- [ ] ENCRYPTION_KEY menggunakan 32 karakter hex
- [ ] DB_PASSWORD kuat (min 16 karakter)
- [ ] SSL/TLS aktif (HTTPS only)
- [ ] HSTS header di-set (Strict-Transport-Security)
- [ ] CSP header aktif (Content-Security-Policy)
- [ ] Rate limiting diaktifkan
- [ ] File upload validation aktif
- [ ] Backup database teratur (cron job)
- [ ] UFW firewall hanya buka port 22, 80, 443

---

## 💾 6. Backup & Recovery

### 6.1 Database Backup (Otomatis via Cron)

```bash
# Setiap hari jam 2 pagi
0 2 * * * /opt/hrms/scripts/backup_db.sh >> /var/log/hrms-backup.log 2>&1
```

### 6.2 Restore Database

```bash
# List backups
ls -la /opt/hrms/backups/

# Restore
pg_restore --clean --no-owner \
  --dbname="postgres://hrms_user:password@localhost:5432/hrms" \
  /opt/hrms/backups/hrms_20260708_020000.dump
```

---

## 📊 7. Monitoring

### Health Check Endpoint

```bash
# API Health
GET /api/health → 200 OK

# Monitoring dengan Uptime Kuma / Better Stack
# Monitor: https://hrms.example.com/api/health
```

### Logging

```bash
# Docker logs
docker compose -f docker-compose.prod.yml logs -f
docker compose -f docker-compose.prod.yml logs -f api --tail 100

# Nginx access log
tail -f /var/log/nginx/hrms.access.log
```

---

## 🐛 8. Troubleshooting

| Masalah | Solusi |
|---------|--------|
| Container restart loop | `docker compose -f docker-compose.prod.yml logs [service]` untuk melihat error |
| Database connection refused | Pastikan `DB_HOST` dan `DB_PORT` benar, database running |
| 502 Bad Gateway | Backend belum siap, cek `docker compose logs api` |
| File upload gagal | Periksa permission `uploads/` directory |
| CORS error | Pastikan `FRONTEND_URL` di .env sesuai dengan domain frontend |
| SSL error | Renew Let's Encrypt: `certbot renew` |
| Migration failed | Rollback: `goose down`, fix migration, `goose up` lagi |

---

## 🔄 9. Rollback

### Rollback Backend

```bash
# Deploy versi sebelumnya
docker build -t hrms-api:previous --target prod -f Dockerfile .
docker compose -f docker-compose.prod.yml up -d api
```

### Rollback Database Migration

```bash
# Turunkan 1 migration terakhir
docker exec hrms-migrate goose -dir /app/database/migrations down

# Atau ke versi spesifik
docker exec hrms-migrate goose -dir /app/database/migrations down-to 00040
```

---

## ⚡ 10. Performance Optimization

| Area | Optimasi |
|------|----------|
| **Frontend** | Gunakan CDN (Cloudflare) untuk static assets |
| **Database** | Pastikan indeks lengkap (migration 00015) |
| **API** | Aktifkan response compression (Gzip/Brotli) |
| **Caching** | Cache headers untuk static assets (1 year immutable) |
| **PWA** | Service worker untuk offline-first experience |
| **Images** | Optimasi upload foto (max 2MB, resize to 800px) |
| **AG Grid** | Virtual scrolling untuk tabel besar (100k+ rows) |

---

## 📝 11. Deployment Checklist (Go-Live)

### Pre-Deployment
- [ ] Backend: `go build ./...` ✅
- [ ] Backend: `go vet ./...` ✅
- [ ] Backend: `go test -short ./...` ✅
- [ ] Frontend: `svelte-check` 0 errors ✅
- [ ] Frontend: `npm run build` sukses ✅
- [ ] CI workflow passing ✅
- [ ] E2E tests passing ✅
- [ ] Semua migration sudah di-test ✅
- [ ] Data seed sudah benar ✅

### Deployment
- [ ] Buat `.env` production
- [ ] Generate JWT_SECRET dan ENCRYPTION_KEY
- [ ] Build Docker images
- [ ] SSH ke server
- [ ] Pull images dan deploy
- [ ] Run migration
- [ ] Seed data (jika first deploy)
- [ ] Verifikasi health endpoint
- [ ] Verifikasi login flow
- [ ] Verifikasi fitur utama (absensi, cuti, payroll)

### Post-Deployment
- [ ] Setup backup cron job
- [ ] Setup monitoring (Uptime Kuma / Better Stack)
- [ ] Setup SSL auto-renewal
- [ ] Test PWA installability
- [ ] Verifikasi mobile responsiveness
- [ ] Performance testing (Lighthouse)
- [ ] Security scan

---

*Dokumen ini diperbarui secara berkala. Last updated: 8 Juli 2026.*
