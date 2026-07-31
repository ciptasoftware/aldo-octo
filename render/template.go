package render

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Engine holds the parsed templates
type Engine struct {
	templates   *template.Template
	dir         string
	localDevDir string // Path ke folder UI saat Dev Mode, dikirim dari luar (bukan hardcoded)
	embedFS     fs.FS
	mu          sync.RWMutex
	prodMode    bool
}

// NewEngine creates a new template engine.
// localDevDir: path ke folder yang berisi templates saat dev mode (contoh: "apps/ui").
// embedFS: embed.FS dari folder apps/ui (dikirim dari bootstrap, bukan dari main.go langsung).
// dir: subfolder templates di dalam embedFS (contoh: "templates").
// prodMode: jika true, templates di-parse dari embedFS dan di-cache.
func NewEngine(dir string, localDevDir string, embedFS fs.FS, prodMode bool) (*Engine, error) {
	e := &Engine{
		dir:         dir,
		localDevDir: localDevDir,
		embedFS:     embedFS,
		prodMode:    prodMode,
	}

	if prodMode {
		if err := e.parseAll(); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (e *Engine) parseAll() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl := template.New("")
	err := fs.WalkDir(e.embedFS, e.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".html" {
			_, err = tmpl.ParseFS(e.embedFS, path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	e.templates = tmpl
	return nil
}

// Render executes a template by name and writes to the response.
// Templates are rendered to a buffer first to prevent sending partial/corrupt
// HTML to clients when a template error occurs.
func (e *Engine) Render(w http.ResponseWriter, status int, name string, data interface{}) {
	var tmpl *template.Template
	var err error

	if !e.prodMode {
		// In dev mode, reparse templates from disk for hot reloading
		localDir := filepath.Join(e.localDevDir, e.dir)
		if _, errStat := os.Stat(localDir); errStat == nil {
			tmpl = template.New("")
			tmpl, err = tmpl.ParseGlob(filepath.Join(localDir, "*.html"))
			if err == nil {
				// Ignore error for subdirectories - they may not exist
				tmpl.ParseGlob(filepath.Join(localDir, "*", "*.html"))
			}
		} else {
			// Fallback to embedFS if local dir does not exist
			tmpl = template.New("")
			err = fs.WalkDir(e.embedFS, e.dir, func(path string, d fs.DirEntry, errWalk error) error {
				if errWalk != nil {
					return errWalk
				}
				if !d.IsDir() && filepath.Ext(path) == ".html" {
					_, errParse := tmpl.ParseFS(e.embedFS, path)
					if errParse != nil {
						return errParse
					}
				}
				return nil
			})
		}

		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		e.mu.RLock()
		tmpl = e.templates
		e.mu.RUnlock()
	}

	// Render to buffer first — prevents partial/corrupt HTML on error
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("template execution error", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}
