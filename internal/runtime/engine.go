package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/archerny/mini-agent/internal/protocol"
)

// Engine is the top-level runtime that wires together all components:
// EventStream, MessageBus, and AgentManager.
type Engine struct {
	EventStream  *EventStream
	MessageBus   *MessageBus
	AgentManager *AgentManager
	startTime    time.Time
}

// NewEngine creates and wires a new runtime engine.
func NewEngine() *Engine {
	es := NewEventStream()
	mb := NewMessageBus(es)
	am := NewAgentManager(es, mb)

	return &Engine{
		EventStream:  es,
		MessageBus:   mb,
		AgentManager: am,
		startTime:    time.Now(),
	}
}

// Shutdown gracefully shuts down all agents.
func (e *Engine) Shutdown(ctx context.Context) {
	e.AgentManager.ShutdownAll("runtime shutdown")
}

// Stats returns the current global statistics.
func (e *Engine) Stats() protocol.Stats {
	agents := e.AgentManager.Agents()
	active := 0
	errorCount := 0
	for _, a := range agents {
		s := a.State()
		if s == protocol.StateBusy {
			active++
		}
		if s == protocol.StateError {
			errorCount++
		}
	}
	return protocol.Stats{
		AgentCount:   len(agents),
		MessageCount: int(e.MessageBus.MessageCount()),
		ActiveAgents: active,
		ErrorCount:   errorCount,
		Uptime:       time.Since(e.startTime).Seconds(),
	}
}

// Topology returns the current network topology snapshot.
func (e *Engine) Topology() protocol.Topology {
	cards := e.AgentManager.AgentCards()
	edgeCounts := e.MessageBus.EdgeCounts()

	var edges []protocol.TopologyEdge
	for key, count := range edgeCounts {
		parts := strings.SplitN(key, "->", 2)
		if len(parts) == 2 {
			edges = append(edges, protocol.TopologyEdge{
				From:         parts[0],
				To:           parts[1],
				MessageCount: count,
			})
		}
	}

	return protocol.Topology{
		Nodes: cards,
		Edges: edges,
	}
}
