# TaskFlow API — CI/CD Pipeline

> **Mata Kuliah**: Operasional Pengembang (DevOps) — Pertemuan 9
> **Tool**: GitHub Actions + GHCR (GitHub Container Registry)

## 📋 Deskripsi

TaskFlow API adalah backend Go untuk aplikasi manajemen proyek. Repository ini mengimplementasikan pipeline CI/CD otomatis menggunakan GitHub Actions untuk memastikan kualitas kode dan deployment yang andal.

## 🏗️ Arsitektur Pipeline

```
Push → Lint (go vet) → Test (race + coverage ≥75%) → Integration Test (PostgreSQL)
                                      ↓
                    Docker Build + Push ke GHCR (tag: sha-xxx, latest)
                                      ↓
                         Smoke Test (/health, /api/v1/stats)
                                      ↓
                              Tag :stable + Notifikasi Telegram
```

## 🚀 Quick Start

### Menjalankan Lokal

```bash
# 1. Clone repository
git clone https://github.com/v0rein/taskflow-api.git
cd taskflow-api

# 2. Jalankan dengan Docker Compose (direkomendasikan)
cd pbl-taskflow-go
docker compose up -d
curl http://localhost:8080/health

# 3. Atau tanpa Docker (data tidak persisten)
cd pbl-taskflow-go
go build -o bin/taskflow-api ./cmd/server
./bin/taskflow-api
```

### Menjalankan Test

```bash
cd pbl-taskflow-go

# Unit test + race detector
go test -race ./... -v

# Coverage report
go test ./... -coverprofile=cov.out && go tool cover -func=cov.out

# Integration test (butuh PostgreSQL)
export DATABASE_URL=postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable
go test -tags=integration -race ./... -v
```

## 🔐 GitHub Secrets yang Diperlukan

| Secret                | Deskripsi                           |
| --------------------- | ----------------------------------- |
| `TELEGRAM_BOT_TOKEN`  | Token bot Telegram untuk notifikasi |
| `TELEGRAM_CHAT_ID`    | Chat ID tujuan notifikasi           |

> `GITHUB_TOKEN` sudah otomatis tersedia, digunakan untuk push ke GHCR.

## 📦 Docker Image

Image tersedia di GitHub Container Registry:

```bash
# Pull image terbaru
docker pull ghcr.io/v0rein/taskflow-api:latest

# Pull versi spesifik
docker pull ghcr.io/v0rein/taskflow-api:sha-<commit>

# Pull versi stabil (lolos semua test + smoke test)
docker pull ghcr.io/v0rein/taskflow-api:stable
```

## 🔄 Rollback

```bash
# Rollback ke versi stabil terakhir
make rollback ROLLBACK_TAG=stable

# Rollback ke commit spesifik
make rollback ROLLBACK_TAG=sha-a3f2c1d
```

Lihat [ROLLBACK_PROCEDURE.md](ROLLBACK_PROCEDURE.md) untuk prosedur lengkap.

## 📂 Struktur Repository

```
.
├── .github/workflows/ci.yml    ← Pipeline CI/CD
├── pbl-taskflow-go/             ← Source code Go
│   ├── cmd/server/main.go       ← Entry point
│   ├── internal/
│   │   ├── handler/             ← HTTP handlers
│   │   ├── model/               ← Data models
│   │   ├── repository/          ← Database layer
│   │   ├── service/             ← Business logic
│   │   └── validator/           ← Input validation
│   ├── Dockerfile               ← Multi-stage build (scratch)
│   ├── Makefile                 ← Build & deploy targets
│   └── docker-compose.yml       ← Local development stack
├── LAPORAN_PROYEK.md            ← Laporan lengkap
├── ROLLBACK_PROCEDURE.md        ← Prosedur rollback
└── README.md                    ← File ini
```

## 📄 API Endpoints

| Method   | Path                 | Deskripsi             |
| -------- | -------------------- | --------------------- |
| `GET`    | `/health`            | Health check          |
| `GET`    | `/api/v1/tasks`      | List semua task       |
| `POST`   | `/api/v1/tasks`      | Buat task baru        |
| `GET`    | `/api/v1/tasks/{id}` | Ambil task by ID      |
| `PUT`    | `/api/v1/tasks/{id}` | Update task           |
| `DELETE` | `/api/v1/tasks/{id}` | Hapus task            |
| `GET`    | `/api/v1/stats`      | Statistik task        |

---

*Dibuat untuk tugas Operasional Pengembang — Pertemuan 9*
