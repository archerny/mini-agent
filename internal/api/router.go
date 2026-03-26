package api

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/archerny/mini-agent/internal/demo"
	"github.com/archerny/mini-agent/internal/runtime"
)

// NewRouter creates an http.Handler with all API routes registered.
//
// If frontendFS is non-nil, the router also serves the embedded frontend
// as a single-page application (SPA), falling back to index.html for
// non-API, non-WebSocket routes.
//
// Routes:
//
//	GET    /api/agents        — list all agents
//	GET    /api/agents/{id}   — get a single agent
//	POST   /api/agents        — spawn a new agent
//	DELETE /api/agents/{id}   — shutdown an agent
//	GET    /api/messages      — get message history
//	POST   /api/messages      — send a message
//	GET    /api/events        — get event history
//	GET    /api/topology      — get network topology
//	GET    /api/stats         — get global statistics
//	POST   /api/demo/pause    — pause demo scenario
//	POST   /api/demo/resume   — resume demo scenario
//	GET    /api/demo/status   — get demo status
//	WS     /ws/events         — real-time event stream
func NewRouter(engine *runtime.Engine, scenario *demo.ResearchPipeline, rootCtx context.Context, frontendFS fs.FS) http.Handler {
	h := NewHandler(engine, scenario, rootCtx)
	hub := NewHub(engine)

	mux := http.NewServeMux()

	// REST API — read endpoints.
	mux.HandleFunc("GET /api/agents", h.GetAgents)
	mux.HandleFunc("GET /api/agents/{id}", h.GetAgent)
	mux.HandleFunc("GET /api/messages", h.GetMessages)
	mux.HandleFunc("GET /api/events", h.GetEvents)
	mux.HandleFunc("GET /api/topology", h.GetTopology)
	mux.HandleFunc("GET /api/stats", h.GetStats)

	// REST API — write endpoints.
	mux.HandleFunc("POST /api/agents", h.PostAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", h.DeleteAgent)
	mux.HandleFunc("POST /api/messages", h.PostMessage)

	// REST API — demo control.
	mux.HandleFunc("POST /api/demo/pause", h.PostDemoPause)
	mux.HandleFunc("POST /api/demo/resume", h.PostDemoResume)
	mux.HandleFunc("GET /api/demo/status", h.GetDemoStatus)

	// WebSocket endpoint.
	mux.HandleFunc("GET /ws/events", hub.ServeWS)

	// Serve embedded frontend as SPA (if provided).
	if frontendFS != nil {
		fileServer := http.FileServerFS(frontendFS)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Serve static files directly; fall back to index.html for SPA routing.
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			// Check if the file exists.
			if f, err := frontendFS.Open(path); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			// SPA fallback: serve index.html for any unmatched path.
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}

	// Wrap with CORS middleware.
	return corsMiddleware(mux)
}

// corsMiddleware adds CORS headers for development.
// In production, this should be locked down.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
