// Package protocol defines the Agent Communication Protocol types.
//
// The protocol is the "constitution" of the mini-agent system.
// All components (Runtime, API, Frontend) are built around these types.
package protocol

import (
	"slices"
	"time"
)

// ---------------------------------------------------------------------------
// Agent States
// ---------------------------------------------------------------------------

// AgentState represents the lifecycle state of an Agent.
type AgentState string

const (
	StateSpawning  AgentState = "spawning"
	StateReady     AgentState = "ready"
	StateIdle      AgentState = "idle"
	StateBusy      AgentState = "busy"
	StateCompleted AgentState = "completed"
	StateError     AgentState = "error"
	StateShutdown  AgentState = "shutdown"
)

// ValidTransitions defines the complete state transition table.
// Key = current state, Value = set of valid next states.
var ValidTransitions = map[AgentState][]AgentState{
	StateSpawning:  {StateReady, StateError},
	StateReady:     {StateIdle},
	StateIdle:      {StateBusy, StateError, StateShutdown},
	StateBusy:      {StateCompleted, StateError},
	StateCompleted: {StateIdle},
	StateError:     {StateIdle, StateShutdown},
	// StateShutdown is terminal — no transitions out.
}

// CanTransition checks whether transitioning from `from` to `to` is valid.
func CanTransition(from, to AgentState) bool {
	targets, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	return slices.Contains(targets, to)
}

// IsTerminal returns true if the state is a terminal state (no further transitions).
func (s AgentState) IsTerminal() bool {
	return s == StateShutdown
}

// IsTransient returns true if the state is transient (auto-transitions to next state).
// ready → idle, completed → idle.
func (s AgentState) IsTransient() bool {
	return s == StateReady || s == StateCompleted
}

// ---------------------------------------------------------------------------
// Message Types
// ---------------------------------------------------------------------------

// MessageType represents the type of an inter-agent message.
type MessageType string

const (
	TypeMessage   MessageType = "agent.message"
	TypeRequest   MessageType = "agent.request"
	TypeResponse  MessageType = "agent.response"
	TypeBroadcast MessageType = "agent.broadcast"
)

// IsBroadcast returns true if the message is a broadcast.
func (t MessageType) IsBroadcast() bool {
	return t == TypeBroadcast
}

// ---------------------------------------------------------------------------
// Content Types
// ---------------------------------------------------------------------------

// ContentType represents the payload content type.
type ContentType string

const (
	ContentText       ContentType = "text"
	ContentJSON       ContentType = "json"
	ContentToolCall   ContentType = "tool_call"
	ContentToolResult ContentType = "tool_result"
)

// ---------------------------------------------------------------------------
// Event Types
// ---------------------------------------------------------------------------

// EventType represents the type of a system event.
type EventType string

const (
	EventAgentSpawned         EventType = "agent.spawned"
	EventAgentStateChanged    EventType = "agent.state_changed"
	EventAgentMessageSent     EventType = "agent.message_sent"
	EventAgentMessageReceived EventType = "agent.message_received"
	EventAgentError           EventType = "agent.error"
	EventAgentShutdown        EventType = "agent.shutdown"
	EventTopologyChanged      EventType = "network.topology_changed"
)

// ---------------------------------------------------------------------------
// Error Types (for agent.error events)
// ---------------------------------------------------------------------------

// ErrorKind classifies the kind of agent error.
type ErrorKind string

const (
	ErrorUndeliverable  ErrorKind = "undeliverable"
	ErrorBackpressure   ErrorKind = "backpressure"
	ErrorInvalidMessage ErrorKind = "invalid_message"
	ErrorPayloadTooLarge ErrorKind = "payload_too_large"
	ErrorPanic          ErrorKind = "panic"
	ErrorTimeout        ErrorKind = "timeout"
	ErrorInternal       ErrorKind = "internal"
)

// ---------------------------------------------------------------------------
// Topology Change Types
// ---------------------------------------------------------------------------

// TopologyChangeType represents the kind of topology change.
type TopologyChangeType string

const (
	TopologyAgentJoined  TopologyChangeType = "agent_joined"
	TopologyAgentLeft    TopologyChangeType = "agent_left"
	TopologyLinkCreated  TopologyChangeType = "link_created"
	TopologyLinkRemoved  TopologyChangeType = "link_removed"
)

// ---------------------------------------------------------------------------
// Shared Helpers
// ---------------------------------------------------------------------------

// BroadcastTarget is the sentinel value for broadcast messages.
const BroadcastTarget = "*"

// MaxPayloadSize is the maximum allowed message payload size in bytes (1MB).
const MaxPayloadSize = 1 << 20

// BroadcastRateLimit is the max broadcasts per agent per second.
const BroadcastRateLimit = 10

// Timestamp is an alias for time.Time with JSON serialization in RFC3339Nano.
type Timestamp = time.Time
