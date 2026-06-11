# Product Requirements Document (PRD)
## Sistem Informasi Manajemen Sumber Daya Manusia (HRMS)

**Versi Dokumen:** 1.0
**Tanggal:** 10 Juni 2026
**Status:** Draft

---

## Daftar Isi

1. [Executive Summary](#1-executive-summary)
2. [Tujuan & Ruang Lingkup](#2-tujuan--ruang-lingkup)
3. [Stakeholder & Target Pengguna](#3-stakeholder--target-pengguna)
4. [Arsitektur Sistem & Tech Stack](#4-arsitektur-sistem--tech-stack)
5. [Fitur Fungsional Detail](#5-fitur-fungsional-detail)
   - 5.1 [Manajemen Karyawan](#51-manajemen-karyawan)
   - 5.2 [Absensi & Kehadiran](#52-absensi--kehadiran)
   - 5.3 [Penggajian (Payroll)](#53-penggajian-payroll)
   - 5.4 [Manajemen Cuti](#54-manajemen-cuti)
   - 5.5 [Reimbursement](#55-reimbursement)
   - 5.6 [Lembur (Overtime)](#56-lembur-overtime)
   - 5.7 [Pinjaman Karyawan](#57-pinjaman-karyawan)
   - 5.8 [Reprimand (Teguran)](#58-reprimand-teguran)
   - 5.9 [KPI & Performance Review](#59-kpi--performance-review)
   - 5.10 [Payslip / Slip Gaji](#510-payslip--slip-gaji)
   - 5.11 [Pengumuman (Announcements)](#511-pengumuman-announcements)
   - 5.12 [Kalender Hari Libur](#512-kalender-hari-libur)
   - 5.13 [Dokumen Karyawan](#513-dokumen-karyawan)
   - 5.14 [Manajemen Schedule Kerja](#514-manajemen-schedule-kerja)
   - 5.14.4 [Request Shift Change](#5144-request-shift-change-tukar--ganti-shift)
   - 5.15 [Fitur Tambahan untuk Konteks Indonesia](#515-fitur-tambahan-untuk-konteks-indonesia)
6. [Mobile Strategy: PWA vs Native](#6-mobile-strategy-pwa-vs-native)
7. [Non-Functional Requirements](#7-non-functional-requirements)
8. [User Roles & Permissions](#8-user-roles--permissions)
9. [UI/UX Guidelines](#9-uiux-guidelines)
10. [Security & Data Privacy](#10-security--data-privacy)
11. [Data Model Overview](#11-data-model-overview)
12. [Integrasi Pihak Ketiga](#12-integrasi-pihak-ketiga)
13. [Roadmap & Milestone](#13-roadmap--milestone)
14. [Success Metrics](#14-success-metrics)

---

## 1. Executive Summary

**HRMS (Human Resource Management System)** adalah aplikasi manajemen sumber daya manusia yang komprehensif, dirancang untuk memenuhi kebutuhan perusahaan di Indonesia. Sistem ini mencakup seluruh siklus hidup karyawan mulai dari pencatatan data, absensi dengan face detection, penggajian yang sesuai regulasi Indonesia (PPh 21, BPJS, THR), hingga manajemen kinerja.

Aplikasi dikembangkan sebagai **Single Page Application (SPA) berbasis web** dengan dukungan **Progressive Web App (PWA)** untuk akses mobile, serta **backend berkinerja tinggi** menggunakan Go dan PostgreSQL dengan enkripsi data sensitif.

**Nilai Bisnis Utama:**
- Digitalisasi seluruh proses HR yang masih manual
- Kepatuhan terhadap regulasi ketenagakerjaan Indonesia
- Efisiensi operasional HR hingga 70%
- Transparansi data bagi karyawan dan manajemen
- Pengambilan keputusan berbasis data melalui KPI & analytics

---

## 2. Tujuan & Ruang Lingkup

### 2.1 Tujuan

1. Menyediakan sistem terpusat untuk manajemen data karyawan yang aman dan terenkripsi
2. Mengotomatiskan proses penggajian termasuk perhitungan pajak, BPJS, dan THR sesuai regulasi Indonesia
3. Memfasilitasi absensi mobile dengan verifikasi biometrik (face detection) dan geolokasi
4. Menyediakan self-service portal bagi karyawan untuk mengajukan cuti, reimbursement, lembur, dan pinjaman
5. Mendukung approval workflow bertingkat untuk berbagai transaksi HR
6. Menyediakan sistem manajemen kinerja berbasis KPI
7. Menyimpan dan mengelola dokumen karyawan secara digital dan aman

### 2.2 Ruang Lingkup

**In Scope:**
- Master data karyawan dengan data sensitif terenkripsi
- Sistem absensi mobile (face detection + koordinat GPS)
- Modul penggajian lengkap (gaji pokok, upah harian, tunjangan, pajak, BPJS, THR)
- Manajemen cuti (multi-jenis) dengan approval bertingkat
- Reimbursement dengan approval bertingkat
- Overtime request dengan approval bertingkat
- Pinjaman karyawan (potong gaji/manual)
- KPI & Performance Review
- Reprimand / teguran
- Pengumuman perusahaan
- Payslip digital
- Kalender hari libur perusahaan
- Dokumen karyawan
- Schedule kerja (5 hari / 6 hari kerja)
- Manajemen organisasi (departemen, jabatan, level)
- Report & analytics

**Out of Scope (Fase 1):**
- Rekrutmen & onboarding
- Learning Management System (LMS)
- Manajemen aset perusahaan
- Payroll outsourcing / third-party payroll integration
- Multi-tenant

---

## 3. Stakeholder & Target Pengguna

### 3.1 Target Pengguna

| Peran | Deskripsi | Tingkat Akses |
|-------|-----------|---------------|
| **Super Admin** | Administrator sistem dengan akses penuh | Semua modul & konfigurasi |
| **HR Manager** | Manajer HR yang mengelola seluruh operasional HR | Semua fitur HR kecuali konfigurasi sistem |
| **HR Staff** | Staf HR yang menjalankan operasional harian | Input & edit data, proses payroll, approval level 1 |
| **Finance / Payroll** | Tim finance yang mengelola penggajian | Payroll, pinjaman, reimbursement finance |
| **Manager / Atasan** | Manajer yang menyetujui permintaan bawahan | Approval, lihat data tim, KPI tim |
| **Karyawan** | Karyawan reguler — setiap karyawan adalah user sistem dengan login & password sendiri | Self-service: cuti, reimbursement, payslip, absensi mobile, profile, dokumen pribadi |
| **Direktur** | Pimpinan perusahaan | Dashboard, laporan, approval level akhir |

### 3.2 Jumlah Pengguna (Estimasi)

- **Fase 1:** 50–500 karyawan
- **Fase 2:** 500–2000 karyawan
- **Skalabilitas:** Mendukung hingga 10.000+ karyawan

---

## 4. Arsitektur Sistem & Tech Stack

### 4.1 Arsitektur High-Level

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                          │
│  ┌─────────────────────┐  ┌────────────────────────┐    │
│  │  Web App (SvelteKit) │  │  Mobile (PWA)          │    │
│  │  - SPA Mode          │  │  - Same SvelteKit App   │    │
│  │  - PWA Service Worker│  │  - Responsive UI        │    │
│  │  - AG Grid (Table)   │  │  - Card-based Layout    │    │
│  │  - Flowbite UI       │  │  - Face Detection       │    │
│  └──────────┬───────────┘  └──────────┬─────────────┘    │
└─────────────┼──────────────────────────┼────────────────┘
              │        HTTPS/REST        │
              ▼                          ▼
┌─────────────────────────────────────────────────────────┐
│                     API GATEWAY                          │
│              Go Fiber Router + Middleware                 │
│     Auth JWT | Rate Limiter | CORS | Logger | Encrypt    │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼────────────────────────────────┐
│                   SERVICE LAYER                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │ Employee │ │Payroll   │ │Attendance│ │Leave     │    │
│  │ Service  │ │Service   │ │Service   │ │Service   │    │
│  ├──────────┤ ├──────────┤ ├──────────┤ ├──────────┤    │
│  │Reimburse │ │Overtime  │ │Loan      │ │KPI       │    │
│  │ Service  │ │Service   │ │Service   │ │Service   │    │
│  ├──────────┤ ├──────────┤ ├──────────┤ ├──────────┤    │
│  │Announce- │ │Reprimand │ │Doc       │ │Auth      │    │
│  │ment      │ │Service   │ │Service   │ │Service   │    │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼────────────────────────────────┐
│                    DATA LAYER                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │ PostgreSQL  │  │  Redis       │  │  S3/MinIO   │      │
│  │  (sqlc)     │  │  (Cache)     │  │  (Storage)  │      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                          │
│  Enkripsi: AES-256-GCM untuk data sensitif di DB         │
└──────────────────────────────────────────────────────────┘
```

### 4.2 Tech Stack Detail

#### Frontend (Web & Mobile PWA)

| Teknologi | Kegunaan |
|-----------|----------|
| **SvelteKit** | Framework SPA & PWA dengan SSR opsional |
| **TypeScript** | Type safety untuk seluruh kode frontend |
| **Flowbite (Tailwind CSS)** | Component library untuk UI konsisten |
| **AG Grid** | Data grid enterprise untuk tabel kompleks |
| **SvelteKit-PWA (vite-plugin-pwa)** | Service Worker, manifest, offline caching |
| **Face-api.js / TensorFlow.js** | Face detection untuk absensi mobile |
| **Geolocation API** | Mengambil koordinat GPS untuk absensi |
| **Chart.js / D3.js** | Visualisasi data dashboard & laporan |
| **Zustand / Svelte Stores** | State management |
| **React Hook Form / Svelte Forms** | Validasi form |

#### Backend

| Teknologi | Kegunaan |
|-----------|----------|
| **Go (Golang)** | Bahasa pemrograman backend |
| **Fiber v2** | Web framework cepat mirip Express.js |
| **sqlc** | Type-safe code generator dari SQL |
| **PostgreSQL 16+** | Database relasional utama |
| **Redis** | Caching, session store, rate limiter |
| **MinIO / S3 Compatible** | Object storage untuk dokumen |
| **JWT (golang-jwt)** | Authentication & authorization |
| **AES-256-GCM** | Enkripsi data sensitif di database |
| **Viper** | Configuration management |
| **Zap / Logrus** | Structured logging |
| **Prometheus** | Metrics & monitoring |
| **Goose / golang-migrate** | Database migration |

#### Infrastruktur

| Teknologi | Kegunaan |
|-----------|----------|
| **Docker** | Containerization |
| **Docker Compose** | Orchestrasi lokal/dev |
| **Nginx / Caddy** | Reverse proxy |
| **GitHub Actions** | CI/CD pipeline |
| **Vercel / Cloudflare Pages** | Hosting frontend (opsional) |
| **VPS / Cloud VM** | Hosting backend & database |

---

## 5. Fitur Fungsional Detail

### 5.1 Manajemen Karyawan

Fitur inti untuk mengelola seluruh data karyawan dengan dukungan enkripsi data sensitif.

#### 5.1.1 Data Karyawan Lengkap

Setiap **karyawan adalah user aplikasi** — memiliki akun login dengan password dan role tertentu untuk mengakses sistem.

**Informasi Akun (untuk login & autentikasi):**
- **Email** — digunakan sebagai username/login ID (wajib, unique)
- **Password** — di-hash (bcrypt/argon2), tidak pernah disimpan plaintext
- **Role** — menentukan hak akses (Super Admin, HR Manager, HR Staff, Finance, Manager, Karyawan, Direktur)
- **Last Login** — timestamp terakhir login
- **Account Status** — aktif / terkunci (lockout setelah 5x gagal login)
- **Session Management** — JWT access token + refresh token

**Informasi Pribadi (terenkripsi):**
- Nama lengkap
- NIK (Nomor Induk Kependudukan) — **terenkripsi**
- NPWP — **terenkripsi**
- Tempat & tanggal lahir
- Jenis kelamin
- Agama
- Status pernikahan
- Jumlah tanggungan (KK, TK, dll)
- Alamat lengkap (KTP & domisili)
- No. telepon
- Email pribadi
- Golongan darah

**Informasi Pekerjaan:**
- NIP/NIK perusahaan (Nomor Induk Pegawai)
- Tanggal bergabung (join date)
- Status karyawan (Tetap, Kontrak, Percobaan, Harian)
- Tanggal berakhir kontrak (jika kontrak)
- Departemen / Divisi
- Jabatan / Posisi
- Level jabatan (Staff, Senior, Supervisor, Manager, General Manager, Direktur)
- Approval Line / penanggung jawab approval (bisa atasan langsung, admin departemen, atau siapa pun yang ditunjuk)
- Lokasi kerja / kantor cabang
- Golongan / Grade / Pangkat

**Informasi Keuangan (terenkripsi):**
- Nomor rekening bank — **terenkripsi**
- Nama bank
- NPWP — **terenkripsi**
- Status PTKP (K/0, K/1, K/2, K/3, TK/0, TK/1, TK/2, TK/3)
- Tarif PPh 21 (TER)
- **Komponen Gaji** — daftar komponen pendapatan (tunjangan) dan potongan yang berlaku untuk karyawan tersebut (disimpan di tabel `employee_salary_components`)
- **Konfigurasi BPJS** — override per karyawan (enable/disable per komponen, custom rate) — disimpan di kolom `bpjs_config` (JSONB)

**Informasi Kontak Darurat:**
- Nama kontak darurat
- Hubungan
- No. telepon
- Alamat

#### 5.1.2 Fitur CRUD Karyawan

- **Create:** Tambah karyawan baru dengan validasi data
- **Read:** Detail karyawan dengan akses role-based (data sensitif hanya untuk HR & Finance)
- **Update:** Edit data karyawan dengan audit trail
- **Delete:** Soft-delete karyawan (non-aktifkan)
- **Import/Export:** Import dari Excel/CSV, Export ke Excel/PDF
- **History:** Riwayat perubahan jabatan, departemen, status, gaji

#### 5.1.3 Struktur Organisasi

- Visualisasi struktur organisasi (org chart)
- Manajemen departemen (tambah/edit/hapus departemen)
- Manajemen jabatan
- Manajemen level / grade

#### 5.1.4 Mutasi & Promosi

- Catatan riwayat mutasi/promosi
- Approval untuk mutasi/promosi
- Efek otomatis ke perubahan gaji dan tunjangan

---

### 5.2 Absensi & Kehadiran

#### 5.2.1 Schedule Kerja

- **Tipe Schedule:**
  - 5 hari kerja (Senin–Jumat) — 8 jam/hari
  - 6 hari kerja (Senin–Sabtu) — 7 jam/hari
- **Shift:** Support multiple shift (pagi, siang, malam)
- **Flexible time:** Konfigurasi jam masuk, jam pulang, toleransi keterlambatan
- **Assign schedule:** Per departemen, per jabatan, atau per individu

#### 5.2.2 Absensi Mobile (Fitur Utama)

**Metode Absensi:**
1. **Face Detection** — Verifikasi wajah menggunakan kamera smartphone
   - Menggunakan face-api.js dengan model wajah yang telah terdaftar
   - Anti-spoofing detection (deteksi foto/video)
   - Pencocokan dengan foto wajah yang tersimpan di database
2. **Geolokasi (GPS)** — Verifikasi koordinat absensi
   - Koordinat absensi bisa lebih dari satu (multiple office locations)
   - Radius toleransi yang bisa dikonfigurasi (default: 100 meter)
   - Beberapa karyawan bisa dikecualikan dari pengecekan koordinat (remote workers, sales)

**Alur Absensi:**
```
1. Karyawan membuka halaman absensi mobile
2. Sistem meminta izin kamera & GPS
3. Karyawan mengambil foto selfie
4. Sistem melakukan face detection & matching
5. Sistem mengambil koordinat GPS & memvalidasi dengan koordinat absensi
6. Absensi berhasil dicatat dengan timestamp, foto, koordinat, dan device info
```

**Tipe Absensi:**
- Check-in (masuk)
- Check-out (pulang)
- Istirahat (break in/out)
- Lembur (overtime in/out)
- Izin tidak masuk (melalui menu cuti/izin)

**Pengecualian Absensi Mobile:**
- Karyawan dengan status "Remote" / "Sales" boleh absensi tanpa koordinat GPS
- Karyawan yang lupa absen bisa mengajukan "Manual Attendance Request" melalui Approval Line-nya

#### 5.2.3 Dashboard Absensi

- **Untuk Karyawan:** Riwayat absensi pribadi, total jam kerja, keterlambatan
- **Untuk HR/Manager:**
  - Rekap absensi per departemen/periode
  - Laporan karyawan tidak hadir
  - Grafik tingkat kehadiran
  - Export laporan ke Excel/PDF

#### 5.2.4 Manual Attendance / Pengajuan Absensi

Fitur ini memungkinkan karyawan yang **lupa melakukan absensi** (check-in atau check-out) pada hari tertentu untuk mengajukan pencatatan kehadiran secara manual dengan **persetujuan Approval Line** (penanggung jawab approval karyawan tersebut).

##### Skenario Pengajuan

| Skenario | Contoh |
|----------|--------|
| **Lupa Check-in** | Karyawan hadir bekerja tapi lupa absen masuk karena terburu-buru |
| **Lupa Check-out** | Karyawan pulang tapi lupa absen pulang |
| **Lupa Keduanya** | Karyawan lupa absen masuk & pulang (full-day manual) |
| **Error Teknis** | Aplikasi error, kamera tidak berfungsi, GPS tidak akurat |
| **Terlambat Absen** | Karyawan absen melebihi batas toleransi waktu yang ditentukan |

##### Alur Pengajuan & Approval

```
Karyawan
  │  Mengajukan manual attendance
  │  - Memilih tanggal yang akan diisi
  │  - Mengisi jam check-in dan/atau check-out yang sebenarnya
  │  - Menyertakan alasan/penjelasan
  ▼
Approval Line (Penanggung Jawab Approval)
  │  Menerima notifikasi pengajuan
  │  Melihat riwayat absensi karyawan (apakah sudah ada check-in/out parsial)
  │  Approve / Reject
  ▼
  ├── ✅ Approved → Data otomatis masuk ke attendance_records sebagai manual entry
  │                  Status record: 'hadir' atau 'terlambat' (sesuai jam yang diajukan)
  │                  is_manual_entry = TRUE
  │                  manual_entry_approved_by = approval_line
  │
  └── ❌ Rejected → Status berubah menjadi 'rejected'
                     Karyawan mendapat notifikasi + alasan penolakan
```

##### Detail Field Pengajuan

| Field | Tipe | Keterangan |
|-------|------|------------|
| **Tanggal** | Date | Tanggal absensi yang ingin diperbaiki |
| **Jam Check-in** | Time | Jam masuk yang sebenarnya (opsional, jika lupa check-in) |
| **Jam Check-out** | Time | Jam pulang yang sebenarnya (opsional, jika lupa check-out) |
| **Alasan** | Text | Penjelasan mengapa lupa absen / error teknis |

##### Aturan Bisnis

| Aturan | Detail |
|--------|--------|
| **Batas Pengajuan** | Maksimal 3x per bulan per karyawan (dapat dikonfigurasi di company settings) |
| **Batas Waktu Pengajuan** | Pengajuan hanya bisa dilakukan H+1 sampai H+3 setelah tanggal absensi yang terlewat (dapat dikonfigurasi) |
| **Batas Waktu Approval** | Approval Line harus merespon dalam 2×24 jam; jika melewati batas, otomatis eskalasi ke HR Manager |
| **Duplikasi** | Sistem mencegah pengajuan ganda untuk tanggal yang sama — jika sudah ada pengajuan pending, tidak bisa buat baru |
| **Konflik dengan Absensi** | Jika sudah ada data absensi (check-in/out via mobile) di tanggal tersebut, pengajuan manual hanya untuk melengkapi data yang kurang (misal: sudah check-in via face detection, tapi lupa check-out) |
| **Dampak Payroll** | Data kehadiran hasil manual attendance tetap diproses dalam perhitungan payroll (potongan ketidakhadiran, tunjangan makan, lembur) |
| **Audit Trail** | Semua pengajuan dan approval tercatat di activity_logs (siapa mengajukan, siapa menyetujui, kapan) |
| **Pembatalan** | Karyawan dapat membatalkan pengajuan yang masih berstatus "pending" |

##### Tampilan di Mobile (PWA)

**Untuk Karyawan:**
- Menu **"Ajukan Absensi Manual"** di halaman Absensi
- Form input: pilih tanggal, isi jam check-in/out, tulis alasan
- Riwayat pengajuan: list semua pengajuan dengan status (pending/approved/rejected)
- Status menunggu approval: ditandai dengan ikon jam/warning

**Untuk Approval Line:**
- **Tab "Pending Approval"** di dashboard — menampilkan semua pengajuan yang menunggu approval
- Detail: nama karyawan, tanggal, jam yang diajukan, alasan
- Tombol **Approve** / **Reject** dengan konfirmasi
- Jika Reject: wajib mengisi alasan penolakan
- Riwayat approval: Approval Line bisa melihat histori keputusan sebelumnya

---

### 5.3 Penggajian (Payroll)

Sistem penggajian yang sesuai dengan regulasi ketenagakerjaan Indonesia.

#### 5.3.1 Komponen Gaji

| Komponen | Deskripsi | Sifat |
|----------|-----------|-------|
| **Gaji Pokok** | Gaji tetap bulanan | Tetap |
| **Upah Harian** | Untuk karyawan harian/borongan | Variabel |
| **Tunjangan Jabatan** | Tunjangan berdasarkan posisi | Tetap |
| **Tunjangan Transport** | Uang transport | Tetap/Variabel |
| **Tunjangan Makan** | Uang makan harian | Variabel |
| **Tunjangan Komunikasi** | Pulsa/internet | Tetap |
| **Tunjangan Keluarga** | Tunjangan istri/anak | Tetap |
| **Tunjangan Kesehatan** | Tambahan di luar BPJS | Tetap |
| **Tunjangan Pendidikan** | Untuk karyawan yang studi | Variabel |
| **Insentif / Bonus** | Bonus bulanan/kinerja | Variabel |
| **Uang Lembur** | Perhitungan lembur 1.5x - 3x | Variabel |
| **THR** | Tunjangan Hari Raya | Tahunan |

#### 5.3.2 Potongan Gaji

| Potongan | Deskripsi |
|----------|-----------|
| **PPh 21** | Pajak penghasilan (Tarif Efektif Rata-rata / TER) |
| **BPJS Kesehatan** | 1% dari gaji (pekerja) + 4% (perusahaan) — **Dapat dikonfigurasi** |
| **BPJS JHT** | 2% dari gaji (pekerja) + 3.7% (perusahaan) — **Dapat dikonfigurasi** |
| **BPJS JP (Pensiun)** | 1% dari gaji (pekerja) + 2% (perusahaan) — **Dapat dikonfigurasi** |
| **BPJS JKK** | 0.24%-1.74% (perusahaan) — **Dapat dikonfigurasi** |
| **BPJS JKM** | 0.3% (perusahaan) — **Dapat dikonfigurasi** |
| **Pinjaman Karyawan** | Cicilan pinjaman |
| **Kasbon / Uang Muka** | Potongan kasbon |
| **Denda Keterlambatan** | Sanksi keterlambatan |
| **Kekurangan Absen** | Potongan ketidakhadiran |

#### 5.3.3 Perhitungan Pajak (PPh 21 TER)

Perhitungan menggunakan **Tarif Efektif Rata-rata (TER)** berdasarkan PP No. 58 Tahun 2023:

1. Tentukan kategori TER (A, B, atau C) berdasarkan status PTKP
2. Hitung penghasilan bruto bulanan
3. Kurangi dengan biaya jabatan (5% dari bruto, max Rp500.000/bulan)
4. Kurangi dengan iuran BPJS (hanya JHT & JP pekerja)
5. Terapkan tarif TER untuk Januari–November
6. Bulan Desember: hitung ulang dengan tarif progresif Pasal 17 untuk selisih

#### 5.3.4 Perhitungan BPJS

> **✅ Semua parameter BPJS sudah dapat dikonfigurasi** — baik persentase iuran maupun ceiling, per perusahaan maupun per karyawan. Tidak ada nilai yang hardcoded.

##### Default Pemerintah (Sebagai Awal)

| Program | Pekerja | Perusahaan | Ceiling (Gaji Maks) |
|---------|---------|-----------|---------------------|
| BPJS Kesehatan | 1% | 4% | Rp12.000.000/bulan |
| BPJS JHT | 2% | 3.7% | Tidak ada batas |
| BPJS JP (Pensiun) | 1% | 2% | Rp10.000.000/bulan |
| BPJS JKK | — | 0.24%–1.74% (risk-based) | Tidak ada batas |
| BPJS JKM | — | 0.3% | Tidak ada batas |

##### Tingkat Fleksibilitas

Sistem mendukung **2 level konfigurasi**:

**Level 1 — Default Perusahaan (`companies.hr_settings`)**

Semua rate dan ceiling bisa diubah di setting perusahaan. HR Admin tinggal edit JSON di halaman pengaturan:

```json
{
  "bpjs": {
    "kesehatan": {
      "employee_rate": 0.01,    // 1% pekerja
      "company_rate": 0.04,     // 4% perusahaan
      "ceiling": 12000000       // Rp12.000.000
    },
    "jht": {
      "employee_rate": 0.02,    // 2% pekerja
      "company_rate": 0.037,    // 3.7% perusahaan
      "ceiling": null            // Tidak ada batas
    },
    "jp": {
      "employee_rate": 0.01,    // 1% pekerja
      "company_rate": 0.02,     // 2% perusahaan
      "ceiling": 10000000       // Rp10.000.000
    },
    "jkm": {
      "company_rate": 0.003,   // 0.3% perusahaan
      "ceiling": null
    }
  }
}
```

**Level 2 — Per Karyawan (`employees.bpjs_config`)**

Setiap karyawan bisa memiliki konfigurasi BPJS sendiri:

```json
{
  "kesehatan": {"enabled": false},             // Nonaktifkan BPJS Kesehatan (misal: sudah punya mandiri)
  "jht": {"employee_rate": 0.01},              // Rate JHT pekerja 1% (bukan 2%)
  "jp": {"company_rate": 0.015},               // Rate JP perusahaan 1.5% (bukan 2%)
  "jkm": {"enabled": false}                    // Nonaktifkan JKM
}
```

##### Hierarki Resolusi

```
Ada override per employee (bpjs_config)?
  ├── Komponen enabled=false? → Hitung 0 untuk komponen itu
  ├── Ada override rate? → Pakai rate dari bpjs_config
  │
  └── Tidak ada override → Pakai default dari perusahaan
       ├── Ada setting di hr_settings.bpjs? → Pakai rate & ceiling dari situ
       └── Tidak ada setting → Pakai nilai default pemerintah
```

##### Contoh Skenario

| Karyawan | BPJS Kesehatan | BPJS JHT | BPJS JP | BPJS JKM |
|----------|---------------|----------|---------|----------|
| **Budi** (Staff) | ✅ 1% (default) | ✅ 2% (default) | ✅ 1% (default) | ✅ 0.3% |
| **Siti** (Sudah punya BPJS mandiri) | ❌ **Disabled** | ✅ 2% | ✅ 1% | ✅ 0.3% |
| **Rudi** (Expat, negosiasi khusus) | ❌ **Disabled** | ❌ **Disabled** | ❌ **Disabled** | ❌ **Disabled** |
| **Dewi** (Direktur, rate khusus) | ✅ 0% pekerja, 4% perusahaan | ✅ **1%** (override) | ✅ 1% | ✅ 0.3% |

##### Update Tanpa Deploy

Karena semua parameter BPJS disimpan di `companies.hr_settings` (bukan hardcoded di kode):

```
Perubahan rate BPJS oleh pemerintah
  → HR Admin edit JSON di halaman Company Settings
  → Simpan
  → Payroll bulan berikutnya otomatis pakai rate baru
  → ✅ Tidak perlu deploy ulang aplikasi
```

> **Referensi:** Perpres BPJS Kesehatan, PP BPJS Ketenagakerjaan — perbarui parameter di Company Settings setiap ada perubahan regulasi.
> Jika perusahaan ingin menerapkan kebijakan sendiri (tidak sesuai pemerintah), cukup ubah nilai di settings.

#### 5.3.5 Perhitungan THR

- **Karyawan > 12 bulan:** 1 bulan gaji penuh (gaji pokok + tunjangan tetap)
- **Karyawan 1–12 bulan:** Proporsional (masa kerja/12 × 1 bulan gaji)
- **THR dikenakan PPh 21** (digabung dengan gaji bulan berjalan, dihitung menggunakan TER)
- **THR tidak dikenakan BPJS** — potongan BPJS di tabel 5.3.2 hanya berlaku untuk gaji reguler bulanan, bukan untuk THR

#### 5.3.6 Fitur Payroll Lainnya

- **Payroll Period:** Monthly (1–30/31)
- **Prorate Gaji:** Untuk karyawan baru/keluar di tengah bulan
- **Pay Slip (Slip Gaji):** Generated otomatis setiap bulan
- **History:** Riwayat perubahan gaji per karyawan
- **Approval Flow:** Payroll perlu di-approve oleh HR Manager dan Finance
- **Disbursement Summary:** Rekap total penggajian untuk transfer bank
- **Export:** Export payroll report ke Excel, CSV, atau format bank

#### 5.3.7 Master Data Komponen Gaji per Karyawan

Setiap karyawan bisa memiliki **komponen gaji yang berbeda-beda** sesuai kebijakan perusahaan. Sistem menyimpan master data komponen gaji per karyawan agar HR tidak perlu input manual setiap bulan.

##### Tabel `employee_salary_components`

| Field | Tipe | Keterangan |
|-------|------|------------|
| **employee_id** | UUID (FK) | Karyawan yang memiliki komponen ini |
| **component_name** | VARCHAR | Nama komponen (Tunj. Jabatan, Tunj. Transport, Tunj. Makan, dll) |
| **component_type** | ENUM | `allowance` (pendapatan) atau `deduction` (potongan) |
| **amount** | DECIMAL | Nilai nominal komponen |
| **is_active** | BOOLEAN | Aktif/non-aktif (bisa dinonaktifkan tanpa hapus data) |
| **effective_date** | DATE | Tanggal mulai berlaku |

**Contoh data:**

| Karyawan | Komponen | Tipe | Amount |
|----------|----------|------|-------|
| Budi (Staff) | Tunj. Jabatan | allowance | Rp500.000 |
| Budi (Staff) | Tunj. Transport | allowance | Rp300.000 |
| Budi (Staff) | Pinjaman Koperasi | deduction | Rp200.000 |
| Siti (Supervisor) | Tunj. Jabatan | allowance | Rp1.000.000 |
| Siti (Supervisor) | Tunj. Transport | allowance | Rp500.000 |
| Siti (Supervisor) | Tunj. Komunikasi | allowance | Rp200.000 |
| Siti (Supervisor) | BPJS Tambahan | deduction | Rp150.000 |

##### Alur Penggunaan

```
HR mengatur komponen gaji per karyawan (via form master data)
  │  Setiap karyawan memiliki daftar komponen sendiri
  ▼
Saat membuat payroll bulanan:
  │  Sistem membaca employee_salary_components yang aktif
  │  untuk setiap karyawan
  ▼
  ├── Komponen allowance → masuk ke allowances JSONB di payroll_items
  └── Komponen deduction → masuk ke deductions JSONB di payroll_items
```

##### Riwayat Perubahan Komponen (`employee_salary_component_histories`)

Setiap perubahan pada komponen gaji (tambah, ubah nilai, non-aktifkan) tercatat otomatis:

| Field | Keterangan |
|-------|------------|
| **employee_id** | Karyawan |
| **component_name** | Nama komponen yang berubah |
| **old_amount** | Nilai sebelum perubahan |
| **new_amount** | Nilai setelah perubahan |
| **change_reason** | Alasan perubahan (promosi, kenaikan tahunan, dll) |
| **changed_by** | Siapa yang mengubah |
| **changed_at** | Kapan perubahan dilakukan |

**Contoh riwayat:**
| Tanggal | Karyawan | Komponen | Lama | Baru | Alasan |
|---------|----------|----------|------|------|-------|
| 01-01-2026 | Budi | Tunj. Jabatan | Rp300.000 | Rp500.000 | Promosi Staff |
| 01-07-2026 | Budi | Tunj. Transport | Rp300.000 | Rp400.000 | Kenaikan tahunan |

---

### 5.4 Manajemen Cuti

#### 5.4.1 Jenis Cuti

| Jenis Cuti | Durasi Default | Periode | Catatan |
|------------|---------------|---------|---------|
| **Cuti Tahunan** | 12 hari/tahun | Jan–Des | Untuk karyawan tetap |
| **Cuti Sakit** | 14 hari/tahun | Jan–Des | Dengan/dokter tanpa surat |
| **Cuti Hamil/Melahirkan** | 90 hari | Per kejadian | 1.5 bulan sebelum & sesudah |
| **Cuti Keguguran** | 30 hari | Per kejadian | Dengan surat dokter |
| **Cuti Menikah** | 3 hari | Per kejadian | Untuk diri sendiri |
| **Cuti Menikahkan Anak** | 2 hari | Per kejadian | |
| **Cuti Khitanan/Baptis** | 2 hari | Per kejadian | Untuk anak |
| **Cuti Keluarga Meninggal** | 2 hari | Per kejadian | Suami/istri/anak/orangtua/mertua |
| **Cuti Ibadah Haji** | 40 hari | Per kejadian | Sekali selama bekerja |
| **Cuti Alasan Penting** | Variabel | Per kejadian | Persetujuan manajemen |
| **Cuti Bersama** | Libur nasional | Tahunan | Ditetapkan perusahaan |

#### 5.4.2 Alur Approval Cuti

```
Karyawan → Approval Line → HR Manager → (Jika > 3 hari) → Direktur
```

**Ketentuan:**
- Approval bertingkat berdasarkan durasi cuti
- Yang bisa di-cut: karyawan tetap & kontrak
- Cuti sakit dengan surat dokter: bypass approval queue
- Kuota cuti: reset setiap awal tahun
- Sisa cuti tahunan bisa di-rollover (max 6 hari) atau di-cashout

#### 5.4.3 Fitur Tambahan Cuti

- **Cuti Bersama:** Libur nasional bersama yang ditetapkan pemerintah
- **Sisa Cuti:** Dashboard sisa cuti per jenis
- **History Cuti:** Riwayat pengajuan dan persetujuan
- **Cancel Cuti:** Pembatalan cuti yang sudah di-approve
- **Export:** Laporan cuti per departemen/periode

---

### 5.5 Reimbursement

#### 5.5.1 Jenis Reimbursement

| Jenis | Deskripsi | Max (Konfigurabel) |
|-------|-----------|-------------------|
| **Biaya Medis** | Penggantian biaya berobat | Rp2.000.000/tahun |
| **Biaya Perjalanan Dinas** | Tiket, hotel, transportasi | Per kejadian |
| **Biaya Pelatihan** | Kursus, seminar, sertifikasi | Per kejadian |
| **Biaya Bahan Kerja** | Pembelian supplies | Per bulan |
| **Biaya Lain-lain** | Kategori umum | Per kejadian |

#### 5.5.2 Alur Approval Reimbursement

```
Karyawan → Approval Line → HR Manager → Finance → (Jika > limit) → Direktur
```

**Ketentuan:**
- Lampiran bukti (foto/faktur/invoice) — upload wajib
- Approval bertingkat berdasarkan nominal
- Batas maksimum per kategori
- Pembayaran via payroll (bulanan) atau langsung (transfer manual)

---

### 5.6 Lembur (Overtime)

#### 5.6.1 Perhitungan Lembur

Sesuai UU Ketenagakerjaan Indonesia:

| Jam Lembur | Hari Kerja | Hari Libur |
|------------|-----------|------------|
| Jam ke-1 | 1.5x upah sejam | 2x upah sejam |
| Jam ke-2+ | 2x upah sejam | 3x upah sejam |
| Jam ke-8+ | — | 4x upah sejam |

**Rumus upah sejam:** 1/173 × gaji pokok bulanan

#### 5.6.2 Alur Approval Lembur

```
Karyawan → Approval Line → HR Manager
```

**Ketentuan:**
- Maksimal 3 jam/hari atau 14 jam/minggu
- Wajib lembur atau sukarela
- Pengajuan minimal H-1
- Check-in/out lembur via mobile

#### 5.6.3 Konfigurasi Rate Lembur (Fleksibel)

Sistem mendukung **3 level konfigurasi** untuk perhitungan rate lembur, dengan hierarki resolusi:

```
Ada override per employee? → Pakai employee_overtime_rates
  ↓ Tidak
Ada override per position_grade? → Pakai position_grade_overtime_rates
  ↓ Tidak
Pakai default dari company settings (hr_settings)
```

##### Level 1: Default Perusahaan (`companies.hr_settings`)

Konfigurasi global yang berlaku untuk semua karyawan (default):

```json
{
  "overtime": {
    "hourly_rate_method": "base_salary / 173",
    "default_rates": {
      "weekday": [
        {"hour_from": 1, "hour_to": 1, "multiplier": 1.5},
        {"hour_from": 2, "hour_to": null, "multiplier": 2.0}
      ],
      "weekend": [
        {"hour_from": 1, "hour_to": 7, "multiplier": 2.0},
        {"hour_from": 8, "hour_to": null, "multiplier": 3.0}
      ],
      "holiday": [
        {"hour_from": 1, "hour_to": 7, "multiplier": 2.0},
        {"hour_from": 8, "hour_to": null, "multiplier": 3.0}
      ]
    }
  }
}
```

##### Level 2: Per Position Grade (Override Grup)

Bisa diatur untuk level jabatan tertentu (tabel `position_grade_overtime_rates`):

| Position Grade | Weekday | Weekend |
|---------------|---------|--------|
| Staff (level 1-2) | Jam 1: 1.5x, Jam 2-3: 2x, Jam 4+: 2.5x | Jam 1-7: 2x, Jam 8+: 3x |
| Senior Staff (level 3-4) | Jam 1: 1.5x, Jam 2+: 2x | Jam 1-7: 2x, Jam 8+: 3x |
| Supervisor+ (level 5+) | Tidak ada lembur (all-in gaji) | — |

##### Level 3: Per Employee (Override Individu)

Untuk kebutuhan spesifik seperti contoh Anda — setiap karyawan bisa punya pengali yang berbeda:

**Tabel `employee_overtime_rates`:**

| Field | Tipe | Keterangan |
|-------|------|------------|
| **employee_id** | UUID (FK) | Karyawan |
| **overtime_type** | ENUM | weekday / weekend / holiday |
| **hour_from** | INTEGER | Mulai jam ke- |
| **hour_to** | INTEGER (nullable) | Sampai jam ke- (NULL = seterusnya) |
| **multiplier** | DECIMAL | Pengali (1.5, 2.0, 2.5, dst) |
| **is_active** | BOOLEAN | Aktif/non-aktif |
| **effective_date** | DATE | Tanggal mulai berlaku |

**Contoh data:**

| Karyawan | Hari | Jam | Pengali |
|----------|------|-----|---------|
| Karyawan A (Staff) | Weekday | 1-2 | 1.5x |
| Karyawan A (Staff) | Weekday | 3-4 | 2.0x |
| Karyawan A (Staff) | Weekday | 5+ | 2.5x |
| Karyawan B (Spv) | Weekday | 1-2 | 2.0x |
| Karyawan B (Spv) | Weekday | 3+ | 2.5x |
| Karyawan C (Remote) | Weekday | 1-3 | 1.0x (no overtime) |
| Karyawan D | Weekend | 1-7 | 3.0x (double) |

##### Rumus Perhitungan

```
hourly_rate = base_salary / 173

overtime_pay = SUM(
    untuk setiap segment jam:
        jam_di_segment × hourly_rate × multiplier_segment
)

Contoh: Karyawan A lembur 4 jam di Weekday
  Jam 1-2: 2 jam × (base_salary/173) × 1.5 = RpX
  Jam 3-4: 2 jam × (base_salary/173) × 2.0 = RpY
  Total: RpX + RpY
```

---

### 5.7 Pinjaman Karyawan

#### 5.7.1 Fitur Pinjaman

| Fitur | Deskripsi |
|-------|-----------|
| **Jenis Pinjaman** | Pinjaman reguler, pinjaman darurat, pinjaman pendidikan |
| **Maksimal Pinjaman** | Konfigurabel (default 10x gaji pokok) |
| **Tenor** | 3–24 bulan (konfigurabel) |
| **Bunga** | 0% (tanpa bunga) atau bunga rendah (konfigurabel) |
| **Metode Pembayaran** | Potong gaji otomatis atau bayar manual via transfer |

#### 5.7.2 Alur Approval Pinjaman

```
Karyawan → Approval Line → HR Manager → Finance → Direktur
```

**Ketentuan:**
- Total cicilan pinjaman tidak boleh melebihi 30% dari gaji bulanan
- Sisa pinjaman existing akan tampil saat pengajuan baru
- Potong gaji: otomatis masuk sebagai komponen potongan di payroll
- Bayar manual: karyawan transfer ke rekening perusahaan

---

### 5.8 Reprimand (Teguran)

#### 5.8.1 Jenis Teguran

| Jenis | Tingkat | Dampak |
|-------|---------|--------|
| **Teguran Lisan** | Ringan | Peringatan verbal tercatat |
| **SP1 (Surat Peringatan 1)** | Sedang | Tertulis, masa berlaku 6 bulan |
| **SP2 (Surat Peringatan 2)** | Berat | Tertulis, masa berlaku 6 bulan |
| **SP3 (Surat Peringatan 3)** | Sangat Berat | Dapat berujung PHK |

#### 5.8.2 Alur Reprimand

```
HR / Atasan → Membuat Reprimand → Ditandatangani (digital) → Karyawan → Arsip
```

**Fitur:**
- Template surat peringatan
- Tanda tangan digital / konfirmasi baca
- Riwayat teguran per karyawan
- Notifikasi ke karyawan
- Cetak/export surat peringatan dalam PDF
- Automated escalation: jika dalam 6 bulan ada pelanggaran berulang, naik level

---

### 5.9 KPI & Performance Review

#### 5.9.1 Komponen KPI

| Komponen | Deskripsi | Bobot (Default) |
|----------|-----------|-----------------|
| **KPI Individu** | Target individu per periode | 60% |
| **KPI Tim** | Target tim/departemen | 20% |
| **Kompetensi** | Soft skills, leadership, values | 20% |

#### 5.9.2 Fitur KPI

- **Template KPI:** Per jabatan/departemen
- **Setting KPI:** Di awal periode (tahunan/kuartalan)
- **Penilaian:** Tengah tahun & akhir tahun
- **Skala Penilaian:** 1–5 (Sangat Kurang – Sangat Baik)
- **Rating Weighted:** Total skor × bobot
- **Kategori Hasil:**
  - 4.5–5.0: Outstanding
  - 3.5–4.4: Meets Expectations / Above
  - 2.5–3.4: Meets Expectations
  - 1.5–2.4: Needs Improvement
  - 1.0–1.4: Underperform

#### 5.9.3 Alur Performance Review

```
Setting KPI (HR & Manager) → Self Assessment (Karyawan) → 
Manager Review → HR Review → Calibration → Final Score
```

**Dampak Hasil Review:**
- Kenaikan gaji berkala
- Bonus kinerja
- Promosi
- Improvement Plan (jika perlu)

---

### 5.10 Payslip / Slip Gaji

#### 5.10.1 Fitur Payslip

- **Akses:** Karyawan, HR, Finance
- **Periode:** Bulanan, dapat diakses setelah payroll di-approve
- **Format:** Digital (web/mobile) + Download PDF
- **Detail yang ditampilkan:**
  - Gaji pokok
  - Semua tunjangan
  - Lembur
  - Semua potongan (PPh 21, BPJS, pinjaman, dll)
  - Total take home pay
  - Sisa cuti tahunan

#### 5.10.2 Keamanan Payslip

- **Akses terbatas:** Karyawan hanya bisa melihat payslip sendiri
- **Enkripsi:** Data payslip ditampilkan melalui koneksi HTTPS terenkripsi
- **PIN Protection:** Payslip bisa dilindungi PIN tambahan
- **Audit Log:** Setiap akses ke payslip tercatat

---

### 5.11 Pengumuman (Announcements)

#### 5.11.1 Fitur Pengumuman

- **Pembuat:** HR Admin, Manager, Direktur
- **Target:** Semua karyawan / per departemen / per level
- **Tipe:**
  - Pengumuman biasa
  - Pengumuman penting (dengan popup notification)
  - Pengumuman darurat
- **Lampiran:** Bisa menyertakan file (PDF, gambar)
- **Expired Date:** Pengumuman bisa memiliki masa berlaku
- **Pinned:** Pengumuman penting bisa di-pin di atas

---

### 5.12 Kalender Hari Libur

#### 5.12.1 Fitur Kalender

- **Tipe Hari Libur:**
  - Libur Nasional (ditentukan pemerintah)
  - Libur Bersama (ditetapkan pemerintah)
  - Libur Perusahaan (cuti bersama perusahaan)
  - Event perusahaan (gathering, rapat tahunan)
- **Integrasi:** Sync dengan Google Calendar / Outlook Calendar
- **Visualisasi:** Calendar view, list view
- **Notifikasi:** Pengingat menjelang hari libur
- **Dampak Payroll:** Otomatis mempengaruhi perhitungan kehadiran & lembur

---

### 5.13 Dokumen Karyawan

#### 5.13.1 Jenis Dokumen

| Dokumen | Keterangan |
|---------|------------|
| KTP / KK | Scan dokumen kependudukan |
| Ijazah & Transkrip | Pendidikan terakhir |
| Sertifikat Pelatihan | Pelatihan & sertifikasi |
| Kontrak Kerja | Perjanjian kerja |
| Surat Pengalaman Kerja | Dari perusahaan sebelumnya |
| Foto Karyawan | Foto formal & foto untuk face recognition |
| NPWP | Kartu NPWP |
| BPJS Kesehatan & Ketenagakerjaan | Kartu BPJS |
| Dokumen Pendukung Lainnya | Bebas |

#### 5.13.2 Fitur Manajemen Dokumen

- **Upload:** Karyawan upload dokumen, HR memverifikasi
- **Verifikasi Status:** Pending → Verified → Rejected
- **Expiry Date:** Peringatan dokumen yang akan expired (KTP, kontrak)
- **Storage:** Encrypted at rest di S3/MinIO
- **Akses:** Role-based (dokumen sensitif hanya untuk HR)
- **Download:** Encrypted download dengan audit trail

---

### 5.14 Manajemen Schedule Kerja

#### 5.14.1 Tipe Schedule

- **5 Hari Kerja:** Senin–Jumat (8 jam/hari = 40 jam/minggu)
  - Jam masuk: 08:00 / 09:00 (konfigurabel)
  - Jam pulang: 17:00 / 18:00 (konfigurabel)
  - Istirahat: 12:00–13:00
- **6 Hari Kerja:** Senin–Sabtu (7 jam/hari = 40 jam/minggu)
  - Jam masuk: 08:00
  - Jam pulang: 16:00 (Jumat: 11:30–16:00)
  - Istirahat: 12:00–13:00

#### 5.14.2 Shift (Opsional)

- Shift Pagi: 07:00–15:00
- Shift Siang: 15:00–23:00
- Shift Malam: 23:00–07:00

#### 5.14.3 Assignment

- **Default per departemen:** Setiap departemen memiliki jadwal kerja default (via `departments.work_schedule_id`). Karyawan yang berada di departemen tersebut secara otomatis mengikuti jadwal default departemen.
- **Override per individu:** Karyawan tertentu bisa memiliki jadwal berbeda dari default departemen (via `employees.work_schedule_id`).
- **Contoh penerapan (Hotel):**
  - Departemen **Housekeeping** — jadwal 6 hari kerja (Senin–Sabtu), shift pagi 07:00–15:00
  - Departemen **Engineering** — jadwal 5 hari kerja (Senin–Jumat), shift pagi 08:00–17:00 dengan shift malam
  - Departemen **Front Office** — jadwal shift rotating (pagi, siang, malam)
- **Hierarki resolusi jadwal:** `Individual override (jika ada)` → `Departemen default` → `Global default`
- **Periode schedule:** Bulanan / tetap
- **Rotasi shift:** Bisa dijadwalkan otomatis per departemen

#### 5.14.4 Request Shift Change (Tukar & Ganti Shift)

Fitur yang memungkinkan karyawan untuk mengajukan perubahan shift melalui **aplikasi web (desktop) maupun mobile (PWA)** dengan sistem approval.

##### Jenis Request

| Jenis | Deskripsi | Contoh |
|-------|-----------|--------|
| **Ganti Shift Individu** | Karyawan meminta perubahan shift untuk tanggal tertentu (misal: minta shift pagi instead of malam) | Andi (shift malam) request shift pagi 17 Juni karena ada acara keluarga |
| **Tukar Shift (Swap)** | 2 karyawan saling bertukar jadwal shift di tanggal yang sama atau berbeda | Andi (shift pagi 17 Juni) ingin tukar dengan Budi (shift siang 17 Juni), atau tukar tanggal berbeda |

##### Alur Pengajuan & Approval

**Ganti Shift Individu:**
```
Karyawan
  │  Mengajukan shift change
  │  - Pilih tanggal
  │  - Pilih shift tujuan (pagi/siang/malam)
  │  - Alasan perubahan
  ▼
Approval Line (Penanggung Jawab Approval)
  │  Approve / Reject
  ▼
  ├── ✅ Approved → Jadwal shift karyawan berubah untuk tanggal tersebut
  │                  Tercatat di shift_change_requests (status: approved)
  │                  Notifikasi ke karyawan
  │
  └── ❌ Rejected → Status rejected + alasan penolakan
```

**Tukar Shift (Swap):**
```
Karyawan A mengajukan tukar shift dengan Karyawan B
  │  - Pilih tanggal shift A
  │  - Pilih shift A saat ini
  │  - Pilih karyawan B (hanya ditampilkan dari departemen/divisi yang sama)
  │  - Pilih tanggal shift B (opsional: bisa sama atau berbeda)
  │  - Alasan
  ▼
Karyawan B mendapat notifikasi "Permintaan Tukar Shift"
  │  Konfirmasi: Setuju / Tidak Setuju
  ▼
  ├── ❌ Karyawan B tidak setuju → Request dibatalkan
  │
  └── ✅ Karyawan B setuju → Lanjut ke Approval Line kedua atasan
       ▼
Approval Line Karyawan A + Approval Line Karyawan B
  │  Approve / Reject
  ▼
  ├── ✅ Approved → Kedua jadwal shift tertukar
  │                  Tercatat di shift_change_requests
  │
  └── ❌ Rejected → Status rejected
```

##### Detail Field Pengajuan (`shift_change_requests`)

| Field | Tipe | Keterangan |
|-------|------|------------|
| **request_type** | ENUM | `individual` (ganti sendiri) / `swap` (tukar dengan karyawan lain) |
| **employee_id** | UUID (FK) | Karyawan yang mengajukan |
| **target_date** | DATE | Tanggal shift yang ingin diubah |
| **current_schedule_id** | UUID (FK) | Schedule/shift saat ini di tanggal tersebut |
| **requested_schedule_id** | UUID (FK) | Schedule/shift yang diminta |
| **swap_partner_id** | UUID (FK, opsional) | Karyawan tujuan tukar (hanya untuk swap) |
| **swap_partner_date** | DATE (opsional) | Tanggal shift partner yang ditukar (jika berbeda) |
| **swap_partner_schedule_id** | UUID (FK, opsional) | Schedule shift partner yang ditukar |
| **reason** | TEXT | Alasan pengajuan |
| **swap_partner_confirmed** | BOOLEAN | Apakah partner sudah setuju? (untuk swap) |
| **swap_partner_confirmed_at** | TIMESTAMPTZ | Waktu konfirmasi partner |
| **status** | ENUM | `pending` / `partner_pending` / `approved` / `rejected` / `cancelled` |
| **approval_trail** | JSONB | Riwayat approval (siapa approve/reject, kapan) |

> Status `partner_pending` khusus untuk swap — menunggu konfirmasi dari karyawan partner sebelum lanjut ke approval.

##### Aturan Bisnis

| Aturan | Detail |
|--------|--------|
| **Batas Pengajuan** | Maksimal **3x per bulan per karyawan** (dapat dikonfigurasi di `companies.approval_config.shift_change_max_per_month`) |
| **Batas Waktu Pengajuan** | Minimal H-2 sebelum tanggal shift yang diminta |
| **Hanya Shift** | Fitur ini hanya untuk perubahan **shift** (pagi/siang/malam) — tidak untuk jadwal 5/6 hari tetap |
| **Duplikasi** | Tidak boleh ada 2 request pending untuk karyawan + tanggal yang sama |
| **Konflik Jadwal** | Tanggal tujuan harus kosong (tidak sudah di-assign jadwal lain) |
| **Swap Partner** | Partner harus dari **departemen yang sama** (untuk memastikan kompetensi setara) |
| **Konfirmasi Partner** | Untuk swap, partner punya waktu **24 jam** untuk konfirmasi; jika lewat, request otomatis expired |
| **Approval** | Approval dari **masing-masing Approval Line** (untuk swap: butuh approval dari kedua atasan) |
| **Pembatalan** | Pengaju dapat membatalkan request yang masih `pending` atau `partner_pending` |
| **Audit Trail** | Semua perubahan shift tercatat di activity_logs |
| **Dampak Absensi** | Setelah approved, sistem absensi menggunakan shift baru untuk tanggal tersebut |

##### Tampilan di Web & Mobile

**Untuk Karyawan (Web & Mobile):**
- **Web:** Menu **"Request Shift Change"** di sidebar navigasi → form dengan pilih tanggal, shift tujuan, alasan — ditampilkan dengan AG Grid untuk riwayat
- **Mobile:** Menu **"Request Shift Change"** di halaman Request/Schedule (Tab 2) — form dengan card-based layout
- Pilih jenis: **Ganti Shift** (individu) atau **Tukar Shift** (swap)
- Pilih tanggal → pilih shift tujuan → tulis alasan
- Untuk swap: cari karyawan (auto-suggest dari departemen yang sama) → pilih tanggal shift partner
- Riwayat request: list semua pengajuan dengan status

**Untuk Approval Line (Web & Mobile):**
- **Web:** Tabel pending approval di dashboard (AG Grid) dengan filter per jenis request
- **Mobile:** **Tab "Pending Approval"** — termasuk request shift change
- Detail: nama karyawan, shift lama → shift baru, tanggal, alasan
- Tombol Approve / Reject

##### Konfigurasi di Company Settings

Ditambahkan ke `companies.approval_config`:

```json
{
  "shift_change_max_per_month": 3,
  "shift_change_min_hours_before": 48,
  "shift_change_swap_partner_timeout_hours": 24
}
```

---

### 5.15 Fitur Tambahan untuk Konteks Indonesia

#### 5.15.1 Manajemen BPJS

- Perhitungan otomatis iuran BPJS Kesehatan & Ketenagakerjaan
- Tracking status kepesertaan setiap karyawan
- Generate laporan untuk pelaporan BPJS
- Batch update data peserta BPJS
- Integrasi export file untuk laporan bulanan BPJS (format .txt/.csv)

#### 5.15.2 Perhitungan Pajak PPh 21 (Full Year)

- Penerapan skema TER (Tarif Efektif Rata-rata)
- Penghitungan ulang di bulan Desember (gross-up)
- A1/A2 tahunan untuk pelaporan SPT Tahunan
- Generate bukti potong 1721 A1/A2

#### 5.15.3 THR (Tunjangan Hari Raya)

- Perhitungan otomatis THR untuk semua karyawan
- THR proporsional untuk karyawan < 1 tahun
- THR untuk karyawan yang resign sebelum hari raya (prorata)
- Cutoff date untuk kelayakan THR

#### 5.15.4 Pph 21 Gross-up Method

- Opsi perhitungan gross-up untuk karyawan level tertentu
- Otomatis menghitung gross-up amount

#### 5.15.5 Report & Analytics Khas Indonesia

- Laporan Wajib Lapor Ketenagakerjaan Perusahaan (WLKP)
- Laporan rata-rata upah per jabatan
- Statistik komposisi karyawan (usia, gender, agama, pendidikan)
- Laporan turnover rate
- Rekap absensi untuk payroll
- Laporan PPh 21 (SPT Masa)

#### 5.15.6 Company Profile & Settings

- Informasi perusahaan (nama, alamat, NPWP perusahaan, logo)
- Kebijakan HR (jam kerja, aturan cuti, limit lembur, dll)
- Struktur organisasi perusahaan

#### 5.15.7 Log Aktivitas (Audit Trail)

- Semua perubahan data dicatat dengan: siapa, apa, kapan, IP address
- Log immutable (tidak bisa dihapus/diedit)
- Filter dan search berdasarkan user, tipe aksi, tanggal

#### 5.15.8 Notifikasi

- **In-App Notification:** Notification bell di web/mobile
- **Push Notification:** Untuk mobile PWA (via service worker)
- **Email Notification:** Untuk approval requests, pengumuman penting
- **WhatsApp Notification (Opsional):** Integrasi dengan API WhatsApp Business

#### 5.15.9 Dashboard & Report

- **Executive Dashboard:** Jumlah karyawan, absensi, payroll summary
- **HR Dashboard:** Turnover, headcount by department, cuti summary
- **Manager Dashboard:** Approval pending, KPI team, absensi team
- **Custom Reports:** Filter by date range, department, status, dll

#### 5.15.10 Employee Self-Service (ESS)

- **Profil Karyawan:** Edit data pribadi (terbatas)
- **Request:** Cuti, reimbursement, lembur, pinjaman
- **Documents:** Upload & manage dokumen pribadi
- **Payslip:** Lihat & download slip gaji
- **Attendance:** Riwayat absensi
- **KPI:** Lihat target KPI & hasil review
- **Reprimand:** Lihat riwayat teguran
- **Kalendar:** Lihat hari libur & event perusahaan

#### 5.15.11 Masa Percobaan (Probation Period)

**Aturan sesuai UU Ketenagakerjaan:**
- Maksimal 3 bulan masa percobaan
- Karyawan masuk dengan status `employment_status = 'percobaan'` (bukan `'tetap'`)
- Selama masa percobaan, berlaku upah minimum yang sama dengan karyawan tetap
- Tidak boleh ada masa percobaan untuk karyawan kontrak (PKWT)
- Wajib ada surat perjanjian kerja yang ditandatangani

**Fitur Sistem:**
- Tracking masa percobaan per karyawan
- Reminder otomatis H-30, H-14, H-7 menjelang berakhirnya masa percobaan
- Evaluasi akhir masa percobaan (form penilaian) yang diisi oleh atasan
- **Status otomatis berubah** dari `'percobaan'` → `'tetap'` jika melewati masa percobaan dengan sukses
- Opsi perpanjangan masa percobaan (dengan approval HR Manager)
- Notifikasi ke HR jika karyawan tidak lulus masa percobaan (status tetap `'percobaan'` atau diakhiri kontrak)

#### 5.15.12 Warning System / Notifikasi Peringatan

- Notifikasi otomatis untuk kontrak yang akan berakhir (H-30, H-14, H-7)
- Peringatan dokumen yang akan expired
- Peringatan sisa cuti yang akan hangus
- Notifikasi ulang tahun karyawan
- Peringatan pinjaman yang akan jatuh tempo

#### 5.15.13 Manajemen Resign & Exit

- Pengajuan resign dengan notice period
- Exit clearance checklist (HR, Finance, IT, Admin)
- Perhitungan hak karyawan resign (gaji akhir, THR proporsional, pesangon)
- Data alumni perusahaan

#### 5.15.14 Form Dynamic & Approval Workflow

- **Dynamic Form Builder:** HR bisa membuat form dengan field dinamis
- **Approval Flow Config:** Bisa dikonfigurasi per jenis transaksi
- **Multi-Level Approval:** Hingga 3+ level approval
- **Escalation:** Jika approval melebihi batas waktu, auto-escalate

---

## 6. Mobile Strategy: PWA vs Native

### 6.1 Rekomendasi: PWA (Progressive Web App)

**Rekomendasi utama:** **PWA berbasis SvelteKit** untuk fase pertama, dengan opsi migrasi ke native di fase berikutnya jika diperlukan.

### 6.2 Perbandingan

| Aspek | PWA | Native (Flutter/React Native) |
|-------|-----|------------------------------|
| **Biaya Pengembangan** | ✅ Lebih murah (1 codebase) | ❌ 2 codebase terpisah |
| **Waktu ke Market** | ✅ Lebih cepat | ❌ Lebih lambat |
| **Face Detection** | ✅ Face-api.js (TensorFlow.js) | ✅ Lebih optimal (ML Kit) |
| **GPS / Geolocation** | ✅ Geolocation API | ✅ Native GPS API |
| **Offline Access** | ✅ Service Worker + Cache | ✅ Lebih baik |
| **Push Notification** | ✅ Service Worker Push | ✅ Native Push |
| **Camera Access** | ✅ WebRTC / getUserMedia | ✅ Native Camera API |
| **Performance** | ⚠️ Cukup baik | ✅ Lebih baik |
| **Install ke Home Screen** | ✅ Via browser prompt | ✅ App Store/Play Store |
| **Akses Hardware** | ⚠️ Terbatas | ✅ Full access |
| **Biaya Publikasi** | ✅ Gratis (HTTPS saja) | ❌ Biaya store |
| **Updates** | ✅ Real-time (refresh) | ❌ Via store review |
| **Battery Efficiency** | ⚠️ Kurang efisien | ✅ Lebih efisien |

### 6.3 Strategi Implementasi PWA

1. **Gunakan SvelteKit dalam SPA Mode** dengan `vite-plugin-pwa`
2. **Configurasi Service Worker** untuk caching strategis:
   - Static assets: Cache-first
   - API responses: Network-first dengan fallback cache
   - Never cache PII (Personally Identifiable Information)
3. **Mobile-Specific PWA Features:**
   - Full-screen mode (display: standalone)
   - Splash screen
   - Offline fallback page
   - Background sync untuk absensi offline
4. **Caching Strategy:**
   - Stale-while-revalidate untuk static assets
   - Network-only untuk data sensitif
   - Background sync untuk absensi di daerah dengan sinyal lemah

### 6.4 Mobile-Only Features

| Fitur | Web | Mobile |
|-------|-----|--------|
| Absensi Face Detection | ❌ | ✅ |
| Absensi GPS Check-in | ❌ | ✅ |
| Cuti | ✅ | ✅ |
| Reimbursement | ✅ | ✅ |
| Payslip | ✅ | ✅ |
| Shift Change Request | ✅ | ✅ |
| KPI Review | ✅ | ✅ (read-only) |
| Manajemen Karyawan (CRUD) | ✅ | ❌ |
| Payroll Processing | ✅ | ❌ |
| Laporan & Analytics | ✅ | ❌ |
| Approve Request | ✅ | ✅ |
| Kalender Libur | ✅ | ✅ |
| Pengumuman | ✅ | ✅ |
| Reprimand | ✅ | ✅ (read) |
| Pinjaman | ✅ | ✅ |
| Lembur Request | ✅ | ✅ |
| Dokumen Karyawan | ✅ | ✅ (upload) |

### 6.5 Kapan Migrasi ke Native?

- Jika performa face detection kurang optimal di PWA
- Jika ada kebutuhan akses hardware spesifik (fingerprint, NFC untuk absent)
- Jika battery drain menjadi masalah signifikan
- Jika adopsi pengguna mobile rendah karena pengalaman PWA kurang memuaskan

---

## 7. Non-Functional Requirements

### 7.1 Performance

| Metrik | Target |
|--------|--------|
| Page Load Time (First Paint) | < 2 detik |
| Page Load Time (Full Load) | < 3 detik |
| API Response Time (95th percentile) | < 200ms |
| API Response Time (p95 untuk laporan berat) | < 2 detik |
| Concurrent Users | Minimal 500 concurrent |
| Data Grid Render (1000 rows) | < 1 detik |
| PWA Startup Time | < 3 detik |
| Face Detection + Matching | < 5 detik |

### 7.2 Availability & Reliability

- **Uptime:** 99.9% (kecuali maintenance terencana)
- **RTO (Recovery Time Objective):** < 4 jam
- **RPO (Recovery Point Objective):** < 15 menit
- **Backup:** Daily full backup, hourly incremental backup
- **Maintenance Window:** Minggu 00:00–04:00 WIB

### 7.3 Environment Strategy

| Environment | Tujuan | Konfigurasi |
|-------------|--------|-------------|
| **Development (dev)** | Pengembangan harian & debugging | Database development, mock email, log level debug |
| **Staging (stg)** | UAT & integration testing | Database staging terpisah, test email server, data dummy mendekati produksi |
| **Production (prod)** | Live system untuk pengguna riil | Database production, disaster recovery, full security, encrypted data riil |

**Data Security antar Environment:**
- Data sensitif di dev/staging menggunakan **data dummy/anonymized** — tidak boleh menggunakan data riil
- Enkripsi AES-256-GCM hanya aktif di production (dev/staging boleh plaintext untuk debugging)
- Production database tidak boleh direstore ke dev/staging tanpa proses anonymization terlebih dahulu
- Environment variables untuk kredensial: `.env` file (dev) → Secrets Manager / Vault (staging/prod)

### 7.4 Security

- **Enkripsi Data Sensitif:**
  - AES-256-GCM untuk data di database (KTP, NPWP, rekening bank, gaji)
  - TLS 1.3 untuk data in-transit
  - Encrypted at rest untuk file dokumen di S3/MinIO
- **Authentication:**
  - JWT dengan refresh token
  - MFA opsional
  - Session timeout: 8 jam (web), 30 hari (mobile dengan refresh)
- **Authorization:**
  - Role-Based Access Control (RBAC)
  - Row-level security untuk data per organisasi
  - Fine-grained permission per modul & aksi
- **Audit Trail:**
  - Semua mutasi data tercatat
  - Log immutable
  - Retention: minimal 5 tahun
- **Security Headers:**
  - CSP (Content Security Policy)
  - X-Frame-Options: DENY
  - X-Content-Type-Options: nosniff
  - Strict-Transport-Security
- **Rate Limiting:** Per endpoint, per user, per IP
- **Input Validation:** Server-side validation untuk semua input
- **SQL Injection Prevention:** Via sqlc (parameterized queries)

### 7.5 Scalability

- **Horizontal Scaling:** Backend Go bisa di-scale horizontally
- **Database:** PostgreSQL dengan connection pooling (PgBouncer)
- **Caching:** Redis untuk data yang sering diakses
- **CDN:** Static assets via CDN
- **Query Optimization:** Indexing, query analysis, slow query monitoring

### 7.6 Compliance & Regulatory

| Regulasi | Kepatuhan |
|----------|-----------|
| **UU Ketenagakerjaan No. 13/2003** | Cuti, lembur, THR, PHK, pesangon |
| **UU Cipta Kerja No. 11/2020** | Omnibus law ketenagakerjaan |
| **PP No. 58/2023** | PPh 21 Tarif Efektif Rata-rata |
| **BPJS Kesehatan & Ketenagakerjaan** | Iuran & pelaporan |
| **UU PDP No. 27/2022** | Perlindungan data pribadi |
| **PP No. 35/2021** | PKWT, ALIH DAYA, WFH, Upah |

### 7.7 Usability

- **Responsive Design:** Mobile-first untuk mobile, desktop untuk web
- **Offline Mode (PWA):** Beberapa fitur dapat diakses offline
- **Accessibility:** WCAG 2.1 Level AA
- **Bahasa:** Bahasa Indonesia (default), Inggris (opsional)
- **Onboarding:** User guide & tooltips untuk fitur baru

### 7.8 Maintainability

- **Code Quality:** TypeScript (frontend) + Go (backend) dengan strict typing
- **Testing:**
  - Unit test coverage: minimal 80%
  - Integration test untuk API endpoints
  - E2E test untuk critical flows
- **Documentation:**
  - API docs (OpenAPI/Swagger)
  - Developer README
  - Deployment guide
  - User manual
- **CI/CD:** Automated testing & deployment pipeline

---

## 8. User Roles & Permissions

### 8.1 Role Assignment

Setiap **karyawan memiliki satu role** yang ditentukan saat pembuatan akun. Role inheren dengan data karyawan — tidak ada pemisahan tabel user dan employee. Informasi role disimpan di kolom `role_id` pada tabel `employees`.

**Aturan Assignment:**
- Role ditentukan saat onboarding / pembuatan akun pertama kali
- Super Admin dapat mengubah role karyawan kapan saja
- Role menentukan akses ke setiap modul (lihat matrix di bawah)
- Satu karyawan = satu role (tidak multi-role)

### 8.2 Role Matrix

| Modul | Super Admin | HR Manager | HR Staff | Finance | Manager | Karyawan | Direktur |
|-------|:-----------:|:----------:|:--------:|:-------:|:-------:|:--------:|:--------:|
| **Manajemen Karyawan** | CRUD | CRUD | CRUD | R | R (tim) | R (self) | R |
| **Data Sensitif** | R | R | R | R | — | — | R |
| **Absensi** | CRUD | CRUD | CRUD | R | R (tim) | R (self) | R |
| **Manual Attendance** | CRUD | CRUD | CRUD | — | Appr | CRUD | R |
| **Shift Change** | CRUD | CRUD | R | — | Appr | CRUD | R |
| **Payroll** | CRUD | CRUD | U | CRUD | — | R (self) | R |
| **Cuti** | CRUD | CRUD | CRUD | R | Appr | CRUD | R |
| **Reimbursement** | CRUD | CRUD | CRUD | Appr | Appr | CRUD | R |
| **Lembur** | CRUD | CRUD | CRUD | R | Appr | CRUD | R |
| **Pinjaman** | CRUD | CRUD | CRUD | Appr | Appr | CRUD | Appr |
| **KPI** | CRUD | CRUD | U | — | CRUD (tim) | CRUD (self) | R |
| **Reprimand** | CRUD | CRUD | U | — | R (tim) | R (self) | R |
| **Payslip** | R | R | R | R | R (tim) | R (self) | R |
| **Pengumuman** | CRUD | CRUD | CRUD | CRUD | CRUD | R | CRUD |
| **Dokumen** | CRUD | CRUD | CRUD | R | R (tim) | CRUD (self) | R |
| **Company Settings** | CRUD | R | — | R | — | — | R |
| **User Management** | CRUD | R | — | — | — | — | R |
| **Laporan** | CRUD | CRUD | R | R | R (tim) | R (self) | R |

**Legenda:** C=Create, R=Read, U=Update, D=Delete, Appr=Approve

### 8.3 Custom Roles

- Super Admin dapat membuat custom roles
- Setiap role memiliki permission per modul & aksi
- Permission inheritance berdasarkan jabatan
- Custom roles dapat dibuat dengan kombinasi permission yang spesifik

---

## 9. UI/UX Guidelines

### 9.1 Desktop Web (SvelteKit + Flowbite + AG Grid)

- **Layout:** Sidebar navigation + Top bar + Content area
- **Sidebar:** Navigasi utama modul, collapsible
- **Top bar:** Search global, notifications, user profile
- **Content:** AG Grid untuk semua tampilan tabel data
  - Column sorting, filtering, grouping
  - Row selection (checkbox)
  - Export to Excel/CSV/PDF
  - Inline editing (untuk HR)
  - Pagination (default 25, 50, 100 rows)
- **Forms:** Modal forms untuk create/edit
- **Dashboard:** Card-based layout dengan charts

### 9.2 Mobile PWA (No Hamburger Menu, Card-Based)

- **Layout:** Bottom tab navigation (max 5 tabs) + Top app bar
  - Tab 1: Dashboard / Home
  - Tab 2: Request (Cuti, Reimburse, Lembur, Pinjaman)
  - Tab 3: Absensi (Check-in/out)
  - Tab 4: Payslip & Profile
  - Tab 5: Lainnya (More) — berisi akses cepat ke:
    - Kalender Hari Libur
    - Pengumuman
    - Kalender Pribadi (Cuti, Izin)
    - FAQ / Bantuan
- **Instead of AG Grid:** Card-based list views
  - Setiap item dirender sebagai card dengan informasi utama
  - Tapping card → Detail page
  - Swipe actions (approve/reject)
  - Pull-to-refresh
  - Infinite scroll
- **FAB (Floating Action Button):** Untuk aksi utama (absen, buat request baru)
- **Bottom Sheet:** Untuk menu actions (edit, delete, share)
- **Navigation:** Stack navigation (push/pop) + Deep linking

### 9.3 Design System

- **Color Palette:** Professional blue-primary (#1A56DB) + neutral grays
- **Typography:** Inter font family (optimized untuk web & mobile)
- **Icons:** Heroicons (via `unplugin-icons`) — varian `20-solid` untuk Flowbite compatibility. Lucide sebagai fallback jika Heroicons tidak memiliki icon yang diperlukan.
- **Spacing:** Consistent 4-point grid system
- **Dark Mode:** Support dark mode (berdasarkan system preference)
- **Loading States:** Skeleton loading untuk semua page
- **Empty States:** Ilustrasi + copy untuk data kosong
- **Error States:** Error message yang human-readable
- **Toast Notifications:** Untuk success/error/info feedback

### 9.4 Responsive Breakpoints

| Breakpoint | Target Device | Layout |
|-----------|--------------|--------|
| < 640px | Mobile phones | Mobile layout (card-based) |
| 640–1023px | Tablets | Hybrid (sidebar collapsed) |
| 1024–1279px | Desktop small | Desktop layout |
| 1280px+ | Desktop large | Desktop layout (full) |

---

## 10. Security & Data Privacy

### 10.1 Enkripsi Data Sensitif

**Data yang wajib dienkripsi (AES-256-GCM):**
- NIK (Nomor Induk Kependudukan)
- NPWP
- Nomor rekening bank
- Gaji pokok & komponen gaji
- Alamat lengkap (KTP)
- Dokumen pribadi (KTP, KK, ijazah)

**Implementasi Enkripsi:**
```
Encrypt AKS:
[Plaintext] → AES-256-GCM → [Base64 Ciphertext]

Decrypt:
[Base64 Ciphertext] → AES-256-GCM → [Plaintext]

Key Management:
- Master key disimpan terpisah dari database (environment variable / vault)
- Key rotation setiap 90 hari
- Setiap data sensitif memiliki key yang berbeda (derived key)
```

### 10.2 Authentication & Authorization

- **JWT Access Token:** 15 menit expiry
- **JWT Refresh Token:** 7 hari expiry (single-use)
- **MFA (Opsional):** TOTP via Google Authenticator
- **Password Policy:**
  - Minimum 8 karakter
  - Harus mengandung huruf besar, kecil, angka, dan simbol
  - Tidak boleh sama dengan 5 password terakhir
  - Expiry setiap 90 hari
- **Failed Login:** Lockout setelah 5 percobaan gagal (15 menit)

### 10.3 Data Privacy (UU PDP)

- **Data Retention:** Data karyawan diarsipkan 5 tahun setelah resign
- **Data Deletion:** Permintaan hapus data sesuai UU PDP
- **Consent:** Persetujuan penggunaan data pribadi saat onboarding
- **Data Access:** Karyawan bisa meminta export data pribadi kapan saja
- **Data Breach Notification:** Notifikasi ke pihak berwenang dalam 72 jam

---

## 11. Data Model Overview

### 11.1 Entity Relationship (High-Level)

```
roles
  ├── id (PK)
  ├── name (Super Admin / HR Manager / HR Staff / Finance / Manager / Karyawan / Direktur)
  ├── slug (super_admin / hr_manager / hr_staff / finance / manager / employee / director)
  ├── description
  ├── is_system_role (true untuk role bawaan, tidak bisa dihapus)
  ├── permissions (JSONB — daftar permission per modul & aksi)
  ├── is_active
  └── created_at, updated_at

employees
  ├── id (PK)
  ├── employee_id (NIP, unique)
  ├── full_name
  ├── encrypted_nik
  ├── encrypted_npwp
  ├── encrypted_bank_account
  ├── birth_place, birth_date
  ├── gender
  ├── religion
  ├── marital_status
  ├── marital_status_count (tanggungan)
  ├── address_ktp, address_domicile
  ├── phone
  ├── email (digunakan sebagai username / login ID, unique)
  ├── password_hash (bcrypt/argon2 — untuk autentikasi login)
  ├── role_id (FK ke roles — menentukan hak akses)
  ├── last_login_at (timestamp terakhir login)
  ├── is_locked (boolean — terkunci karena gagal login)
  ├── locked_until (timestamp — sampai kapan terkunci)
  ├── emergency_contact_name, emergency_contact_phone
  ├── join_date
  ├── end_date (jika kontrak/resign)
  ├── employment_status (tetap/kontrak/percobaan/harian)
  ├── ptkp_status (K/0, K/1, dll)
  ├── tax_method (gross/gross-up/nett)
  ├── face_embedding (vector untuk face recognition)
  ├── is_remote (boleh absen tanpa koordinat)
  ├── department_id (FK)
  ├── position_id (FK)
  ├── position_level_id (FK)
  ├── approval_line_id (FK ke employees — penanggung jawab approval untuk karyawan ini)
  ├── work_schedule_id (FK) — override individu (jika berbeda dari default departemen)
  ├── bpjs_config (JSONB) — override BPJS per karyawan: enable/disable per komponen (kesehatan, jht, jp, jkk, jkm) + custom rate
  ├── photo_url
  ├── is_active
  ├── resigned_at
  ├── created_at, updated_at, deleted_at
  └──

departments
  ├── id (PK)
  ├── name
  ├── code
  ├── parent_id (FK, self-referencing)
  ├── head_id (FK ke employees)
  ├── work_schedule_id (FK ke work_schedules — jadwal default untuk semua karyawan di departemen ini)
  └── sort_order

positions
  ├── id (PK)
  ├── name
  ├── department_id (FK)
  └── grade_id (FK)

position_grades / levels
  ├── id (PK)
  ├── name (Staff/Supervisor/Manager/Direktur)
  └── level (1-10)

work_schedules
  ├── id (PK)
  ├── name
  ├── type (5_day / 6_day)
  ├── monday_start, monday_end
  ├── tuesday_start, tuesday_end
  ├── ... (per hari)
  └── break_start, break_end

attendance_records
  ├── id (PK)
  ├── employee_id (FK)
  ├── date
  ├── check_in_time
  ├── check_out_time
  ├── check_in_photo_url
  ├── check_out_photo_url
  ├── check_in_lat, check_in_lng
  ├── check_out_lat, check_out_lng
  ├── check_in_location_name
  ├── check_out_location_name
  ├── face_match_score
  ├── status (hadir/terlambat/izin/sakit/tanpa_keterangan)
  ├── is_manual_entry
  └── device_info (user agent, platform)

attendance_locations
  ├── id (PK)
  ├── name (Kantor Pusat, Cabang A, Gudang)
  ├── latitude
  ├── longitude
  ├── radius (meter)
  ├── address
  └── is_active

payroll_periods
  ├── id (PK)
  ├── month, year
  ├── start_date, end_date
  ├── status (draft/completed/approved/paid)
  └── approved_by, approved_at

payroll_items
  ├── id (PK)
  ├── payroll_period_id (FK)
  ├── employee_id (FK)
  ├── base_salary
  ├── daily_wage
  ├── allowances (JSON: {position, transport, meal, ...})
  ├── overtime_pay
  ├── thr_amount
  ├── gross_salary
  ├── deductions (JSON: {pph21, bpjs_kesehatan, bpjs_jht, bpjs_jp, loan, ...})
  ├── company_contributions (JSON: {bpjs_ks, bpjs_jht, bpjs_jp, bpjs_jkk, bpjs_jkm})
  ├── net_salary (take home pay)
  └── status

leave_types
  ├── id (PK)
  ├── name
  ├── default_quota (per year)
  ├── is_paid (paid/unpaid)
  ├── requires_document
  ├── max_consecutive_days
  └── can_rollover

leave_requests
  ├── id (PK)
  ├── employee_id (FK)
  ├── leave_type_id (FK)
  ├── start_date, end_date
  ├── total_days
  ├── reason
  ├── document_url (surat dokter, dll)
  ├── status (pending/approved/rejected/cancelled)
  └── approved_by (JSON cascade)

leave_balances
  ├── id (PK)
  ├── employee_id (FK)
  ├── leave_type_id (FK)
  ├── year
  ├── total_quota
  ├── used
  ├── remaining
  └── rolled_over_from_previous_year

reimbursements
  ├── id (PK)
  ├── employee_id (FK)
  ├── type (medical/training/travel/supplies/other)
  ├── amount
  ├── description
  ├── receipt_url (lampiran)
  ├── status
  └── approval_trail (JSON)

overtime_requests
  ├── id (PK)
  ├── employee_id (FK)
  ├── date
  ├── start_time, end_time
  ├── total_hours
  ├── type (weekday/weekend/holiday)
  ├── reason
  ├── is_approved
  └── approval_trail (JSON)

loans
  ├── id (PK)
  ├── employee_id (FK)
  ├── type (regular/emergency/education)
  ├── amount
  ├── installment_count (tenor)
  ├── installment_amount (per bulan)
  ├── interest_rate (0% default)
  ├── payment_method (payroll_deduction/manual_transfer)
  ├── remaining_balance
  ├── status (pending/approved/active/completed/rejected)
  └── approval_trail (JSON)

loan_installments
  ├── id (PK)
  ├── loan_id (FK)
  ├── installment_number
  ├── amount
  ├── due_date
  ├── paid_date
  └── status (pending/paid/skipped)

kpi_templates
  ├── id (PK)
  ├── title
  ├── position_id (FK)
  ├── period_type (yearly/quarterly)
  └── year

kpi_indicators
  ├── id (PK)
  ├── kpi_template_id (FK)
  ├── name
  ├── target
  ├── weight (percentage)
  └── measurement_unit

kpi_reviews
  ├── id (PK)
  ├── employee_id (FK)
  ├── kpi_template_id (FK)
  ├── period (Q1, Q2, Q3, Q4, or yearly)
  ├── year
  ├── self_rating (JSON: indicator_id, score, note)
  ├── manager_rating (JSON)
  ├── final_score
  ├── category (outstanding/above/meets/needs/underperform)
  ├── status (draft/self_review/manager_review/hr_review/completed)
  └── reviewed_by, reviewed_at

reprimands
  ├── id (PK)
  ├── employee_id (FK)
  ├── type (verbal/sp1/sp2/sp3)
  ├── title
  ├── description
  ├── issued_by (FK ke employees)
  ├── issued_date
  ├── acknowledgment_date
  ├── document_url
  └── status (issued/acknowledged/expired)

announcements
  ├── id (PK)
  ├── title
  ├── content (HTML/markdown)
  ├── type (general/important/emergency)
  ├── target_department_id (FK, nullable)
  ├── target_position_level (nullable)
  ├── attachment_url (nullable)
  ├── is_pinned
  ├── created_by (FK ke employees)
  ├── published_at
  └── expired_at

company_holidays
  ├── id (PK)
  ├── date
  ├── name
  ├── type (national/joint/company)
  ├── is_recurring_yearly
  └── description

employee_documents
  ├── id (PK)
  ├── employee_id (FK)
  ├── type (ktp/kk/ijazah/sertifikat/kontrak/npwp/bpjs/photo/other)
  ├── file_name
  ├── file_url (encrypted storage path)
  ├── file_size
  ├── mime_type
  ├── status (pending/verified/rejected)
  ├── verified_by (FK)
  ├── verified_at
  ├── expiry_date (untuk dokumen dengan masa berlaku)
  └── notes

activity_logs (audit trail)
  ├── id (PK)
  ├── user_id (FK)
  ├── action (create/update/delete/approve/reject/view)
  ├── entity_type (employee/leave/payroll/etc)
  ├── entity_id
  ├── old_values (JSON)
  ├── new_values (JSON)
  ├── ip_address
  ├── user_agent
  └── created_at

companies
  ├── id (PK)
  ├── name
  ├── legal_name
  ├── address
  ├── city, province, postal_code
  ├── phone, email, website
  ├── npwp
  ├── logo_url
  ├── bpjs_ks_number            (BPJS Kesehatan)
  ├── bpjs_jht_number           (BPJS JHT)
  ├── bpjs_jp_number            (BPJS JP)
  ├── bpjs_jkk_rate             (risk-based: 0.24%–1.74%)
  ├── hr_settings (JSONB)       (HR config: BPJS rates, overtime defaults, dll)
  ├── approval_config (JSONB)   (Approval flow: leave limit, loan limit, shift change, dll)
  ├── is_active
  ├── created_at, updated_at, deleted_at
  └──

employee_emergency_contacts
  ├── id (PK)
  ├── employee_id (FK)
  ├── name
  ├── relationship
  ├── phone
  ├── address
  ├── is_primary
  ├── created_at, updated_at
  └──

notification_preferences
  ├── id (PK)
  ├── employee_id (FK)
  ├── notification_type (approval_request / approved / rejected / announcement / reminder / system)
  ├── channel (in_app / email / whatsapp)
  ├── is_enabled (BOOLEAN DEFAULT TRUE)
  ├── created_at, updated_at
  └──

user_sessions
  ├── id (PK)
  ├── user_id (FK ke employees)
  ├── refresh_token (hashed)
  ├── device_info (user agent, platform)
  ├── ip_address
  ├── expires_at
  ├── is_revoked
  ├── created_at
  └──

password_reset_tokens
  ├── id (PK)
  ├── user_id (FK ke employees)
  ├── token (hashed)
  ├── expires_at
  ├── is_used
  ├── created_at
  └──

login_attempts
  ├── id (PK)
  ├── user_id (FK ke employees, nullable — untuk log gagal login dengan email salah)
  ├── email_attempted
  ├── ip_address
  ├── is_success
  ├── failed_reason
  ├── created_at
  └──

employee_salary_components
  ├── id (PK)
  ├── employee_id (FK — karyawan pemilik komponen ini)
  ├── component_name (Tunj. Jabatan, Tunj. Transport, dll)
  ├── component_type (allowance / deduction)
  ├── amount (nilai nominal)
  ├── is_active (TRUE/FALSE — nonaktif tanpa hapus)
  ├── effective_date (tanggal mulai berlaku)
  └── created_at, updated_at

employee_salary_component_histories
  ├── id (PK)
  ├── employee_id (FK)
  ├── component_name
  ├── old_amount (nilai sebelum perubahan)
  ├── new_amount (nilai setelah perubahan)
  ├── change_reason (alasan perubahan)
  ├── changed_by (FK ke employees — siapa yang mengubah)
  └── changed_at

position_grade_overtime_rates
  ├── id (PK)
  ├── position_grade_id (FK — level jabatan)
  ├── overtime_type (weekday / weekend / holiday)
  ├── hour_from (mulai jam ke-)
  ├── hour_to (sampai jam ke-, NULL = tak terbatas)
  ├── multiplier (pengali: 1.5, 2.0, 2.5, dst)
  ├── is_active
  ├── effective_date
  └── created_at, updated_at

employee_overtime_rates
  ├── id (PK)
  ├── employee_id (FK — override individu)
  ├── overtime_type (weekday / weekend / holiday)
  ├── hour_from (mulai jam ke-)
  ├── hour_to (sampai jam ke-, NULL = tak terbatas)
  ├── multiplier (pengali: 1.5, 2.0, 2.5, dst)
  ├── is_active
  ├── effective_date
  └── created_at, updated_at

shift_change_requests
  ├── id (PK)
  ├── request_type (individual / swap)
  ├── employee_id (FK — pengaju)
  ├── target_date (tanggal shift yang ingin diubah)
  ├── current_schedule_id (FK ke work_schedules — shift saat ini)
  ├── requested_schedule_id (FK ke work_schedules — shift tujuan)
  ├── swap_partner_id (FK ke employees — untuk swap, opsional)
  ├── swap_partner_date (tanggal shift partner, opsional)
  ├── swap_partner_schedule_id (FK ke work_schedules, opsional)
  ├── reason (alasan pengajuan)
  ├── swap_partner_confirmed (BOOLEAN — apakah partner sudah setuju?)
  ├── swap_partner_confirmed_at
  ├── status (pending / partner_pending / approved / rejected / cancelled)
  ├── approval_trail (JSONB — riwayat approval)
  ├── rejection_reason
  ├── created_at, updated_at
  └──

notifications
  ├── id (PK)
  ├── user_id (FK)
  ├── type (approval_request/approved/rejected/announcement/reminder)
  ├── title
  ├── body
  ├── data (JSON, untuk deep linking)
  ├── is_read
  └── created_at
```

---

## 12. Integrasi Pihak Ketiga

### 12.1 Integrasi Wajib (Fase 1)

| Layanan | Kegunaan | Metode |
|---------|----------|--------|
| **SMTP Email** | Notifikasi email (approval, pengumuman) | SMTP / SendGrid / Mailgun |
| **Object Storage** | Dokumen karyawan, foto absensi | MinIO / AWS S3 / DigitalOcean Spaces |

### 12.2 Integrasi Opsional (Fase 2+)

| Layanan | Kegunaan |
|---------|----------|
| **WhatsApp Business API** | Notifikasi WhatsApp untuk approval request |
| **Cloud Face Recognition** | API face comparison (AWS Rekognition, Azure Face API) |
| **Coretax DJP** | Pelaporan PPh 21 otomatis (saat API tersedia) |
| **BPJS Online** | Pelaporan BPJS otomatis |
| **Bank API** | Disbursement payroll otomatis ke rekening |
| **Google Calendar / Outlook** | Sync kalender cuti & hari libur |
| **Slack / Teams** | Notifikasi HR ke workspace |
| **E-Signature** | Tanda tangan digital (Mekari Sign, VIDA, Privy) |

---

## 13. Roadmap & Milestone

### Fase 1 — MVP (3–4 bulan)

**Prioritas: Foundation & Core HR Operations**

1. ✅ Setup project: SvelteKit SPA + PWA + Go Fiber + PostgreSQL
2. ✅ Authentication & Role Management
3. ✅ Master Data: Employee CRUD + Department + Position
4. ✅ Dokumen Karyawan (upload, verify)
5. ✅ Schedule Kerja (5/6 hari)
6. ✅ Absensi Mobile (Face Detection + GPS Check-in/out)
7. ✅ Manajemen Cuti (alur dasar)
8. ✅ Dashboard dasar & struktur organisasi
9. ✅ Pengumuman
10. ✅ Kalender Hari Libur

### Fase 2 — Payroll & Financial (2–3 bulan)

1. ✅ Payroll Engine (Gaji pokok, tunjangan, upah harian)
2. ✅ PPh 21 TER (Jan–Nov + Dec recap)
3. ✅ BPJS Kesehatan & Ketenagakerjaan
4. ✅ THR (Tunjangan Hari Raya)
5. ✅ Slip Gaji (Payslip)
6. ✅ Reimbursement
7. ✅ Lembur (Overtime)
8. ✅ Pinjaman Karyawan

### Fase 3 — Advanced HR (2–3 bulan)

1. ✅ KPI & Performance Review
2. ✅ Reprimand / Surat Peringatan
3. ✅ Mutasi & Promosi
4. ✅ Resign & Exit Management
5. ✅ Laporan & Analytics Lengkap
6. ✅ Notifikasi (In-app, Email, Push)
7. ✅ Audit Trail Enhanced

### Fase 4 — Optimization & Integration (2 bulan)

1. ✅ Integrasi WhatsApp Notification
2. ✅ Coretax DJP Integration
3. ✅ Performance Optimization
4. ✅ User Acceptance Testing
5. ✅ Deployment & Go-Live

---

## 14. Success Metrics

### 14.1 Key Performance Indicators

| Metrik | Target (3 bulan setelah Go-Live) |
|--------|----------------------------------|
| Adopsi Pengguna (Aktif/Monthly) | > 90% dari total karyawan |
| Adopsi Absensi Mobile | > 95% dari total absensi |
| Waktu Proses Payroll | < 3 hari kerja (dari sebelumnya manual) |
| Waktu Approval Cuti | < 24 jam |
| Waktu Approval Reimbursement | < 48 jam |
| Time-to-Hire (rekrutmen) | N/A (fase 1) |
| Kepuasan Pengguna (Survey) | > 80% |
| Error Rate | < 1% dari total transaksi |
| Data Entry Error | Berkurang 90% (dari manual) |
| Downtime | < 1 jam/bulan |

### 14.2 Business Impact Metrics

- **Efisiensi HR:** Berkurangnya 70% waktu administratif HR
- **Paperless:** Berkurangnya 90% penggunaan kertas untuk HR operations
- **Transparansi:** 100% karyawan bisa mengakses payslip & data kehadiran
- **Compliance:** 100% kepatuhan terhadap regulasi ketenagakerjaan & perpajakan

---

## Appendices

### Appendix A: Glossary

| Istilah | Definisi |
|---------|----------|
| **SPA** | Single Page Application — Aplikasi web satu halaman |
| **PWA** | Progressive Web App — Aplikasi web yang bisa diinstal seperti native |
| **sqlc** | Code generator Go dari query SQL |
| **Fiber** | Web framework Go yang cepat |
| **PPh 21** | Pajak Penghasilan Pasal 21 (pajak gaji) |
| **TER** | Tarif Efektif Rata-rata — skema perhitungan PPh 21 terbaru |
| **BPJS** | Badan Penyelenggara Jaminan Sosial |
| **THR** | Tunjangan Hari Raya |
| **PTKP** | Penghasilan Tidak Kena Pajak |
| **SP** | Surat Peringatan |
| **PKWT** | Perjanjian Kerja Waktu Tertentu (kontrak) |
| **PKWTT** | Perjanjian Kerja Waktu Tidak Tertentu (tetap) |
| **RBAC** | Role-Based Access Control |
| **AES-256-GCM** | Advanced Encryption Standard 256-bit Galois/Counter Mode |

### Appendix B: Risk Assessment

| Risiko | Dampak | Mitigasi |
|--------|--------|----------|
| Service worker cache menyimpan data sensitif | Tinggi | Never cache PII; network-only untuk data payroll |
| Face detection tidak akurat di kondisi minim cahaya | Sedang | Guidance pencahayaan; fallback approval manual |
| Koneksi internet tidak stabil untuk absensi | Tinggi | Background sync; queue absensi offline |
| Perubahan regulasi pajak/BPJS | Tinggi | Parameterized config; update berkala dari sumber resmi |
| Adopsi pengguna rendah | Sedang | Training; UI/UX yang intuitif; gamification |

---

*Dokumen ini disusun sebagai panduan pengembangan aplikasi HRMS. Setiap perubahan pada dokumen ini harus melalui proses change request dan disetujui oleh stakeholder terkait.*
