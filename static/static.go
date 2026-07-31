package static

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// Serve creates a handler for static files.
// localDevBase: path ke folder UI saat dev mode, dikirim dari luar (contoh: "apps/ui").
// dir: subfolder public di dalam localDevBase dan embedFS (contoh: "public").
// embedFS: embed.FS dari folder apps/ui.
// prodMode: jika true, serve dari embedFS.
func Serve(dir string, localDevBase string, embedFS fs.FS, prodMode bool) http.Handler {
	if prodMode {
		// Serve from embed.FS (binary)
		subFS, err := fs.Sub(embedFS, dir)
		if err != nil {
			return http.FileServer(http.FS(embedFS))
		}
		return http.FileServer(http.FS(subFS))
	}

	// Serve from disk (Dev Mode) for hot reloading
	localDir := filepath.Join(localDevBase, dir)
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		// Fallback to embedFS if local dir does not exist
		// (e.g. running binary from a different directory)
		subFS, err := fs.Sub(embedFS, dir)
		if err != nil {
			return http.FileServer(http.FS(embedFS))
		}
		return http.FileServer(http.FS(subFS))
	}
	return http.FileServer(http.Dir(localDir))
}
