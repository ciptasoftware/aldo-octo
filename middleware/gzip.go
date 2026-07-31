package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipPool minimizes memory allocations for gzip writers
var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

// gzipResponseWriter wraps http.ResponseWriter to compress the output.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Flush supports the http.Flusher interface for streaming responses.
func (w *gzipResponseWriter) Flush() {
	// Flush the gzip writer
	if flusher, ok := w.Writer.(*gzip.Writer); ok {
		flusher.Flush()
	}
	// Flush the underlying response writer
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Gzip creates a middleware that compresses HTTP responses using gzip
// if the client supports it (Accept-Encoding: gzip).
func Gzip() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client supports gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Bypass compression for WebSockets (requires Hijack which gzipResponseWriter doesn't implement)
			if strings.Contains(strings.ToLower(r.Header.Get("Upgrade")), "websocket") {
				next.ServeHTTP(w, r)
				return
			}

			// Bypass compression for Server-Sent Events (requires immediate unbuffered streaming)
			if strings.Contains(strings.ToLower(r.Header.Get("Accept")), "text/event-stream") {
				next.ServeHTTP(w, r)
				return
			}

			// Some content types shouldn't be compressed or are already compressed.
			// The caller handler should set Content-Type, but gzip writer
			// will intercept Write() anyway. We just set the header.
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")

			// Get a writer from the pool
			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				gz.Close()
				gzipPool.Put(gz)
			}()

			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
			}

			next.ServeHTTP(gzw, r)
		})
	}
}
