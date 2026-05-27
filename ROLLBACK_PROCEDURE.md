# Prosedur Rollback — TaskFlow API

## Kapan Rollback Dibutuhkan?

Rollback dilakukan ketika:
- Versi baru menyebabkan error/bug kritis di production
- Endpoint penting (`/health`, `/api/v1/stats`) tidak berfungsi
- Performance turun drastis setelah deploy

---

## Langkah-Langkah Rollback

### 1. Identifikasi Masalah
```bash
# Cek apakah server masih merespons
curl -f http://localhost:8080/health

# Cek endpoint utama
curl -f http://localhost:8080/api/v1/stats
```

### 2. Temukan Tag Image Versi Sebelumnya
```bash
# Lihat daftar tag image yang tersedia di GHCR
# Buka: https://github.com/v0rein/taskflow-api/pkgs/container/taskflow-api
# Atau gunakan tag `stable` — selalu mengarah ke versi terakhir yang lolos smoke test

# Cek image yang sedang berjalan
docker inspect taskflow-api --format '{{.Config.Image}}'
```

### 3. Jalankan Rollback
```bash
# Rollback menggunakan Makefile (cara yang direkomendasikan)
make rollback ROLLBACK_TAG=sha-<commit-lama>

# Atau rollback ke tag stable
make rollback ROLLBACK_TAG=stable
```

#### Detail perintah `make rollback`:
1. Pull image versi lama dari registry
2. Stop container yang sedang berjalan
3. Jalankan container baru dengan image versi lama
4. Tunggu server siap (5 detik)
5. Verifikasi health check otomatis

### 4. Verifikasi Rollback Berhasil
```bash
# Health check
curl -f http://localhost:8080/health
# Output: {"status":"ok","service":"taskflow-api",...}

# Cek endpoint yang bermasalah
curl http://localhost:8080/api/v1/stats
# Pastikan completion_rate_percent benar (bukan 0)
```

### 5. Investigasi dan Perbaikan
- Review commit yang menyebabkan masalah
- Perbaiki bug di branch terpisah
- Push fix → pipeline CI/CD berjalan otomatis
- Jika semua test PASS → deploy otomatis tanpa perlu rollback lagi

---

## Tabel Perbandingan: Rollback Manual vs Otomatis

|                      | Cara Lama (Manual)                             | Dengan CI/CD + Tagging           |
| -------------------- | ---------------------------------------------- | -------------------------------- |
| **Langkah**          | SSH → stop → cari binary lama → copy → run   | `make rollback ROLLBACK_TAG=...` |
| **Waktu**            | ~25 menit                                      | < 60 detik                       |
| **Risiko Kesalahan** | Tinggi (banyak langkah manual)                 | Rendah (satu perintah)           |
| **Verifikasi**       | Manual cek browser                             | Health check otomatis            |
| **Traceability**     | Tidak ada (binary tanpa label)                 | Tag SHA terlacak di registry     |

---

## Tag Strategy

| Tag               | Kapan Diperbarui                            | Kegunaan                    |
| ------------------ | ------------------------------------------- | --------------------------- |
| `sha-<7char>`     | Setiap push ke `main`                       | Identifikasi unik per commit |
| `latest`          | Setiap push ke `main`                       | Versi terbaru (bisa bermasalah) |
| `stable`          | Hanya saat smoke test PASS                  | **Versi aman untuk rollback** |

---

*Dokumen ini dibuat sebagai bagian dari tugas CI/CD Pipeline — Pertemuan 9.*
