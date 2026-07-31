package security

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// --- Security Headers Tests ---

func TestSecurityHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Headers(handler)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	tests := []struct {
		header   string
		expected string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		got := w.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("header %s: expected %q, got %q", tt.header, tt.expected, got)
		}
	}

	// HSTS should be present
	hsts := w.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("expected Strict-Transport-Security header")
	}

	// CSP should be present
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header")
	}

	// Permissions-Policy should be present
	pp := w.Header().Get("Permissions-Policy")
	if pp == "" {
		t.Error("expected Permissions-Policy header")
	}
}

// --- CORS Tests ---

func TestCORSWildcard(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")
	os.Unsetenv("APP_ENV")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao != "*" {
		t.Errorf("expected ACAO=%q, got %q", "*", acao)
	}
}

func TestCORSSpecificOrigin(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "https://example.com,https://app.example.com")
	defer os.Unsetenv("CORS_ORIGINS")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS(handler)

	// Allowed origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao != "https://example.com" {
		t.Errorf("expected ACAO=%q, got %q", "https://example.com", acao)
	}

	// Vary header should be set for non-wildcard CORS
	vary := w.Header().Get("Vary")
	if vary != "Origin" {
		t.Errorf("expected Vary=Origin, got %q", vary)
	}
}

func TestCORSBlockedOrigin(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "https://allowed.com")
	defer os.Unsetenv("CORS_ORIGINS")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao != "" {
		t.Errorf("expected empty ACAO for blocked origin, got %q", acao)
	}
}

func TestCORSPreflight(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	})

	wrapped := CORS(handler)
	req := httptest.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for preflight, got %d", w.Code)
	}

	// Body should be empty for preflight
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body for preflight, got %q", w.Body.String())
	}

	// Should have allow-methods
	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}
