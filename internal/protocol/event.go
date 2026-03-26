package protocol

import (
	"time"

	"github.com/archerny/mini-agent/internal/protocol/uid"
)

// ---------------------------------------------------------------------------
// Event — system-level event (produced by MessageBus / Agent Manager)
// ---------------------------------------------------------------------------

// Event represents a system event in the event stream.
// Events are produced automatically — agents never emit events directly.
type Event struct {
	// ID is a unique identifier (UUID v7, time-ordered).
	ID string `json:"id"`

	// Type is the event type.
	Type EventType `json:"type"`

	// Sequence is a globally incrementing sequence number.
	// Used by the frontend to detect gaps after reconnection.
	Sequence uint64 `json:"sequence"`

	// AgentID is the agent this event relates to (empty for network-level events).
	AgentID string `json:"agent_id,omitempty"`

	// Timestamp is when the event was produced.
	Timestamp time.Time `json:"timestamp"`

	// Data carries event-specific information.
	// The concrete type depends on the event Type.
	Data any `json:"data"`

	// Metadata is an extensible key-value bag for future use.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Event Data types — one per EventType
// ---------------------------------------------------------------------------

// AgentSpawnedData is the data payload for EventAgentSpawned.
type AgentSpawnedData struct {
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities"`
}

// AgentStateChangedData is the data payload for EventAgentStateChanged.
type AgentStateChangedData struct {
	PrevState AgentState `json:"prev_state"`
	NewState  AgentState `json:"new_state"`
	Reason    string     `json:"reason,omitempty"`
}

// AgentMessageSentData is the data payload for EventAgentMessageSent.
// Contains the full Message object (MVP accepts data redundancy).
type AgentMessageSentData struct {
	Message *Message `json:"message"`
}

// AgentMessageReceivedData is the data payload for EventAgentMessageReceived.
type AgentMessageReceivedData struct {
	Message *Message `json:"message"`
}

// AgentErrorData is the data payload for EventAgentError.
type AgentErrorData struct {
	Error       string    `json:"error"`
	Kind        ErrorKind `json:"kind"`
	Recoverable bool      `json:"recoverable"`
	Stack       string    `json:"stack,omitempty"`
}

// AgentShutdownData is the data payload for EventAgentShutdown.
type AgentShutdownData struct {
	Reason   string `json:"reason"`
	ExitCode int    `json:"exit_code"`
}

// TopologyChangedData is the data payload for EventTopologyChanged.
type TopologyChangedData struct {
	ChangeType TopologyChangeType `json:"change_type"`
	Details    map[string]any     `json:"details,omitempty"`
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// newEvent creates a base event. Sequence is set to 0 — the EventStream
// is responsible for assigning the actual sequence number.
func newEvent(eventType EventType, agentID string, data any) *Event {
	return &Event{
		ID:        uid.New(),
		Type:      eventType,
		Sequence:  0, // assigned by EventStream
		AgentID:   agentID,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// NewAgentSpawnedEvent creates an agent.spawned event.
func NewAgentSpawnedEvent(agentID, name, role string, capabilities []string) *Event {
	return newEvent(EventAgentSpawned, agentID, AgentSpawnedData{
		Name:         name,
		Role:         role,
		Capabilities: capabilities,
	})
}

// NewAgentStateChangedEvent creates an agent.state_changed event.
func NewAgentStateChangedEvent(agentID string, prev, next AgentState, reason string) *Event {
	return newEvent(EventAgentStateChanged, agentID, AgentStateChangedData{
		PrevState: prev,
		NewState:  next,
		Reason:    reason,
	})
}

// NewAgentMessageSentEvent creates an agent.message_sent event.
func NewAgentMessageSentEvent(agentID string, msg *Message) *Event {
	return newEvent(EventAgentMessageSent, agentID, AgentMessageSentData{
		Message: msg,
	})
}

// NewAgentMessageReceivedEvent creates an agent.message_received event.
func NewAgentMessageReceivedEvent(agentID string, msg *Message) *Event {
	return newEvent(EventAgentMessageReceived, agentID, AgentMessageReceivedData{
		Message: msg,
	})
}

// NewAgentErrorEvent creates an agent.error event.
func NewAgentErrorEvent(agentID string, err string, kind ErrorKind, recoverable bool) *Event {
	return newEvent(EventAgentError, agentID, AgentErrorData{
		Error:       err,
		Kind:        kind,
		Recoverable: recoverable,
	})
}

// NewAgentShutdownEvent creates an agent.shutdown event.
func NewAgentShutdownEvent(agentID, reason string, exitCode int) *Event {
	return newEvent(EventAgentShutdown, agentID, AgentShutdownData{
		Reason:   reason,
		ExitCode: exitCode,
	})
}

// NewTopologyChangedEvent creates a network.topology_changed event.
func NewTopologyChangedEvent(changeType TopologyChangeType, details map[string]any) *Event {
	return newEvent(EventTopologyChanged, "", TopologyChangedData{
		ChangeType: changeType,
		Details:    details,
	})
}
