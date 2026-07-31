package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// reqIDKey is an unexported type for context keys defined in this package.
type reqIDKey struct{}

// RequestID generates a unique ID for each request if not provided by the client,
// injects it into the request context, and adds it to the response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateID()
		}

		ctx := context.WithValue(r.Context(), reqIDKey{}, reqID)
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID returns the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(reqIDKey{}).(string); ok {
		return reqID
	}
	return ""
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
