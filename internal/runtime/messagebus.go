package runtime

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/protocol"
)

// MessageBus routes messages between agents and is the Single Source of
// message-related Events (message_sent, message_received, error).
//
// It does NOT manage agent lifecycle — that's the AgentManager's job.
// The MessageBus only needs a way to look up agents and their inboxes.
type MessageBus struct {
	mu           sync.RWMutex
	agents       map[string]*agent.Agent
	eventStream  *EventStream
	messageCount atomic.Int64

	// Rate limiting: track broadcast counts per agent per second.
	broadcastMu    sync.Mutex
	broadcastCount map[string]*rateBucket

	// Topology tracking: message counts between agent pairs.
	topologyMu sync.Mutex
	edgeCounts map[string]int // key: "from->to"
}

type rateBucket struct {
	count  int
	window time.Time
}

// NewMessageBus creates a new MessageBus.
func NewMessageBus(es *EventStream) *MessageBus {
	return &MessageBus{
		agents:         make(map[string]*agent.Agent),
		eventStream:    es,
		broadcastCount: make(map[string]*rateBucket),
		edgeCounts:     make(map[string]int),
	}
}

// RegisterAgent makes an agent available for message routing.
func (mb *MessageBus) RegisterAgent(a *agent.Agent) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.agents[a.ID()] = a
}

// UnregisterAgent removes an agent from message routing.
func (mb *MessageBus) UnregisterAgent(id string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	delete(mb.agents, id)
}

// Send routes a message to the target agent(s).
// This is the ONLY entry point for all inter-agent messaging.
//
// It validates the message, routes it, and emits events.
func (mb *MessageBus) Send(msg *protocol.Message) error {
	// Validate message.
	if err := msg.Validate(); err != nil {
		mb.emitError(msg.From, err.Error(), protocol.ErrorInvalidMessage)
		return err
	}

	// Emit message_sent event.
	mb.eventStream.Publish(protocol.NewAgentMessageSentEvent(msg.From, msg))
	mb.messageCount.Add(1)

	// Route based on message type.
	if msg.Type.IsBroadcast() {
		return mb.broadcastMessage(msg)
	}
	return mb.routePointToPoint(msg)
}

// routePointToPoint delivers a message to a single target agent.
func (mb *MessageBus) routePointToPoint(msg *protocol.Message) error {
	mb.mu.RLock()
	target, ok := mb.agents[msg.To]
	mb.mu.RUnlock()

	if !ok {
		errMsg := fmt.Sprintf("target agent %q not found", msg.To)
		mb.emitError(msg.From, errMsg, protocol.ErrorUndeliverable)
		return errors.New(errMsg)
	}

	// Check if target is shutdown.
	if target.State() == protocol.StateShutdown {
		errMsg := fmt.Sprintf("target agent %q is shutdown", msg.To)
		mb.emitError(msg.From, errMsg, protocol.ErrorUndeliverable)
		return errors.New(errMsg)
	}

	// Try to deliver to inbox (non-blocking).
	select {
	case target.Inbox <- msg:
		mb.eventStream.Publish(protocol.NewAgentMessageReceivedEvent(msg.To, msg))
		mb.trackEdge(msg.From, msg.To)
		return nil
	default:
		errMsg := fmt.Sprintf("agent %q inbox full (backpressure)", msg.To)
		mb.emitError(msg.From, errMsg, protocol.ErrorBackpressure)
		return errors.New(errMsg)
	}
}

// broadcastMessage fans out a message to all agents except the sender.
func (mb *MessageBus) broadcastMessage(msg *protocol.Message) error {
	// Rate limit check.
	if !mb.checkBroadcastRate(msg.From) {
		errMsg := fmt.Sprintf("agent %q broadcast rate limit exceeded", msg.From)
		mb.emitError(msg.From, errMsg, protocol.ErrorBackpressure)
		return errors.New(errMsg)
	}

	mb.mu.RLock()
	var targets []*agent.Agent
	for id, a := range mb.agents {
		if id != msg.From && a.State() != protocol.StateShutdown {
			targets = append(targets, a)
		}
	}
	mb.mu.RUnlock()

	for _, target := range targets {
		select {
		case target.Inbox <- msg:
			mb.eventStream.Publish(protocol.NewAgentMessageReceivedEvent(target.ID(), msg))
			mb.trackEdge(msg.From, target.ID())
		default:
			log.Printf("[messagebus] broadcast: agent %q inbox full, skipping", target.ID())
		}
	}
	return nil
}

// checkBroadcastRate enforces the per-agent broadcast rate limit.
func (mb *MessageBus) checkBroadcastRate(agentID string) bool {
	mb.broadcastMu.Lock()
	defer mb.broadcastMu.Unlock()

	now := time.Now()
	bucket, ok := mb.broadcastCount[agentID]
	if !ok || now.Sub(bucket.window) >= time.Second {
		mb.broadcastCount[agentID] = &rateBucket{count: 1, window: now}
		return true
	}
	bucket.count++
	return bucket.count <= protocol.BroadcastRateLimit
}

// trackEdge increments the message count for an agent pair (for topology).
func (mb *MessageBus) trackEdge(from, to string) {
	mb.topologyMu.Lock()
	defer mb.topologyMu.Unlock()
	key := from + "->" + to
	mb.edgeCounts[key]++
}

// emitError publishes an agent error event.
func (mb *MessageBus) emitError(agentID, errMsg string, kind protocol.ErrorKind) {
	mb.eventStream.Publish(protocol.NewAgentErrorEvent(agentID, errMsg, kind, true))
}

// MessageCount returns the total number of messages sent.
func (mb *MessageBus) MessageCount() int64 {
	return mb.messageCount.Load()
}

// EdgeCounts returns a snapshot of the topology edge counts.
func (mb *MessageBus) EdgeCounts() map[string]int {
	mb.topologyMu.Lock()
	defer mb.topologyMu.Unlock()
	result := make(map[string]int, len(mb.edgeCounts))
	maps.Copy(result, mb.edgeCounts)
	return result
}

// SendFunc returns a send function bound to this MessageBus.
// Used to inject into Agent executors.
func (mb *MessageBus) SendFunc() agent.SendFunc {
	return func(msg *protocol.Message) error {
		return mb.Send(msg)
	}
}
