# Go-AI Framework: Panduan Arsitektur & Aturan Coding (Developer & AI Agent)

Dokumen ini ditujukan bagi **Tim Developer** dan **AI Coding Agent** (seperti Antigravity, Cursor, Copilot, ChatGPT) agar dapat mengembangkan fitur baru secara paralel tanpa merusak bagian inti (*Core Framework*) maupun modul aplikasi lainnya.

---

## 🛑 1. Aturan Emas: Pemisahan Core vs Apps (Core vs Apps Separation)

Struktur repositori `go-ai` terbagi secara tegas menjadi dua zona utama:

```
go-ai/
├── [CORE FRAMEWORK — DILARANG DIBUAT/DIUBAH]
│   ├── main.go           ← Server orchestrator
│   ├── config/           ← Env configuration reader
│   ├── database/         ← DB connection pool (SQLite/MySQL)
│   ├── logger/           ← Structured logger (slog)
│   ├── middleware/       ← Rate limiter, panic recovery, CORS
│   ├── render/           ← Template & JSON rendering engine
│   ├── router/           ← Enhanced ServeMux wrapper (Go 1.22+)
│   ├── security/         ← Security headers
│   └── static/           ← Static assets server
│
└── [APPS ZONE — WILAYAH KERJA DEVELOPER & AI]
    ├── apps/bootstrap/   ← ⭐ Titik tunggal pendaftaran modul ke Core
    ├── apps/ui/          ← Master HTML templates & global CSS/JS
    ├── apps/auth/        ← Modul Otentikasi & Sesi
    ├── apps/shortener/   ← Modul URL Shortener
    ├── apps/custom/      ← Modul Dynamic Custom Fields
    └── apps/<nama_modul>/ ← Folder modul baru Anda
```

### Aturan Zona:
* **CORE ZONE (DILARANG SENTUH)**: File di root (selain `/apps`, `/docs`, dan `.agents`) milik Core Framework. **Dilarang** mengubah `main.go`, `router/`, `render/`, dll.
* **APPS ZONE (AREA BEBAS)**: Seluruh logika bisnis, fitur baru, dan tampilan UI **wajib** dibuat di dalam folder `/apps/`.

---

## 📦 2. Pola Terisolasi Setiap Modul (`/apps/<nama_modul>/`)

Setiap modul di dalam `/apps/` berdiri secara mandiri (Modular Monolith). Berikut adalah konvensi struktur filenya:

### File Standar Modul
1. **`models.go`**:
   - Mendefinisikan `struct` data modul.
   - Menghadapi database menggunakan SQL murni (`database/sql`).
   - Menyediakan fungsi `MigrateTable(db *sql.DB) error` yang bertanggung jawab atas tabelnya sendiri (`CREATE TABLE IF NOT EXISTS`).
2. **`routes.go`**:
   - Menyediakan fungsi `RegisterRoutes(r *router.Router, tmpl *render.Engine, log logger.Logger, db *sql.DB)`.
   - Mengelola pemetaan Endpoint HTTP & *handler functions*.
   - Menangani pengembalian JSON atau HTML Partials (HTMX).
3. **`middleware.go`** *(Opsional)*:
   - Jika modul membutuhkan validasi/proteksi khusus (misal `auth.RequireAuth(db)`).
4. **`utils.go` / `helpers.go`** *(Opsional)*:
   - Fungsi pembantu internal modul (misal parsing string, generator ID acak).
5. **`*_test.go`** *(Opsional)*:
   - Pengujian otomatis (*unit testing*) menggunakan `testing` dan SQLite `:memory:`.

---

## 🔌 3. Cara Mendaftarkan Modul Baru (`apps/bootstrap/bootstrap.go`)

Developer dan AI Agent **TIDAK BUKAN** mengimpor modul di `main.go`. Sebagai gantinya, pendaftaran dilakukan di `apps/bootstrap/bootstrap.go`:

```go
package bootstrap

import (
	"database/sql"
	"go-ai/apps/auth"
	"go-ai/apps/custom"
	"go-ai/apps/shortener"
	"go-ai/apps/modulbaru" // 1. Import modul baru
	...
)

func Register(r *router.Router, tmplEngine *render.Engine, log logger.Logger, db *sql.DB) {
	// 2. Jalankan migrasi tabel modul
	if err := modulbaru.MigrateTable(db); err != nil {
		log.Error("Failed to migrate modulbaru table", "error", err)
	}

	// 3. Daftarkan rute modul ke router
	modulbaru.RegisterRoutes(r, tmplEngine, log, db)
}
```

---

## ⚡ 4. Arsitektur Dynamic Extension (No-Redeploy Architecture)

Untuk fitur yang membutuhkan kustomisasi field tanpa perlu kompilasi/deploy ulang binary (seperti pada modul `apps/custom`):
* **Core Tables (Tetap)**: Menyimpan kolom-kolom inti yang pasti.
* **Tabel Skema (`custom_fields`)**: Menyimpan definisi input kustom baru (`name`, `label`, `field_type`, `options`, `required`).
* **Kolom `metadata` (JSON)**: Menyimpan nilai-nilai input kustom sebagai JSON String di database.
* **UI Dynamic Rendering**: Form dan tabel membaca metadata `custom_fields` dan merender elemen HTML secara otomatis via loop `{{ range .CustomFields }}`.

---

## 🌐 5. Antarmuka & HTMX Integration (`apps/ui/`)

1. **Templates**: Disimpan di `apps/ui/templates/<nama_modul>/` atau `apps/ui/templates/`.
2. **Partial Swapping**: Saat HTMX melakukan *request* (`HX-Request == true`), *handler* mengembalikan potongan HTML saja (misal `list_partial.html`). Saat dibuka via browser biasa, *handler* mengembalikan *full page* (`dashboard.html`).
3. **Event Triggers**: Manfaatkan header `HX-Trigger` untuk menyegarkan tampilan UI antar elemen secara otomatis tanpa penguncian status (*stateless*).
4. **Styling**: Gunakan kelas CSS standar di `/static/style.css` (berbasis variabel native CSS, *Dark Theme*, *Glassmorphism*). Dilarang mengimpor Tailwind/Bootstrap kecuali diminta.

---

## 🔒 6. Prinsip Kode & Batasan Teknis

1. **Pure Go Standard Library**: Prioritaskan paket bawaan Go (`net/http`, `database/sql`, `crypto/`, `encoding/json`). Dilarang menginstal *third-party framework* berat seperti Gin, Echo, atau GORM.
2. **Dependency Injection**: Jangan membuat variabel global `db` atau `log` di dalam modul. Selalu alirkan *dependency* melalui argumen fungsi `RegisterRoutes`.
3. **Penyimpanan Password**: Gunakan `golang.org/x/crypto/bcrypt` dengan cost factor minimal 12. **JANGAN** gunakan SHA-256/MD5 untuk hashing password — hash ini terlalu cepat dan rentan brute-force. bcrypt memiliki *cost factor* yang membuat setiap percobaan memakan waktu, melindungi dari serangan GPU massal.
4. **Keamanan Sesi Cookie**: Kunci sesi (*session token*) disimpan di SQLite dan dikirim ke browser via *HTTP-Only Cookie* (`SameSite=Lax`).

---

## 🔨 7. Kompilasi & Pengujian

### Perintah Build Rilis (Optimized Binary)
```powershell
# Windows (.exe) - Disimpan di .build\win\
go build -ldflags="-s -w" -trimpath -o .build\win\go-ai.exe main.go

# Linux - Disimpan di .build/linux/
$env:GOOS="linux"; go build -ldflags="-s -w" -trimpath -o .build/linux/go-ai-linux main.go
```

### Menjalankan Server
```powershell
.\.build\win\go-ai.exe
```

---

Dengan mengikuti pedoman di dokumen ini, seluruh **Tim Developer** dan **AI Agent** dapat memproduksi kode yang konsisten, aman, aman dari bentrok antar-modul, serta memenuhi standar *production-grade*.
