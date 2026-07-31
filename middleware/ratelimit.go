package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter creates a middleware that limits requests per IP.
// r is the requests per second allowed, b is the burst capacity.
// It correctly reads the real client IP even when behind a reverse proxy (Nginx, Cloudflare, etc.)
// by checking X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
// ctx is used to cancel the background cleanup goroutine when shutting down.
func RateLimiter(ctx context.Context, r rate.Limit, b int) Middleware {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var mu sync.Mutex
	clients := make(map[string]*client)

	// Background cleanup routine to prevent memory leaks.
	// Removes IPs that haven't been seen for 3 minutes.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return // Stop the goroutine when context is cancelled
			case <-ticker.C:
				mu.Lock()
				for ip, c := range clients {
					if time.Since(c.lastSeen) > 3*time.Minute {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			}
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip := realIP(req)

			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(r, b)}
			}
			c := clients[ip]
			c.lastSeen = time.Now()
			mu.Unlock()

			if !c.limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// realIP extracts the real client IP, even behind proxies like Cloudflare or Nginx.
func realIP(r *http.Request) string {
	// 1. Check CF-Connecting-IP (Cloudflare)
	// This is the most secure check if the app is behind Cloudflare.
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.Split(cfIP, ",")[0]
	}

	// 2. Check X-Forwarded-For (Standard reverse proxy)
	// Note: Can be spoofed if not behind a trusted proxy!
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0]) // Return the first (original) client IP
	}

	// 3. Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 4. Fallback to RemoteAddr (TCP connection IP)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
