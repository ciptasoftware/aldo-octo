package middleware

import (
	"bufio"
	"fmt"
	"go-ai/logger"
	"net"
	"net/http"
	"runtime/debug"
)

// statusRecorder wraps http.ResponseWriter to track whether headers have been sent.
// This prevents "http: superfluous response.WriteHeader call" warnings
// when the recover middleware tries to write a 500 after partial response.
type statusRecorder struct {
	http.ResponseWriter
	wroteHeader bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	if !sr.wroteHeader {
		sr.wroteHeader = true
		sr.ResponseWriter.WriteHeader(code)
	}
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	sr.wroteHeader = true // Write implicitly sends headers if not yet sent
	return sr.ResponseWriter.Write(b)
}

// Hijack supports the http.Hijacker interface, required for WebSockets.
func (sr *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := sr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// Flush supports the http.Flusher interface.
func (sr *statusRecorder) Flush() {
	if flusher, ok := sr.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Recover creates a middleware that catches panics, logs them, and returns a 500 Internal Server Error.
// If headers have already been sent before the panic, it logs the error but cannot change the response.
func Recover(log logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sr := &statusRecorder{ResponseWriter: w}
			defer func() {
				if err := recover(); err != nil {
					// http.ErrAbortHandler is used by the standard library to abort a handler.
					// We must NOT recover this, otherwise we break the HTTP server's connection abort semantics.
					if err == http.ErrAbortHandler {
						panic(err)
					}

					// Log the error and the stack trace
					log.Error("Panic recovered", "error", err, "stack", string(debug.Stack()))

					// Only write error response if headers haven't been sent yet
					if !sr.wroteHeader {
						sr.Header().Set("Content-Type", "application/json")
						sr.WriteHeader(http.StatusInternalServerError)
						fmt.Fprint(sr, `{"error":"internal server error"}`)
					} else {
						log.Error("Panic after headers sent — cannot write error response", "error", err)
					}
				}
			}()
			next.ServeHTTP(sr, r)
		})
	}
}
