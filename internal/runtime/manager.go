package runtime

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/protocol"
)

// AgentManager manages the lifecycle of all agents in the runtime.
// It is the Single Source of lifecycle Events (spawned, state_changed, shutdown).
type AgentManager struct {
	mu          sync.RWMutex
	agents      map[string]*agent.Agent
	eventStream *EventStream
	messageBus  *MessageBus
}

// NewAgentManager creates a new AgentManager.
func NewAgentManager(es *EventStream, mb *MessageBus) *AgentManager {
	return &AgentManager{
		agents:      make(map[string]*agent.Agent),
		eventStream: es,
		messageBus:  mb,
	}
}

// Spawn creates a new agent, registers it, starts its executor, and emits events.
func (am *AgentManager) Spawn(ctx context.Context, cfg agent.Config) (*agent.Agent, error) {
	// Create the agent.
	a := agent.New(cfg)

	// Wire up state change callback → event stream.
	a.SetOnStateChange(func(agentID string, prev, next protocol.AgentState, reason string) {
		am.eventStream.Publish(
			protocol.NewAgentStateChangedEvent(agentID, prev, next, reason),
		)
	})

	// Register with message bus.
	am.messageBus.RegisterAgent(a)

	// Register with manager.
	am.mu.Lock()
	am.agents[a.ID()] = a
	am.mu.Unlock()

	// Start the executor goroutine.
	if err := a.Start(ctx, am.messageBus.SendFunc()); err != nil {
		am.mu.Lock()
		delete(am.agents, a.ID())
		am.mu.Unlock()
		am.messageBus.UnregisterAgent(a.ID())
		return nil, fmt.Errorf("spawn agent %s: %w", cfg.ID, err)
	}

	// Emit spawned event.
	am.eventStream.Publish(
		protocol.NewAgentSpawnedEvent(a.ID(), a.Name(), a.Role(), a.Capabilities()),
	)
	// Emit topology changed event.
	am.eventStream.Publish(
		protocol.NewTopologyChangedEvent(protocol.TopologyAgentJoined, map[string]any{
			"agent_id": a.ID(),
			"name":     a.Name(),
		}),
	)

	log.Printf("[manager] spawned agent %s (%s)", a.Name(), a.ID())
	return a, nil
}

// ShutdownAgent requests a specific agent to shut down.
func (am *AgentManager) ShutdownAgent(id string, reason string) error {
	am.mu.RLock()
	a, ok := am.agents[id]
	am.mu.RUnlock()
	if !ok {
		return fmt.Errorf("agent %q not found", id)
	}

	if err := a.Shutdown(reason); err != nil {
		return err
	}

	// Wait for executor to finish.
	<-a.Done()

	// Unregister.
	am.messageBus.UnregisterAgent(id)
	am.mu.Lock()
	delete(am.agents, id)
	am.mu.Unlock()

	// Emit events.
	am.eventStream.Publish(
		protocol.NewAgentShutdownEvent(id, reason, 0),
	)
	am.eventStream.Publish(
		protocol.NewTopologyChangedEvent(protocol.TopologyAgentLeft, map[string]any{
			"agent_id": id,
		}),
	)

	log.Printf("[manager] shutdown agent %s", id)
	return nil
}

// ShutdownAll shuts down all agents.
func (am *AgentManager) ShutdownAll(reason string) {
	am.mu.RLock()
	ids := make([]string, 0, len(am.agents))
	for id := range am.agents {
		ids = append(ids, id)
	}
	am.mu.RUnlock()

	for _, id := range ids {
		if err := am.ShutdownAgent(id, reason); err != nil {
			log.Printf("[manager] failed to shutdown agent %s: %v", id, err)
		}
	}
}

// GetAgent returns an agent by ID.
func (am *AgentManager) GetAgent(id string) (*agent.Agent, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	a, ok := am.agents[id]
	return a, ok
}

// Agents returns all current agents.
func (am *AgentManager) Agents() []*agent.Agent {
	am.mu.RLock()
	defer am.mu.RUnlock()
	result := make([]*agent.Agent, 0, len(am.agents))
	for _, a := range am.agents {
		result = append(result, a)
	}
	return result
}

// AgentCards returns all current agent cards (for REST API).
func (am *AgentManager) AgentCards() []protocol.AgentCard {
	agents := am.Agents()
	cards := make([]protocol.AgentCard, len(agents))
	for i, a := range agents {
		cards[i] = a.Card()
	}
	return cards
}
