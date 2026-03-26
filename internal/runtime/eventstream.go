package runtime

import (
	"sync"
	"sync/atomic"

	"github.com/archerny/mini-agent/internal/protocol"
)

// Subscriber is a function that receives events from the EventStream.
// Subscribers MUST NOT block — if they need to do work, they should
// queue the event internally.
type Subscriber func(event *protocol.Event)

// EventStream is the ordered stream of all system events.
// It assigns globally incrementing sequence numbers and fans out
// events to all subscribers.
//
// Thread-safe: all methods can be called from any goroutine.
type EventStream struct {
	mu          sync.RWMutex
	sequence    atomic.Uint64
	subscribers []Subscriber
	history     []*protocol.Event
	maxHistory  int
}

const defaultMaxHistory = 1000

// NewEventStream creates a new EventStream.
func NewEventStream() *EventStream {
	return &EventStream{
		maxHistory: defaultMaxHistory,
	}
}

// Subscribe adds a subscriber that will receive all future events.
// Returns an unsubscribe function.
func (es *EventStream) Subscribe(sub Subscriber) func() {
	es.mu.Lock()
	idx := len(es.subscribers)
	es.subscribers = append(es.subscribers, sub)
	es.mu.Unlock()

	return func() {
		es.mu.Lock()
		defer es.mu.Unlock()
		// Nil out instead of remove to keep indices stable during iteration.
		if idx < len(es.subscribers) {
			es.subscribers[idx] = nil
		}
	}
}

// Publish assigns a sequence number to the event and fans it out to all subscribers.
func (es *EventStream) Publish(event *protocol.Event) {
	// Assign sequence number (atomic, lock-free).
	seq := es.sequence.Add(1)
	event.Sequence = seq

	// Append to history.
	es.mu.Lock()
	es.history = append(es.history, event)
	if len(es.history) > es.maxHistory {
		// Trim oldest events.
		es.history = es.history[len(es.history)-es.maxHistory:]
	}
	// Snapshot subscribers to avoid holding lock during callbacks.
	subs := make([]Subscriber, len(es.subscribers))
	copy(subs, es.subscribers)
	es.mu.Unlock()

	// Fan-out to all subscribers (synchronous in MVP).
	for _, sub := range subs {
		if sub != nil {
			sub(event)
		}
	}
}

// EventsSince returns all events with sequence > sinceSequence.
// Used by the WebSocket reconnection gap compensation.
func (es *EventStream) EventsSince(sinceSequence uint64, limit int) []*protocol.Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var result []*protocol.Event
	for _, evt := range es.history {
		if evt.Sequence > sinceSequence {
			result = append(result, evt)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

// LastSequence returns the last assigned sequence number.
func (es *EventStream) LastSequence() uint64 {
	return es.sequence.Load()
}

// History returns a copy of the event history (for REST API).
func (es *EventStream) History() []*protocol.Event {
	es.mu.RLock()
	defer es.mu.RUnlock()
	result := make([]*protocol.Event, len(es.history))
	copy(result, es.history)
	return result
}
