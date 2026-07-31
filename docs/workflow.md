# Go-AI Framework: Developer Workflow Guide

Selamat datang di tim *engineering* Go-AI! Proyek ini dibangun dengan standar **Enterprise Production Grade**, memanfaatkan *100% Go Standard Library* tanpa ketergantungan pada *framework* pihak ketiga seperti Gin atau Echo.

Dokumen ini adalah panduan wajib bagi seluruh *developer* (Frontend, Backend, maupun AI Agent) untuk memahami alur kerja dan pembagian tugas agar pengembangan berjalan paralel tanpa merusak inti sistem.

---

## 🏗️ Zona Kerja (Safe Work Lanes)

Kami membagi struktur folder menjadi dua zona yang tegas:

```
CORE FRAMEWORK (Root) — JANGAN DISENTUH ⛔
APPS (apps/)           — ZONA KERJA DEVELOPER ✅
```

---

### ⛔ Zona Inti (Core Engine — ROOT)

Folder-folder berikut adalah mesin utama penggerak *framework*. **Dilarang keras mengubah folder ini** kecuali ada kebutuhan kritis dari Lead Engineer:

| Folder/File | Fungsi |
|---|---|
| `main.go` | Orkestrator server — hanya memanggil `bootstrap.Register()` |
| `config/` | Pembaca environment variables (`APP_ENV`, `APP_PORT`, dll) |
| `database/` | Connection Pool multi-driver (MySQL + SQLite Pure Go) |
| `logger/` | Structured JSON logging via `log/slog` Go 1.21+ |
| `middleware/` | Recovery Panic, Rate Limiter (Token Bucket), Chain helper |
| `render/` | Template Engine HTML + JSON renderer + BindJSON/BindQuery |
| `router/` | Wrapper `http.ServeMux` Go 1.22 (dengan method helpers) |
| `security/` | Security Headers + CORS middleware |
| `static/` | Static file server (disk hot-reload + embed fallback) |
| `cache/` | In-memory TTL cache dengan auto garbage collector |
| `websocket/` | WebSocket handler wrapper |

---

### ✅ Zona Aplikasi (Apps — `/apps`)

Semua pekerjaan developer ada di sini. Folder `/apps` terbagi menjadi:

#### 1. `apps/bootstrap/` — ⭐ Penghubung Core & Apps
- **Hanya file ini yang boleh diubah saat menambah atau menghapus App baru.**
- `bootstrap.go` mendaftarkan semua migrasi DB dan routes ke Core.
- Core Framework (`main.go`) tidak perlu disentuh sama sekali.

#### 2. `apps/ui/` — Frontend (HTML + CSS + JS)
- `apps/ui/templates/`: Folder penyimpan file HTML (mendukung subfolder per-fitur).
- `apps/ui/public/`: Folder CSS, JS, gambar, dan aset statis lainnya.
- `apps/ui/embed.go`: Pembungkus `go:embed` — **jangan ubah kecuali ada folder baru**.
- Seluruh file di sini akan di-*embed* ke dalam binary saat dikompilasi.

#### 3. `apps/namafitur/` — Domain/Fitur Baru
- Setiap domain bisnis baru (contacts, payments, products, dll) harus memiliki foldernya sendiri.
- Minimal 2 file per App:
  - `models.go` — struct + migrasi + query SQL
  - `routes.go` — fungsi `RegisterRoutes(...)` berisi handler HTTP/HTMX

**Aturan:**
1. Jangan pernah saling mengimpor antar App (agar tetap *decoupled*).
2. App hanya boleh mengimpor paket Core (`go-ai/render`, `go-ai/router`, dll).

---

## 🚀 Alur Menambahkan Fitur Baru

Semua pekerjaan ada di dalam `/apps`. Ikuti 3 langkah berikut:

### Langkah 1: Buat Folder App Baru

```
apps/
  products/
    models.go   ← struct + migrasi + query SQLite
    routes.go   ← handler HTTP/HTMX
```

**Contoh `models.go`:**
```go
package products

import (
    "context"
    "database/sql"
)

type Product struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Price float64 `json:"price"`
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
    rows, err := db.QueryContext(ctx, "SELECT id, name, price FROM products")
    // ... scan rows ...
}
```

**Contoh `routes.go`:**
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
    // REST API JSON
    r.Get("/api/products", func(w http.ResponseWriter, req *http.Request) {
        items, err := GetAll(req.Context(), db)
        if err != nil {
            log.Error("Failed to get products", "error", err)
            http.Error(w, "Internal Server Error", 500)
            return
        }
        render.JSON(w, http.StatusOK, items)
    })

    // HTMX Full Page (opsional)
    r.Get("/products", func(w http.ResponseWriter, req *http.Request) {
        items, _ := GetAll(req.Context(), db)
        tmpl.Render(w, http.StatusOK, "products.html", map[string]interface{}{
            "Products": items,
        })
    })
}
```

### Langkah 2: Daftarkan ke Bootstrap

Buka **`apps/bootstrap/bootstrap.go`** — satu-satunya file yang perlu diubah:

```go
import "go-ai/apps/products"

func Register(r *router.Router, tmplEngine *render.Engine, log logger.Logger, db *sql.DB) {
    // Migrasi DB (tambahkan baris baru di sini)
    products.MigrateTable(db)

    // Routes (tambahkan baris baru di sini)
    products.RegisterRoutes(r, tmplEngine, log, db)
}
```

> ✅ **`main.go` dan semua Core tidak perlu disentuh sama sekali.**

### Langkah 3: Buat UI (Jika HTMX)

- Buat file HTML di `apps/ui/templates/products/list.html`.
- Edit CSS di `apps/ui/public/style.css` jika perlu styling baru.
- Panggil dari `routes.go`: `tmpl.Render(w, 200, "list.html", data)`.

---

## 🔨 Perintah Penting

### Build Binary

```powershell
# Windows (.exe) — simpan ke .build/win/
go build -ldflags="-s -w" -trimpath -o .build\win\go-ai.exe main.go

# Linux (VPS) — simpan ke .build/linux/
$env:GOOS="linux"; go build -ldflags="-s -w" -trimpath -o .build/linux/go-ai-linux main.go
```

### Menjalankan untuk Testing

```powershell
# Jalankan binary (WAJIB — tidak perlu approval Windows Firewall)
.\.build\win\go-ai.exe

# Mode Production (aktifkan Rate Limiter + embed mode)
$env:APP_ENV="production"; .\.build\win\go-ai.exe
```

> ⚠️ **Jangan gunakan `go run main.go` untuk testing** — memerlukan approval Windows Firewall setiap kali.

### Testing

```bash
go test -v ./...
```

---

## 🧰 Menggunakan Cache Bawaan

Framework menyediakan `cache.Cache` untuk menyimpan data sementara di memori:

```go
import "go-ai/cache"

// Buat cache dengan cleanup setiap 5 menit
c := cache.New(5 * time.Minute)

// Simpan dengan TTL 10 menit
c.Set("key", someData, 10*time.Minute)

// Ambil
if val, ok := c.Get("key"); ok {
    data := val.(MyStruct)
}
```

---

## 📋 Konvensi Kode

1. **Nama package** = nama folder (lowercase, tanpa underscore).
2. **Nama file** = deskriptif (`models.go`, `routes.go`, `middleware.go`).
3. **Error handling** = selalu gunakan `log.Error(...)`, jangan `fmt.Println`.
4. **Database** = selalu gunakan `Context` di semua query (`db.QueryContext(ctx, ...)`).
5. **Tidak ada global state** = semua dependency di-inject melalui fungsi.
