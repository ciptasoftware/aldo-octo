package dev

import (
	"database/sql"
	"go-ai/logger"
	"go-ai/render"
	"go-ai/router"
	"go-ai/websocket"
	"net/http"
	"os"

	ws "golang.org/x/net/websocket"
)

// RegisterRoutes mendaftarkan rute dev, testing, dan WebSocket.
// Rute berbahaya (/api/panic) hanya aktif di mode Development.
func RegisterRoutes(r *router.Router, tmplEngine *render.Engine, log logger.Logger, db *sql.DB) {
	isProd := os.Getenv("APP_ENV") == "production"

	// REST API Route — tersedia di semua mode
	r.Get("/api/hello", func(w http.ResponseWriter, req *http.Request) {
		render.JSON(w, http.StatusOK, map[string]string{
			"message": "Hello from go-ai framework!",
		})
	})

	// POST Route untuk testing JSON body parsing
	r.Post("/api/echo", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]string
		if err := render.BindJSON(req, &body); err != nil {
			render.JSONError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		name := render.BindQuery(req, "name", "Guest")
		body["query_name"] = name

		render.JSON(w, http.StatusOK, body)
	})

	// Panic Route — HANYA aktif di Development Mode
	// Di production, rute ini disabled untuk keamanan server
	if !isProd {
		r.Get("/api/panic", func(w http.ResponseWriter, req *http.Request) {
			panic("Testing panic recovery!")
		})
		log.Info("Dev routes enabled", "routes", "/api/panic")
	}

	// WebSocket Route — tersedia di semua mode
	r.Handle("GET /ws", websocket.Handler(func(conn *ws.Conn) {
		log.Info("Websocket connected")
		var msg string
		for {
			if err := ws.Message.Receive(conn, &msg); err != nil {
				log.Error("Websocket error", "error", err)
				break
			}
			log.Info("Received message", "msg", msg)
			ws.Message.Send(conn, "Echo: "+msg)
		}
	}))
}
