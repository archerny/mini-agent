// Package api provides the HTTP API layer for the mini-agent runtime.
//
// REST endpoints provide snapshot/history data and control operations.
// WebSocket provides real-time event streaming.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/demo"
	"github.com/archerny/mini-agent/internal/protocol"
	"github.com/archerny/mini-agent/internal/runtime"
)

// Handler holds the REST API handlers.
type Handler struct {
	engine   *runtime.Engine
	scenario *demo.ResearchPipeline
	rootCtx  context.Context
}

// NewHandler creates a new Handler bound to the given runtime engine.
func NewHandler(engine *runtime.Engine, scenario *demo.ResearchPipeline, rootCtx context.Context) *Handler {
	return &Handler{engine: engine, scenario: scenario, rootCtx: rootCtx}
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
// POST /api/messages — send a message between agents
// Body: { "type": "agent.message", "from": "agent-id", "to": "agent-id", "content": "..." }
// ---------------------------------------------------------------------------

type sendMessageRequest struct {
	Type    string `json:"type"`    // message type (agent.message, agent.request, etc.)
	From    string `json:"from"`    // sender agent ID
	To      string `json:"to"`      // recipient agent ID (or "*" for broadcast)
	Content string `json:"content"` // text content
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if req.From == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "from and content are required")
		return
	}

	// Validate sender exists.
	if _, ok := h.engine.AgentManager.GetAgent(req.From); !ok {
		writeError(w, http.StatusNotFound, "sender agent not found: "+req.From)
		return
	}

	// Default type to agent.message.
	msgType := protocol.MessageType(req.Type)
	if msgType == "" {
		msgType = protocol.TypeMessage
	}

	var msg *protocol.Message
	switch msgType {
	case protocol.TypeBroadcast:
		msg = protocol.NewBroadcast(req.From, protocol.TextPayload(req.Content))
	case protocol.TypeRequest:
		if req.To == "" || req.To == "*" {
			writeError(w, http.StatusBadRequest, "request messages require a specific 'to' agent")
			return
		}
		msg = protocol.NewRequest(req.From, req.To, protocol.TextPayload(req.Content))
	case protocol.TypeResponse:
		writeError(w, http.StatusBadRequest, "use agent.message or agent.request; responses are generated by agents")
		return
	default:
		if req.To == "" {
			writeError(w, http.StatusBadRequest, "'to' is required for non-broadcast messages")
			return
		}
		msg = protocol.NewMessage(protocol.TypeMessage, req.From, req.To, protocol.TextPayload(req.Content))
	}

	if err := h.engine.MessageBus.Send(msg); err != nil {
		writeError(w, http.StatusInternalServerError, "send failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": msg.ID, "status": "sent"})
}

// ---------------------------------------------------------------------------
// POST /api/agents — spawn a new agent
// Body: { "id": "my-agent", "name": "My Agent", "role": "helper", "capabilities": ["search"] }
// ---------------------------------------------------------------------------

type spawnAgentRequest struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities"`
}

func (h *Handler) PostAgent(w http.ResponseWriter, r *http.Request) {
	var req spawnAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Generate ID if not provided.
	agentID := req.ID
	if agentID == "" {
		agentID = "agent-" + req.Name + "-" + strconv.FormatInt(time.Now().UnixMilli(), 36)
	}

	// Check for duplicate ID.
	if _, exists := h.engine.AgentManager.GetAgent(agentID); exists {
		writeError(w, http.StatusConflict, "agent already exists: "+agentID)
		return
	}

	// Create a generic echo handler for user-spawned agents.
	handler := func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
		switch msg.Type {
		case protocol.TypeRequest:
			resp := protocol.NewResponse(
				a.ID(), msg.From, msg.ID,
				protocol.TextPayload("["+a.Name()+"] Acknowledged: "+truncateString(msg.Payload.Content, 100)),
			)
			return []*protocol.Message{resp}, nil
		default:
			return nil, nil
		}
	}

	_, err := h.engine.AgentManager.Spawn(h.rootCtx, agent.Config{
		ID:           agentID,
		Name:         req.Name,
		Role:         req.Role,
		Capabilities: req.Capabilities,
		Handler:      handler,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "spawn failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": agentID, "status": "spawned"})
}

// ---------------------------------------------------------------------------
// DELETE /api/agents/{id} — shutdown an agent
// ---------------------------------------------------------------------------

func (h *Handler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing agent id")
		return
	}

	if err := h.engine.AgentManager.ShutdownAgent(id, "user requested shutdown via API"); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "shutdown"})
}

// ---------------------------------------------------------------------------
// POST /api/demo/pause — pause the demo scenario
// ---------------------------------------------------------------------------

func (h *Handler) PostDemoPause(w http.ResponseWriter, r *http.Request) {
	if h.scenario == nil {
		writeError(w, http.StatusNotFound, "no demo scenario running")
		return
	}
	h.scenario.Pause()
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

// ---------------------------------------------------------------------------
// POST /api/demo/resume — resume the demo scenario
// ---------------------------------------------------------------------------

func (h *Handler) PostDemoResume(w http.ResponseWriter, r *http.Request) {
	if h.scenario == nil {
		writeError(w, http.StatusNotFound, "no demo scenario running")
		return
	}
	h.scenario.Resume()
	writeJSON(w, http.StatusOK, map[string]string{"status": "running"})
}

// ---------------------------------------------------------------------------
// GET /api/demo/status — get demo scenario status
// ---------------------------------------------------------------------------

func (h *Handler) GetDemoStatus(w http.ResponseWriter, r *http.Request) {
	if h.scenario == nil {
		writeJSON(w, http.StatusOK, map[string]any{"running": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"running": true,
		"paused":  h.scenario.IsPaused(),
		"name":    h.scenario.Name(),
	})
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

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
