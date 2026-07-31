package websocket

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// HandlerFunc is the function signature for handling websocket connections.
type HandlerFunc func(ws *websocket.Conn)

// Handler wraps a HandlerFunc into an http.Handler.
// For context-aware WebSocket handlers, use the request context
// inside the handler to detect server shutdown.
func Handler(h HandlerFunc) http.Handler {
	return websocket.Handler(h)
}
