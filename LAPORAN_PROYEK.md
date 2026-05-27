# LAPORAN_PROYEK.md — CI/CD Pipeline TaskFlow API

## 1. Identitas Kelompok & Tool

| Atribut           | Detail                                    |
| ----------------- | ----------------------------------------- |
| **Kelompok**      | 1                                         |
| **Tool CI/CD**    | GitHub Actions                            |
| **Platform**      | github.com                                |
| **Registry**      | GitHub Container Registry (GHCR)          |
| **Repository**    | https://github.com/v0rein/taskflow-api    |

---

## 2. Diagram Alur Pipeline

```
Developer push kode ke main/develop
        │
        ▼
┌─── GitHub Actions Pipeline ──────────────────────────────┐
│                                                           │
│  ┌──────────┐  ┌──────────────────┐  ┌────────────────┐ │
│  │  go vet  │  │ Unit Test + Race │  │ Integration    │ │
│  │  (lint)  │  │ Coverage ≥ 75%   │  │ Test (Postgres)│ │
│  └────┬─────┘  └────────┬─────────┘  └───────┬────────┘ │
│       │                 │                     │          │
│       └─────────────────┼─────────────────────┘          │
│                         │                                │
│                         ▼                                │
│              ┌─────────────────────┐                     │
│              │ Docker Build + Push │                     │
│              │  → GHCR :sha-xxx   │                     │
│              │  → GHCR :latest    │                     │
│              └──────────┬──────────┘                     │
│                         │                                │
│                         ▼                                │
│              ┌─────────────────────┐                     │
│              │    Smoke Test       │                     │
│              │  /health ✅         │                     │
│              │  /api/v1/stats ✅   │                     │
│              │  → Tag :stable     │                     │
│              └──────────┬──────────┘                     │
│                         │                                │
│  ┌──────────────────┐   │   ┌──────────────────┐        │
│  │ SAST (gosec)     │   │   │ SCA (trivy fs)   │        │
│  │ Security Report  │   │   │ CVE Report       │        │
│  └──────────────────┘   │   └──────────────────┘        │
│                         │                                │
│                         ▼                                │
│              ┌─────────────────────┐                     │
│              │ Telegram Notifikasi │                     │
│              │  ✅ Sukses / ❌ Gagal│                     │
│              └─────────────────────┘                     │
└──────────────────────────────────────────────────────────┘
```

---

## 3. Tabel 3 Bug: Fix & Test

| # | File | Baris | Kode Salah | Kode Benar | Test yang Mendeteksi |
|---|------|-------|------------|------------|---------------------|
| 1 | `validator_test.go` | 29-33 | Syntax error: string literal dan `} else {` tanpa `if`, menyebabkan compile error | `t.Errorf("IsValidPriority(%q) = %v, want %v", ...)` | `TestIsValidPriority` (semua subtest) |
| 2 | `memory.go` / `postgres.go` | 57 / 112 | `FindByStatus` menggunakan `!=` → mengembalikan task yang BUKAN status tersebut | `t.Status == status` / `WHERE status = $1` | `TestFindByStatus_HanyaTodo`, `TestFindByStatus_HanyaDone` |
| 3 | `service.go` | 170 | `completed / len(tasks) * 100` → integer division, selalu 0 | `float64(completed) / float64(len(tasks)) * 100` | `TestCalculateCompletionRate` (subtest "1 dari 3") |

> **Catatan**: Bug #2 dan #3 sudah diperbaiki di source code yang diberikan. Bug #1 (syntax error di test file) ditemukan dan diperbaiki oleh kami.

---

## 4. Screenshot Pipeline

> **TODO**: Tambahkan screenshot setelah pipeline dijalankan
>
> - [ ] Screenshot pipeline HIJAU (semua test PASS)
> - [ ] Screenshot pipeline MERAH (saat bug dimasukkan kembali)

---

## 5. Perbandingan Ukuran Docker Image

| Metode Build                  | Ukuran Image | Keterangan                        |
| ----------------------------- | ------------ | --------------------------------- |
| `FROM golang:1.22` (single)   | ~800 MB      | Seluruh Go toolchain + OS         |
| Multi-stage (`scratch`)       | ~8 MB        | Hanya binary + CA certificates    |
| **Pengurangan**               | **~99%**     | 800 MB → 8 MB                    |

Mengapa `scratch` aman: Image `scratch` tidak mengandung OS, shell, atau library sistem lainnya. Ini berarti:
- Tidak ada vulnerability OS-level
- Attack surface sangat kecil
- Binary Go bersifat static (CGO_ENABLED=0)

---

## 6. Bukti Image di Registry

> **TODO**: Tambahkan setelah pipeline pertama selesai
>
> - URL GHCR: `ghcr.io/v0rein/taskflow-api`
> - Tag SHA: `sha-xxxxxxx`
> - Tag stable: `stable`

---

## 7. Smoke Test & Notifikasi (S4)

Smoke test berjalan otomatis di pipeline setelah image di-push:

```bash
# Health check
curl -sf http://localhost:8080/health
# → {"status":"ok","service":"taskflow-api","version":"1.0.0",...}

# Stats endpoint
curl -sf http://localhost:8080/api/v1/stats
# → {"total":0,"by_status":{"todo":0,...},"completion_rate_percent":0}
```

> **TODO**: Tambahkan screenshot notifikasi Telegram (sukses & gagal)

---

## 8. Prosedur Rollback (S5)

Lihat file [ROLLBACK_PROCEDURE.md](ROLLBACK_PROCEDURE.md) untuk prosedur lengkap.

**Demo rollback:**
1. Commit kode dengan bug (misalnya: kembalikan integer division)
2. Push → pipeline hijau (bug di logika, bukan compile error)
3. Verifikasi `/api/v1/stats` mengembalikan `completion_rate_percent: 0`
4. Rollback: `make rollback ROLLBACK_TAG=sha-<commit-sebelumnya>`
5. Verifikasi stats kembali benar

---

## 9. Audit Keamanan Pipeline (S6)

### Kategori B — SAST (gosec)

**Tool**: [gosec](https://github.com/securego/gosec) — Static Application Security Testing khusus Go.

**Rule yang relevan untuk TaskFlow:**
- G101: Hardcoded credentials
- G201: SQL injection (relevan karena pakai pgx)
- G304: File path injection
- G501: Weak crypto import

**Temuan**: (TODO: isi setelah pipeline berjalan)

### Kategori A — SCA (trivy fs)

**Tool**: [Trivy](https://github.com/aquasecurity/trivy) — Vulnerability scanner dari Aqua Security.

**Fungsi**: Memeriksa dependency Go (go.mod/go.sum) terhadap database CVE yang diketahui.

**Temuan**: (TODO: isi setelah pipeline berjalan)

**Analisis False Positive vs True Positive**: (TODO: isi setelah review hasil scan)

---

## 10. Refleksi

### Keunggulan GitHub Actions
- Terintegrasi langsung dengan GitHub (tidak perlu setup server terpisah)
- GHCR gratis untuk public repo
- Marketplace dengan ribuan actions siap pakai
- YAML syntax yang relatif mudah dipahami

### Keterbatasan
- Runner gratis terbatas (2000 menit/bulan untuk private repo)
- Tidak bisa di-host sendiri tanpa self-hosted runner
- Debug pipeline lebih sulit dibanding tools lokal seperti Jenkins

---

*Dokumen ini dibuat sebagai bagian dari tugas CI/CD Pipeline — Pertemuan 9.*
