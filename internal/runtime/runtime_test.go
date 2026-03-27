package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/protocol"
)

// ---------------------------------------------------------------------------
// EventStream tests
// ---------------------------------------------------------------------------

func TestEventStream_SequenceIncrement(t *testing.T) {
	es := NewEventStream()

	evt1 := protocol.NewAgentSpawnedEvent("a1", "test", "role", nil)
	evt2 := protocol.NewAgentSpawnedEvent("a2", "test2", "role2", nil)

	es.Publish(evt1)
	es.Publish(evt2)

	if evt1.Sequence != 1 {
		t.Errorf("expected sequence 1, got %d", evt1.Sequence)
	}
	if evt2.Sequence != 2 {
		t.Errorf("expected sequence 2, got %d", evt2.Sequence)
	}
	if es.LastSequence() != 2 {
		t.Errorf("expected last sequence 2, got %d", es.LastSequence())
	}
}

func TestEventStream_Subscribe(t *testing.T) {
	es := NewEventStream()

	var received []*protocol.Event
	var mu sync.Mutex

	es.Subscribe(func(evt *protocol.Event) {
		mu.Lock()
		received = append(received, evt)
		mu.Unlock()
	})

	es.Publish(protocol.NewAgentSpawnedEvent("a1", "test", "role", nil))
	es.Publish(protocol.NewAgentSpawnedEvent("a2", "test2", "role2", nil))

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Errorf("expected 2 events, got %d", len(received))
	}
}

func TestEventStream_Unsubscribe(t *testing.T) {
	es := NewEventStream()
	count := 0
	unsub := es.Subscribe(func(evt *protocol.Event) {
		count++
	})

	es.Publish(protocol.NewAgentSpawnedEvent("a1", "test", "role", nil))
	unsub()
	es.Publish(protocol.NewAgentSpawnedEvent("a2", "test2", "role2", nil))

	if count != 1 {
		t.Errorf("expected 1 event after unsubscribe, got %d", count)
	}
}

func TestEventStream_EventsSince(t *testing.T) {
	es := NewEventStream()

	for range 5 {
		es.Publish(protocol.NewAgentSpawnedEvent("a1", "test", "role", nil))
	}

	events := es.EventsSince(3, 0)
	if len(events) != 2 {
		t.Errorf("expected 2 events since seq 3, got %d", len(events))
	}
	if events[0].Sequence != 4 {
		t.Errorf("expected first event seq 4, got %d", events[0].Sequence)
	}
}

func TestEventStream_EventsSinceWithLimit(t *testing.T) {
	es := NewEventStream()

	for range 10 {
		es.Publish(protocol.NewAgentSpawnedEvent("a1", "test", "role", nil))
	}

	events := es.EventsSince(0, 3)
	if len(events) != 3 {
		t.Errorf("expected 3 events with limit, got %d", len(events))
	}
}

func TestEventStream_History(t *testing.T) {
	es := NewEventStream()

	for range 3 {
		es.Publish(protocol.NewAgentSpawnedEvent("a1", "test", "role", nil))
	}

	history := es.History()
	if len(history) != 3 {
		t.Errorf("expected 3 events in history, got %d", len(history))
	}
}

// ---------------------------------------------------------------------------
// MessageBus tests
// ---------------------------------------------------------------------------

func newTestAgent(id, name string) *agent.Agent {
	return agent.New(agent.Config{
		ID:   id,
		Name: name,
		Role: "test agent",
		Handler: func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
			return nil, nil
		},
	})
}

func TestMessageBus_PointToPoint(t *testing.T) {
	es := NewEventStream()
	mb := NewMessageBus(es)

	a1 := newTestAgent("a1", "alice")
	a2 := newTestAgent("a2", "bob")

	mb.RegisterAgent(a1)
	mb.RegisterAgent(a2)

	// Start agents so they're in idle state.
	ctx := context.Background()
	_ = a1.Start(ctx, mb.SendFunc())
	_ = a2.Start(ctx, mb.SendFunc())

	msg := protocol.NewMessage(protocol.TypeMessage, "a1", "a2", protocol.TextPayload("hello"))
	err := mb.Send(msg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if mb.MessageCount() != 1 {
		t.Errorf("expected message count 1, got %d", mb.MessageCount())
	}
}

func TestMessageBus_TargetNotFound(t *testing.T) {
	es := NewEventStream()
	mb := NewMessageBus(es)

	a1 := newTestAgent("a1", "alice")
	mb.RegisterAgent(a1)
	ctx := context.Background()
	_ = a1.Start(ctx, mb.SendFunc())

	msg := protocol.NewMessage(protocol.TypeMessage, "a1", "nonexistent", protocol.TextPayload("hello"))
	err := mb.Send(msg)
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
}

func TestMessageBus_Broadcast(t *testing.T) {
	es := NewEventStream()
	mb := NewMessageBus(es)

	a1 := newTestAgent("a1", "alice")
	a2 := newTestAgent("a2", "bob")
	a3 := newTestAgent("a3", "charlie")

	mb.RegisterAgent(a1)
	mb.RegisterAgent(a2)
	mb.RegisterAgent(a3)

	ctx := context.Background()
	_ = a1.Start(ctx, mb.SendFunc())
	_ = a2.Start(ctx, mb.SendFunc())
	_ = a3.Start(ctx, mb.SendFunc())

	msg := protocol.NewBroadcast("a1", protocol.TextPayload("hello everyone"))
	err := mb.Send(msg)
	if err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}

	// Give time for messages to be delivered.
	time.Sleep(50 * time.Millisecond)

	// Check edge counts — a1 should have sent to a2 and a3.
	edges := mb.EdgeCounts()
	if edges["a1->a2"] != 1 || edges["a1->a3"] != 1 {
		t.Errorf("expected edges a1->a2=1 a1->a3=1, got %v", edges)
	}
	// a1 should NOT have received its own broadcast.
	if edges["a1->a1"] != 0 {
		t.Errorf("sender should not receive own broadcast")
	}
}

func TestMessageBus_InvalidMessage(t *testing.T) {
	es := NewEventStream()
	mb := NewMessageBus(es)

	// Message with empty From should fail validation.
	msg := &protocol.Message{
		ID:   "test",
		Type: protocol.TypeMessage,
		From: "",
		To:   "a1",
	}
	err := mb.Send(msg)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// ---------------------------------------------------------------------------
// AgentManager tests
// ---------------------------------------------------------------------------

func TestAgentManager_SpawnAndShutdown(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	a, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:   "test-agent",
		Name: "tester",
		Role: "test role",
		Handler: func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if a.State() != protocol.StateIdle {
		t.Errorf("expected idle state, got %s", a.State())
	}

	agents := engine.AgentManager.Agents()
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}

	err = engine.AgentManager.ShutdownAgent("test-agent", "test done")
	if err != nil {
		t.Fatalf("ShutdownAgent failed: %v", err)
	}

	agents = engine.AgentManager.Agents()
	if len(agents) != 0 {
		t.Errorf("expected 0 agents after shutdown, got %d", len(agents))
	}
}

func TestAgentManager_AgentCards(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	_, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           "card-test",
		Name:         "cardy",
		Role:         "card tester",
		Capabilities: []string{"test", "verify"},
		Handler: func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	cards := engine.AgentManager.AgentCards()
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}

	card := cards[0]
	if card.ID != "card-test" {
		t.Errorf("expected ID card-test, got %s", card.ID)
	}
	if card.Name != "cardy" {
		t.Errorf("expected Name cardy, got %s", card.Name)
	}
	if card.Status != protocol.StateIdle {
		t.Errorf("expected status idle, got %s", card.Status)
	}
	if len(card.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(card.Capabilities))
	}

	engine.Shutdown(ctx)
}

// ---------------------------------------------------------------------------
// Engine integration tests
// ---------------------------------------------------------------------------

func TestEngine_TwoAgentExchange(t *testing.T) {
	engine := NewEngine()
	ctx := t.Context()

	// Track events.
	var events []*protocol.Event
	var mu sync.Mutex
	engine.EventStream.Subscribe(func(evt *protocol.Event) {
		mu.Lock()
		events = append(events, evt)
		mu.Unlock()
	})

	// Spawn agent A — echoes back.
	_, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:   "echo-a",
		Name: "echo-a",
		Role: "echo",
		Handler: func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
			if msg.Type == protocol.TypeRequest {
				return []*protocol.Message{
					protocol.NewResponse(a.ID(), msg.From, msg.ID, protocol.TextPayload("echo: "+msg.Payload.Content)),
				}, nil
			}
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Spawn A failed: %v", err)
	}

	// Spawn agent B — no-op handler.
	_, err = engine.AgentManager.Spawn(ctx, agent.Config{
		ID:   "echo-b",
		Name: "echo-b",
		Role: "requester",
		Handler: func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Spawn B failed: %v", err)
	}

	// B sends request to A.
	req := protocol.NewRequest("echo-b", "echo-a", protocol.TextPayload("ping"))
	err = engine.MessageBus.Send(req)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Wait for message flow.
	time.Sleep(200 * time.Millisecond)

	// Verify stats.
	stats := engine.Stats()
	if stats.AgentCount != 2 {
		t.Errorf("expected 2 agents, got %d", stats.AgentCount)
	}
	// At least 2 messages: original request + echo response.
	if stats.MessageCount < 2 {
		t.Errorf("expected at least 2 messages, got %d", stats.MessageCount)
	}

	// Verify topology has edges.
	topo := engine.Topology()
	if len(topo.Nodes) != 2 {
		t.Errorf("expected 2 topology nodes, got %d", len(topo.Nodes))
	}
	if len(topo.Edges) < 1 {
		t.Errorf("expected at least 1 topology edge, got %d", len(topo.Edges))
	}

	// Verify events were generated.
	mu.Lock()
	eventCount := len(events)
	mu.Unlock()
	// Should have: state changes (spawning→ready→idle x2) + spawned x2 + topology x2 +
	// message_sent + message_received + state changes for processing + etc.
	if eventCount < 10 {
		t.Errorf("expected at least 10 events, got %d", eventCount)
	}

	engine.Shutdown(ctx)
}

func TestEngine_Stats(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	stats := engine.Stats()
	if stats.AgentCount != 0 {
		t.Errorf("expected 0 agents, got %d", stats.AgentCount)
	}
	if stats.Uptime <= 0 {
		t.Errorf("expected positive uptime")
	}

	_, _ = engine.AgentManager.Spawn(ctx, agent.Config{
		ID: "s1", Name: "s1", Role: "r",
		Handler: func(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
			return nil, nil
		},
	})

	stats = engine.Stats()
	if stats.AgentCount != 1 {
		t.Errorf("expected 1 agent, got %d", stats.AgentCount)
	}

	engine.Shutdown(ctx)
}
