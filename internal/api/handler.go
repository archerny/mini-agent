// Package api provides the HTTP API layer for the mini-agent runtime.
//
// REST endpoints provide snapshot/history data.
// WebSocket provides real-time event streaming.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/archerny/mini-agent/internal/runtime"
)

// Handler holds the REST API handlers.
type Handler struct {
	engine *runtime.Engine
}

// NewHandler creates a new Handler bound to the given runtime engine.
func NewHandler(engine *runtime.Engine) *Handler {
	return &Handler{engine: engine}
}

// ---------------------------------------------------------------------------
// GET /api/agents — list all agents
// ---------------------------------------------------------------------------

func (h *Handler) GetAgents(w http.ResponseWriter, r *http.Request) {
	cards := h.engine.AgentManager.AgentCards()
	writeJSON(w, http.StatusOK, cards)
}

// ---------------------------------------------------------------------------
// GET /api/agents/:id — get a single agent
// ---------------------------------------------------------------------------

func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	// Extract agent ID from URL path: /api/agents/{id}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing agent id")
		return
	}

	agent, ok := h.engine.AgentManager.GetAgent(id)
	if !ok {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}

	writeJSON(w, http.StatusOK, agent.Card())
}

// ---------------------------------------------------------------------------
// GET /api/messages — get message history
// Query params: ?limit=N (default 100)
// ---------------------------------------------------------------------------

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)

	// Collect messages from event stream (message_sent events contain the message).
	events := h.engine.EventStream.History()
	var messages []json.RawMessage

	for i := len(events) - 1; i >= 0 && len(messages) < limit; i-- {
		evt := events[i]
		if evt.Type == "agent.message_sent" {
			data, err := json.Marshal(evt.Data)
			if err != nil {
				continue
			}
			// Extract the message from the event data.
			var wrapper struct {
				Message json.RawMessage `json:"message"`
			}
			if err := json.Unmarshal(data, &wrapper); err != nil {
				continue
			}
			messages = append(messages, wrapper.Message)
		}
	}

	// Reverse to chronological order.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	writeJSON(w, http.StatusOK, messages)
}

// ---------------------------------------------------------------------------
// GET /api/events — get event history
// Query params: ?since_sequence=N&limit=N
// ---------------------------------------------------------------------------

func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	sinceSeq := queryUint64(r, "since_sequence", 0)
	limit := queryInt(r, "limit", 100)

	events := h.engine.EventStream.EventsSince(sinceSeq, limit)
	writeJSON(w, http.StatusOK, events)
}

// ---------------------------------------------------------------------------
// GET /api/topology — get current network topology
// ---------------------------------------------------------------------------

func (h *Handler) GetTopology(w http.ResponseWriter, r *http.Request) {
	topo := h.engine.Topology()
	writeJSON(w, http.StatusOK, topo)
}

// ---------------------------------------------------------------------------
// GET /api/stats — get global statistics
// ---------------------------------------------------------------------------

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.engine.Stats()
	writeJSON(w, http.StatusOK, stats)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func queryUint64(r *http.Request, key string, defaultVal uint64) uint64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return defaultVal
	}
	return n
}
