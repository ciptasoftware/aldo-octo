# Go-AI Web Framework

Sebuah *web framework* backend dan fullstack **berperforma tinggi, ringan, dan 100% murni Standard Library Go** — tanpa ketergantungan pada framework pihak ketiga seperti Gin atau Echo.

Dibangun dengan filosofi **minimalis, modular, dan benar-benar Production Grade**. Cocok untuk membangun REST API maupun aplikasi web *Server-Side Rendered (SSR)* menggunakan **HOWL Stack** (HTMX + Go + WebSocket).

---

## ✨ Fitur Utama

| Fitur | Keterangan |
|---|---|
| 🚀 **Super Cepat & Ringan** | Murni `net/http` Go 1.22+ (`http.ServeMux` dengan *Enhanced Routing*) |
| 🛡️ **Security Headers** | `X-Frame-Options`, `X-Content-Type-Options`, `HSTS`, `XSS-Protection` otomatis |
| 🔒 **CORS Middleware** | Mendukung *cross-origin request* dengan konfigurasi terpusat |
| ⚡ **Rate Limiter** | Token Bucket berbasis IP, aktif otomatis di Production Mode, anti-DDoS |
| 💥 **Anti-Crash (Recovery)** | *Panic Recovery Middleware* — server tidak akan mati akibat bug di App Layer |
| 🔄 **Graceful Shutdown** | Menyelesaikan request aktif sebelum server benar-benar mati |
| 📋 **Structured Logging** | JSON log terstruktur via `log/slog` (Go 1.21+), berbeda level untuk Dev/Prod |
| 🗄️ **Database Connection Pool** | Siap MySQL/MariaDB + SQLite (Pure Go), konfigurasi max conn & lifetime |
| 📡 **WebSocket** | Komunikasi real-time dua arah via `golang.org/x/net/websocket` |
| 🧠 **In-Memory Cache** | Cache TTL bawaan dengan *auto garbage collector* goroutine |
| 🎨 **HTMX + go:embed** | Template HTML + aset statis (CSS/JS) di-*embed* ke dalam satu binary |

---

## 🏗️ Arsitektur

Framework ini menggunakan pola **Core + Apps** yang tegas memisahkan mesin inti dari logika bisnis.

```
go-ai/
│
├── [CORE FRAMEWORK — JANGAN DISENTUH]
│   ├── main.go           ← Orkestrator server (hanya panggil bootstrap)
│   ├── config/           ← Pembaca environment variables
│   ├── database/         ← Connection Pool (MySQL + SQLite)
│   ├── logger/           ← Structured JSON Logger (slog)
│   ├── middleware/        ← Recover, Rate Limiter, Chain
│   ├── render/           ← Template Engine HTML + JSON renderer
│   ├── router/           ← Wrapper http.ServeMux Go 1.22
│   ├── security/         ← Security Headers + CORS middleware
│   ├── static/           ← Static file server (disk + embed fallback)
│   ├── cache/            ← In-memory TTL cache
│   └── websocket/        ← WebSocket handler wrapper
│
└── [APPS — ZONA KERJA DEVELOPER/AI]
    ├── apps/bootstrap/   ← ⭐ Titik pendaftaran semua Apps ke Core
    ├── apps/ui/          ← Frontend (HTML templates + CSS/JS)
    ├── apps/dev/         ← Dev & testing routes (API echo, panic test, WS)
    └── apps/<namaapp>/   ← Folder App baru Anda (ikuti panduan di bawah)
```

### Aturan Emas
- ✅ **Developer dan AI Agent HANYA boleh bekerja di folder `/apps`**
- ⛔ **Folder Core di root TIDAK BOLEH disentuh** kecuali ada kebutuhan kritis dari Lead Engineer
- 📌 **Untuk menambah App baru**: buat folder di `/apps/namaapp/`, lalu daftarkan di `apps/bootstrap/bootstrap.go`

---

## 🚀 Quick Start

### Prasyarat
- Go 1.22 atau lebih baru
- Git

### Clone & Jalankan

```bash
git clone https://codeberg.org/dodolah/go-ai.git
cd go-ai

# Build binary Windows
go build -ldflags="-s -w" -trimpath -o .build\win\go-ai.exe main.go

# Jalankan
.\.build\win\go-ai.exe
```

Buka browser di `http://localhost:8080`.

---

## 📦 Menambahkan Fitur/App Baru

Seluruh pekerjaan developer cukup di dalam `/apps`. Ikuti 3 langkah ini:

### Langkah 1: Buat Folder App

```
apps/
  products/
    models.go    ← struct + query SQL
    routes.go    ← handler HTTP/HTMX
```

**`apps/products/models.go`**
```go
package products

import (
    "context"
    "database/sql"
)

type Product struct {
    ID    int
    Name  string
    Price float64
}

func MigrateTable(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS products (
        id    INTEGER PRIMARY KEY AUTOINCREMENT,
        name  TEXT NOT NULL,
        price REAL NOT NULL
    )`)
    return err
}

func GetAll(ctx context.Context, db *sql.DB) ([]Product, error) {
    // ... query logic
}
```

**`apps/products/routes.go`**
```go
package products

import (
    "database/sql"
    "go-ai/logger"
    "go-ai/render"
    "go-ai/router"
    "net/http"
)

func RegisterRoutes(r *router.Router, tmpl *render.Engine, log logger.Logger, db *sql.DB) {
    r.Get("/products", func(w http.ResponseWriter, req *http.Request) {
        items, _ := GetAll(req.Context(), db)
        render.JSON(w, http.StatusOK, items)
    })
}
```

### Langkah 2: Daftarkan di Bootstrap

Buka **`apps/bootstrap/bootstrap.go`** — satu-satunya file yang perlu diubah:

```go
import "go-ai/apps/products"

func Register(...) {
    // Tambahkan migrasi
    products.MigrateTable(db)

    // Tambahkan routes
    products.RegisterRoutes(r, tmplEngine, log, db)
}
```

> ✅ **Selesai. `main.go` dan seluruh Core tidak perlu disentuh.**

### Langkah 3: Buat UI (Opsional)

Buat file HTML di `apps/ui/templates/namafitur/halaman.html` dan panggil dari `routes.go`:

```go
tmplEngine.Render(w, http.StatusOK, "halaman.html", data)
```

---

## ⚙️ Konfigurasi Environment

Framework membaca environment variable standar OS (tidak memerlukan library dotenv):

| Variable | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | Set `production` untuk aktifkan Rate Limiter & embed mode |
| `APP_PORT` | `8080` | Port server |
| `DB_DRIVER` | `sqlite` | Driver database (`sqlite` atau `mysql`) |
| `DB_DSN` | `data.db` | Data Source Name |
| `DB_MAX_CONNS` | `100` | Maksimum koneksi database yang dibuka |
| `DB_MAX_IDLE` | `10` | Maksimum koneksi idle di pool |
| `DB_CONN_LIFETIME_MIN` | `5` | Lifetime koneksi dalam menit |
| `CORS_ORIGINS` | `*` | Allowed CORS origins (comma-separated, contoh: `https://example.com,https://app.example.com`) |

**Contoh Production di Linux:**
```bash
export APP_ENV=production
export APP_PORT=80
export DB_DRIVER=sqlite
export DB_DSN=/var/data/go-ai.db
./go-ai-linux
```

---

## 🔨 Perintah Build

Sesuai aturan `.agents/rules.txt`:

```powershell
# Windows (.exe) — simpan ke .build/win/
go build -ldflags="-s -w" -trimpath -o .build\win\go-ai.exe main.go

# Linux (VPS/Server) — simpan ke .build/linux/
$env:GOOS="linux"; go build -ldflags="-s -w" -trimpath -o .build/linux/go-ai-linux main.go
```

> Binary hasil build sudah mengandung seluruh file HTML, CSS, dan driver SQLite di dalamnya. **Kirim satu file saja ke server.**

---

## 🧪 Testing

```bash
# Jalankan binary untuk testing (tidak memerlukan approval Windows Firewall)
.\.build\win\go-ai.exe

# Mode Production
$env:APP_ENV="production"; .\.build\win\go-ai.exe

# Unit & Integration test
go test -v ./...
```

---

## 📚 Dokumentasi Lengkap

Lihat [`docs/workflow.md`](docs/workflow.md) untuk panduan lengkap alur kerja developer dan AI Agent.

---

## 📋 Dependensi Eksternal

Framework ini sengaja meminimalkan dependensi pihak ketiga:

| Package | Versi | Fungsi |
|---|---|---|
| `golang.org/x/net` | v0.57.0 | WebSocket |
| `golang.org/x/time` | v0.15.0 | Rate Limiter (Token Bucket) |
| `github.com/go-sql-driver/mysql` | v1.10.0 | Driver MySQL/MariaDB |
| `modernc.org/sqlite` | v1.54.0 | Driver SQLite (Pure Go, tanpa CGO) |

---

## 📄 Lisensi

MIT License — bebas digunakan untuk keperluan komersial maupun pribadi.
