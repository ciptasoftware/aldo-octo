package render

import (
	"encoding/json"
	"errors"
	"net/http"
)

// MaxBodySize is the maximum allowed request body size (1 MB).
// This prevents memory exhaustion from large payloads.
const MaxBodySize = 1 << 20 // 1 MB

// ErrBodyTooLarge is returned when the request body exceeds MaxBodySize.
var ErrBodyTooLarge = errors.New("request body too large")

// BindJSON parses the request JSON body into the provided struct.
// It limits the body size to MaxBodySize to prevent memory exhaustion attacks.
// Does NOT require a ResponseWriter — safe to call from any handler.
func BindJSON(r *http.Request, obj any) error {
	defer r.Body.Close()

	// MaxBytesReader limits the request body size and terminates the connection early
	// if the limit is exceeded, preventing resource exhaustion.
	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(obj); err != nil {
		// Detect if the error is due to hitting the MaxBytesReader limit
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return ErrBodyTooLarge
		}
		return err
	}
	return nil
}

// BindQuery retrieves a query parameter, returning defaultVal if empty.
func BindQuery(r *http.Request, key, defaultVal string) string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}
