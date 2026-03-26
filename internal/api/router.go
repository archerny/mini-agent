package api

import (
	"net/http"

	"github.com/archerny/mini-agent/internal/runtime"
)

// NewRouter creates an http.Handler with all API routes registered.
//
// Routes:
//
//	GET  /api/agents        — list all agents
//	GET  /api/agents/{id}   — get a single agent
//	GET  /api/messages      — get message history
//	GET  /api/events        — get event history
//	GET  /api/topology      — get network topology
//	GET  /api/stats         — get global statistics
//	WS   /ws/events         — real-time event stream
func NewRouter(engine *runtime.Engine) http.Handler {
	h := NewHandler(engine)
	hub := NewHub(engine)

	mux := http.NewServeMux()

	// REST API endpoints.
	mux.HandleFunc("GET /api/agents", h.GetAgents)
	mux.HandleFunc("GET /api/agents/{id}", h.GetAgent)
	mux.HandleFunc("GET /api/messages", h.GetMessages)
	mux.HandleFunc("GET /api/events", h.GetEvents)
	mux.HandleFunc("GET /api/topology", h.GetTopology)
	mux.HandleFunc("GET /api/stats", h.GetStats)

	// WebSocket endpoint.
	mux.HandleFunc("GET /ws/events", hub.ServeWS)

	// Wrap with CORS middleware.
	return corsMiddleware(mux)
}

// corsMiddleware adds CORS headers for development.
// In production, this should be locked down.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
