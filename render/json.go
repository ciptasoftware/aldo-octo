package render

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// JSON sends a JSON response with the given status code.
// Data is marshaled to bytes first to catch encoding errors before writing.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		slog.Error("JSON marshal error", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
	w.Write([]byte("\n"))
}

// JSONError is a helper for sending standard error responses
func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
