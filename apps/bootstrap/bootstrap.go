// Package bootstrap adalah satu-satunya tempat yang mendaftarkan semua Apps ke Core Framework.
// Hanya file ini yang boleh diubah saat menambah atau menghapus App baru.
// File-file Core Framework (main.go, render, static, dll.) TIDAK BOLEH disentuh.
package bootstrap

import (
	"database/sql"
	"go-ai/apps/dev"
	"go-ai/apps/ecommerce"
	"go-ai/apps/ui"
	"go-ai/logger"
	"go-ai/render"
	"go-ai/router"
	"io/fs"
)

// UIAssets berisi embed.FS untuk Templates dan Public (CSS/JS).
// Core Framework menggunakannya melalui interface, bukan import langsung.
type UIAssets struct {
	TemplatesFS fs.FS
	PublicFS    fs.FS
	LocalDevDir string // Path ke folder UI saat Dev Mode (hot reload)
}

// GetUI mengembalikan UIAssets dari apps/ui.
// Dipanggil oleh main.go — satu-satunya titik kontak antara Core dan UI.
func GetUI() UIAssets {
	return UIAssets{
		TemplatesFS: ui.TemplatesFS,
		PublicFS:    ui.PublicFS,
		LocalDevDir: "apps/ui",
	}
}

// Register mendaftarkan semua App Migrations dan Routes ke Core.
// Untuk menambah App baru: tambahkan di sini, TIDAK PERLU menyentuh main.go.
func Register(r *router.Router, tmplEngine *render.Engine, log logger.Logger, db *sql.DB) {
	// --- Migrasi Database ---
	if err := ecommerce.MigrateTable(db); err != nil {
		log.Error("Failed to migrate ecommerce tables", "error", err)
	}

	// --- Registrasi Routes ---
	dev.RegisterRoutes(r, tmplEngine, log, db)
	ecommerce.RegisterRoutes(r, tmplEngine, log, db)
}
