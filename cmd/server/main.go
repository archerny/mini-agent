package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/protocol"
	"github.com/archerny/mini-agent/internal/runtime"
)

func main() {
	fmt.Println("🚀 mini-agent — AI Agent Runtime")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("M1: Minimal Runtime — 2 Agents exchanging messages")
	fmt.Println()

	// Create the runtime engine.
	engine := runtime.NewEngine()

	// Subscribe to the event stream — print all events.
	engine.EventStream.Subscribe(func(evt *protocol.Event) {
		data, _ := json.Marshal(evt)
		fmt.Printf("  [event #%d] %s\n", evt.Sequence, string(data))
	})

	// Create a root context with signal handling.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT / SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n🛑 Received shutdown signal...")
		cancel()
	}()

	// --- Demo: Spawn 2 agents that exchange messages ---

	fmt.Println("📡 Spawning agents...")
	fmt.Println()

	// Agent A: "researcher" — receives requests, replies with a research result.
	agentA, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           "agent-researcher",
		Name:         "researcher",
		Role:         "Research Agent — gathers and analyzes information",
		Capabilities: []string{"web_search", "summarize"},
		Handler:      researcherHandler,
	})
	if err != nil {
		log.Fatalf("Failed to spawn researcher: %v", err)
	}

	// Agent B: "writer" — sends a request to researcher, receives response, writes output.
	agentB, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           "agent-writer",
		Name:         "writer",
		Role:         "Writer Agent — generates reports from research data",
		Capabilities: []string{"write", "format"},
		Handler:      writerHandler,
	})
	if err != nil {
		log.Fatalf("Failed to spawn writer: %v", err)
	}

	fmt.Println()
	fmt.Println("💬 Starting message exchange...")
	fmt.Println("───────────────────────────────────────────")
	fmt.Println()

	// Writer sends a request to researcher.
	request := protocol.NewRequest(
		agentB.ID(), agentA.ID(),
		protocol.TextPayload("Please research the topic: 'Benefits of multi-agent systems'"),
	)
	if err := engine.MessageBus.Send(request); err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Wait a bit for messages to flow.
	time.Sleep(500 * time.Millisecond)

	fmt.Println()
	fmt.Println("───────────────────────────────────────────")
	fmt.Println("📊 Runtime Statistics:")
	stats := engine.Stats()
	statsJSON, _ := json.MarshalIndent(stats, "  ", "  ")
	fmt.Printf("  %s\n", string(statsJSON))

	fmt.Println()
	fmt.Println("🌐 Network Topology:")
	topo := engine.Topology()
	topoJSON, _ := json.MarshalIndent(topo, "  ", "  ")
	fmt.Printf("  %s\n", string(topoJSON))

	fmt.Println()
	fmt.Println("🛑 Shutting down...")
	engine.Shutdown(ctx)

	fmt.Println()
	fmt.Println("✅ M1 demo complete!")

	_ = agentA
	_ = agentB
}

// ---------------------------------------------------------------------------
// Demo Handlers
// ---------------------------------------------------------------------------

// researcherHandler processes incoming messages and replies with research results.
func researcherHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s: %s", a.Name(), msg.Type, msg.From, msg.Payload.Content)

	switch msg.Type {
	case protocol.TypeRequest:
		// Simulate research work.
		time.Sleep(50 * time.Millisecond)

		response := protocol.NewResponse(
			a.ID(), msg.From, msg.ID,
			protocol.TextPayload("Research complete: Multi-agent systems enable parallel task execution, "+
				"dynamic role assignment, and emergent collaborative behavior. "+
				"Key benefits include scalability, fault tolerance, and specialization."),
		)
		log.Printf("[%s] sending response to %s", a.Name(), msg.From)
		return []*protocol.Message{response}, nil

	case protocol.TypeMessage:
		log.Printf("[%s] noted message: %s", a.Name(), msg.Payload.Content)
		return nil, nil

	default:
		return nil, nil
	}
}

// writerHandler processes incoming messages. When it receives a research response,
// it sends a follow-up message.
func writerHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s: %s", a.Name(), msg.Type, msg.From, msg.Payload.Content)

	switch msg.Type {
	case protocol.TypeResponse:
		// Got research results, send a thank-you message.
		time.Sleep(30 * time.Millisecond)

		thankYou := protocol.NewMessage(
			protocol.TypeMessage,
			a.ID(), msg.From,
			protocol.TextPayload("Thanks! I've incorporated the research into the report. Great collaboration!"),
		)
		log.Printf("[%s] sending thank-you to %s", a.Name(), msg.From)
		return []*protocol.Message{thankYou}, nil

	default:
		return nil, nil
	}
}
