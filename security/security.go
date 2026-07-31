package security

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// CORS is a middleware that adds CORS headers to the response.
// Reads allowed origins from the CORS_ORIGINS environment variable (comma-separated).
// Defaults to "*" if not set (permissive — suitable for development only).
// In production, set CORS_ORIGINS to specific domains (e.g., "https://example.com,https://app.example.com").
func CORS(next http.Handler) http.Handler {
	allowedRaw := os.Getenv("CORS_ORIGINS")
	if allowedRaw == "" {
		allowedRaw = "*"
		if os.Getenv("APP_ENV") == "production" {
			slog.Warn("CORS_ORIGINS not set — defaulting to '*' (allow all origins). Set CORS_ORIGINS for production security.")
		}
	}

	// Pre-compute allowed origins set for O(1) lookup
	var allowAll bool
	allowedOrigins := make(map[string]bool)
	if allowedRaw == "*" {
		allowAll = true
	} else {
		for _, origin := range strings.Split(allowedRaw, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins[origin] = true
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Headers adds standard production-grade security headers to every response.
// Protects against Clickjacking, MIME sniffing, XSS, and enforces HTTPS.
func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent Clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Prevent MIME-sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Enable XSS protection in legacy browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Strict Transport Security (HSTS) — force HTTPS for 1 year
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// Content Security Policy — restrict resource loading to same origin by default
		// App developers can override this per-route for HTMX/CDN use cases
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self' ws: wss:")
		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Permissions Policy — disable dangerous browser APIs
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
