package main

import (
	"context"
	"go-ai/apps/bootstrap"
	"go-ai/config"
	"go-ai/database"
	"go-ai/logger"
	"go-ai/middleware"
	"go-ai/render"
	"go-ai/router"
	"go-ai/security"
	"go-ai/static"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	// 1. Load konfigurasi dari environment variables
	cfg := config.Load()
	log := logger.New(cfg.IsProd())
	log.Info("Starting framework", "mode", cfg.Env)

	// 2. Inisialisasi Database
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "data.db"
	}

	maxConns := getEnvInt("DB_MAX_CONNS", 100)
	maxIdle := getEnvInt("DB_MAX_IDLE", 10)
	connLifetimeMin := getEnvInt("DB_CONN_LIFETIME_MIN", 5)

	db, err := database.Connect(database.Config{
		Driver:          driver,
		DSN:             dsn,
		MaxConns:        maxConns,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: time.Duration(connLifetimeMin) * time.Minute,
	})
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// 3. Ambil aset UI dari Apps (tanpa import langsung ke apps/ui)
	uiAssets := bootstrap.GetUI()

	// 4. Inisialisasi Template Engine — Core tidak tahu di mana folder UI berada
	tmplEngine, err := render.NewEngine("templates", uiAssets.LocalDevDir, uiAssets.TemplatesFS, cfg.IsProd())
	if err != nil {
		log.Fatal("Failed to initialize templates", "error", err)
	}

	// 5. Buat Router
	r := router.New()

	// 6. Static Files Route — Core tidak tahu di mana folder UI berada
	r.Handle("GET /static/", http.StripPrefix("/static/",
		static.Serve("public", uiAssets.LocalDevDir, uiAssets.PublicFS, cfg.IsProd()),
	))

	// 7. Health Check — Core-level endpoint, selalu tersedia
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}` + "\n"))
	})

	// 8. Daftarkan semua Apps (Migrasi DB + Routes)
	// Untuk menambah App baru: EDIT apps/bootstrap/bootstrap.go, BUKAN file ini!
	bootstrap.Register(r, tmplEngine, log, db.DB)

	// 9. Susun rantai Middleware
	middlewares := []middleware.Middleware{
		middleware.RequestID,
		middleware.AccessLog(log),
		middleware.Recover(log),
		middleware.Gzip(),
		security.Headers,
		security.CORS,
	}
	// Rate limiting hanya aktif di Production Mode
	// Gunakan root context (ctx) agar goroutine cleanup rate limiter bisa dihentikan saat shutdown
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	if cfg.IsProd() {
		middlewares = append([]middleware.Middleware{middleware.RateLimiter(ctx, rate.Limit(10), 30)}, middlewares...)
	}
	handler := middleware.Chain(r, middlewares...)

	// 10. Konfigurasi dan jalankan HTTP Server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
		// BaseContext propagates the context to all incoming requests,
		// ensuring long-running connections (like WebSockets) can detect shutdown.
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 5 * time.Second,  // Prevent Slowloris attack
		ReadTimeout:       30 * time.Second, // Total body read timeout — prevents slow POST attacks
		WriteTimeout:      60 * time.Second, // Handler + response write timeout — prevents hung handlers
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB — prevent large header attacks
	}

	go func() {
		log.Info("Server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
		}
	}()

	// 11. Graceful Shutdown — tunggu sinyal Ctrl+C / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Cancel the root context immediately so WebSockets and long-running handlers know the server is shutting down.
	// This prevents srv.Shutdown from blocking for the full 5 seconds waiting for idle WebSockets.
	cancelCtx()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server exiting")
}

// getEnvInt reads an integer environment variable, returning defaultVal if not set or invalid.
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
