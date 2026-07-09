# Security — HRMS Application

## Security Headers

Semua HTTP response dilengkapi security headers melalui `SecurityHeadersMiddleware` di `backend/internal/middleware/security.go`.

| Header | Value | Keterangan |
|--------|-------|------------|
| `X-Content-Type-Options` | `nosniff` | Mencegah MIME sniffing |
| `X-Frame-Options` | `DENY` | Mencegah clickjacking (tidak boleh di-frame) |
| `X-XSS-Protection` | `1; mode=block` | Mencegah XSS (legacy browser) |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Kirim referrer hanya saat same-origin |
| `Permissions-Policy` | `geolocation=(self), camera=(self), microphone=(), payment=()` | Kontrol API browser |
| `Content-Security-Policy` | Dynamic (lihat konfigurasi di `DefaultSecurityConfig()`) | Mencegah XSS & injection |
| `X-DNS-Prefetch-Control` | `off` | Mencegah DNS prefetching |
| `X-Download-Options` | `noopen` | Mencegah download otomatis (IE/Edge) |
| `X-Permitted-Cross-Domain-Policies` | `none` | Mencegah Flash/PDF cross-domain |
| `Origin-Agent-Cluster` | `?1` | Isolasi process per origin |

## Content Security Policy (CSP)

CSP diatur di `DefaultSecurityConfig()` dan mencakup:

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com;
style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com https://fonts.googleapis.com;
img-src 'self' data: blob: https://*.tile.openstreetmap.org https://unpkg.com;
connect-src 'self' http://localhost:8590 http://localhost:8900 https://api.emailjs.com ws: wss:;
font-src 'self' data: https://fonts.gstatic.com;
form-action 'self';
frame-ancestors 'none';
base-uri 'self'
```

### Catatan CSP

- `connect-src` secara eksplisit mencantumkan port localhost (`:8590`, `:8900`) karena Chrome tidak me-matching `http://localhost` (tanpa port) ke URL yang menggunakan non-default port.
- Untuk production, ganti `http://localhost:8590 http://localhost:8900` dengan domain aktual (contoh: `https://api.hrms.company.com`).
- `'unsafe-inline'` dan `'unsafe-eval'` di `script-src` diperlukan oleh SvelteKit. Untuk production yang lebih ketat, terapkan nonce/hash.

## Helmet Middleware

Fiber `helmet.New()` TIDAK digunakan (lihat `backend/main.go`).

**Alasan:** `helmet.New()` default mengirim:
- `Cross-Origin-Embedder-Policy: require-corp`
- `Cross-Origin-Resource-Policy: same-origin`

Kedua header ini memblokir Chrome dari memuat gambar upload (`ERR_BLOCKED_BY_RESPONSE.NotSameOrigin`) meskipun dari origin yang sama.

`SecurityHeadersMiddleware` mencakup semua header kritis yang disediakan oleh helmet, tanpa efek samping tersebut.

## File Upload Validation

Semua upload file divalidasi oleh `FileUploadValidator` middleware:

- **Max size:** 5 MB (dapat dikonfigurasi di `DefaultSecurityConfig()`)
- **Allowed MIME types:** `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `application/pdf`, `.xlsx`, `.xls`, `.csv`, `.doc`, `.docx`
- **Allowed extensions:** `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.pdf`, `.xlsx`, `.xls`, `.csv`, `.doc`, `.docx`
- File dengan ekstensi/MIME tidak sesuai akan ditolak dengan HTTP 400

## HSTS (HTTP Strict Transport Security)

Hanya aktif saat koneksi HTTPS:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

## Best Practice untuk Production

1. **CSP:** Ganti `http://localhost:*` dengan domain aktual. Implementasikan nonce/hash untuk `script-src`.
2. **HSTS:** Pastikan HTTPS diaktifkan. HSTS otomatis aktif untuk koneksi HTTPS.
3. **CORS:** Batasi `Access-Control-Allow-Origin` ke domain frontend saja (bukan echo origin).
4. **Rate Limiting:** Sesuaikan tier rate limit di `middleware/ratelimit.go`.
5. **JWT:** Set `JWT_SECRET` di environment (jangan pakai default).
6. **Enkripsi:** Set `ENCRYPTION_KEY` di environment untuk enkripsi data sensitif.
