// Package agent implements the Agent struct and execution engine.
//
// An Agent is a stateful entity (struct + goroutine) that:
//   - Has an identity (ID, name, role, capabilities)
//   - Maintains lifecycle state (spawning → ready → idle ↔ busy → ...)
//   - Processes messages from its inbox channel
//   - Delegates actual message handling to a pluggable Handler function
package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/archerny/mini-agent/internal/protocol"
)

// Handler is the function that an Agent calls to process each message.
// The handler receives the agent (for identity/state) and the incoming message,
// and returns zero or more outbound messages to send.
//
// If the handler panics, the executor will recover and transition to error state.
type Handler func(ctx context.Context, a *Agent, msg *protocol.Message) ([]*protocol.Message, error)

// Agent is a stateful entity in the runtime.
// It is NOT a goroutine — it is a struct that OWNS a goroutine (the executor).
type Agent struct {
	mu sync.RWMutex

	// Identity
	id           string
	name         string
	role         string
	capabilities []string
	accepts      []protocol.MessageType

	// State
	state protocol.AgentState

	// Inbox is the buffered channel where incoming messages are queued.
	// The executor goroutine reads from this channel.
	Inbox chan *protocol.Message

	// Handler is the pluggable message-processing function.
	handler Handler

	// Metadata is an extensible key-value bag.
	metadata map[string]any

	// cancel stops the executor goroutine.
	cancel context.CancelFunc

	// done is closed when the executor goroutine exits.
	done chan struct{}

	// onStateChange is called when the agent's state changes.
	// Set by the AgentManager to emit events.
	onStateChange func(agentID string, prev, next protocol.AgentState, reason string)
}

// Config holds the configuration for creating a new Agent.
type Config struct {
	ID           string
	Name         string
	Role         string
	Capabilities []string
	Accepts      []protocol.MessageType
	Handler      Handler
	Metadata     map[string]any
	InboxSize    int // defaults to 100
}

const defaultInboxSize = 100

// New creates a new Agent in the spawning state.
// The agent is NOT started — call Start() to begin processing.
func New(cfg Config) *Agent {
	inboxSize := cfg.InboxSize
	if inboxSize <= 0 {
		inboxSize = defaultInboxSize
	}
	if cfg.Accepts == nil {
		cfg.Accepts = []protocol.MessageType{
			protocol.TypeMessage,
			protocol.TypeRequest,
			protocol.TypeResponse,
			protocol.TypeBroadcast,
		}
	}
	return &Agent{
		id:           cfg.ID,
		name:         cfg.Name,
		role:         cfg.Role,
		capabilities: cfg.Capabilities,
		accepts:      cfg.Accepts,
		state:        protocol.StateSpawning,
		Inbox:        make(chan *protocol.Message, inboxSize),
		handler:      cfg.Handler,
		metadata:     cfg.Metadata,
		done:         make(chan struct{}),
	}
}

// ---------------------------------------------------------------------------
// Identity accessors (read-only, no lock needed after creation)
// ---------------------------------------------------------------------------

func (a *Agent) ID() string                       { return a.id }
func (a *Agent) Name() string                     { return a.name }
func (a *Agent) Role() string                     { return a.role }
func (a *Agent) Capabilities() []string            { return a.capabilities }
func (a *Agent) Accepts() []protocol.MessageType   { return a.accepts }
func (a *Agent) Metadata() map[string]any          { return a.metadata }

// State returns the current lifecycle state.
func (a *Agent) State() protocol.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Done returns a channel that is closed when the executor goroutine exits.
func (a *Agent) Done() <-chan struct{} {
	return a.done
}

// SetOnStateChange sets the callback for state changes.
func (a *Agent) SetOnStateChange(fn func(agentID string, prev, next protocol.AgentState, reason string)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onStateChange = fn
}

// ---------------------------------------------------------------------------
// State machine
// ---------------------------------------------------------------------------

// transition attempts a state transition. Returns an error if the transition is invalid.
func (a *Agent) transition(to protocol.AgentState, reason string) error {
	a.mu.Lock()
	prev := a.state
	if !protocol.CanTransition(prev, to) {
		a.mu.Unlock()
		return fmt.Errorf("agent %s: invalid transition %s → %s", a.id, prev, to)
	}
	a.state = to
	cb := a.onStateChange
	a.mu.Unlock()

	if cb != nil {
		cb(a.id, prev, to, reason)
	}
	return nil
}

// Card returns the current AgentCard snapshot.
func (a *Agent) Card() protocol.AgentCard {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return protocol.AgentCard{
		ID:           a.id,
		Name:         a.name,
		Role:         a.role,
		Capabilities: a.capabilities,
		Accepts:      a.accepts,
		Status:       a.state,
		Metadata:     a.metadata,
	}
}
