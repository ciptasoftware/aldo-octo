package middleware

import (
	"bufio"
	"go-ai/logger"
	"net"
	"net/http"
	"time"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// AccessLog middleware to record the HTTP status code and response size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.status == 0 {
		rw.status = code
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
		rw.ResponseWriter.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Hijack supports the http.Hijacker interface, required for WebSockets.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// Flush supports the http.Flusher interface.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// AccessLog creates a middleware that logs incoming HTTP requests.
// It logs method, path, remote IP, response status, processing time, and bytes written.
func AccessLog(log logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Default status is 200 OK if WriteHeader wasn't explicitly called
			status := rw.status
			if status == 0 {
				status = http.StatusOK
			}

			// Extract Request ID if available
			reqID := GetRequestID(r.Context())

			log.Info("Request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration", duration.String(),
				"size", rw.size,
				"ip", realIP(r),
				"req_id", reqID,
			)
		})
	}
}
