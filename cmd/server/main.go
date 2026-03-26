package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/api"
	"github.com/archerny/mini-agent/internal/protocol"
	"github.com/archerny/mini-agent/internal/runtime"
)

const (
	defaultAddr = ":8080"
)

func main() {
	fmt.Println("🚀 mini-agent — AI Agent Runtime")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()

	// Create the runtime engine.
	engine := runtime.NewEngine()

	// Create HTTP router (REST + WebSocket).
	router := api.NewRouter(engine)

	// Determine listen address.
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = defaultAddr
	} else {
		addr = ":" + addr
	}

	// Create HTTP server.
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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

		// Graceful HTTP shutdown.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	// Start the demo scenario in the background.
	go runDemoScenario(ctx, engine)

	// Start HTTP server.
	fmt.Printf("🌐 HTTP server listening on %s\n", addr)
	fmt.Printf("   REST API:  http://localhost%s/api/agents\n", addr)
	fmt.Printf("   WebSocket: ws://localhost%s/ws/events\n", addr)
	fmt.Println()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}

	// Wait for graceful shutdown.
	fmt.Println("🛑 Shutting down agents...")
	engine.Shutdown(ctx)
	fmt.Println("✅ Server stopped.")
}

// ---------------------------------------------------------------------------
// Demo Scenario — runs in the background, spawns agents and generates traffic
// ---------------------------------------------------------------------------

func runDemoScenario(ctx context.Context, engine *runtime.Engine) {
	// Wait a moment for the HTTP server to start.
	time.Sleep(500 * time.Millisecond)

	fmt.Println("📡 Demo: Spawning agents...")

	// Spawn researcher agent.
	agentA, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           "agent-researcher",
		Name:         "researcher",
		Role:         "Research Agent — gathers and analyzes information",
		Capabilities: []string{"web_search", "summarize"},
		Handler:      researcherHandler,
	})
	if err != nil {
		log.Printf("[demo] Failed to spawn researcher: %v", err)
		return
	}

	// Spawn writer agent.
	agentB, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           "agent-writer",
		Name:         "writer",
		Role:         "Writer Agent — generates reports from research data",
		Capabilities: []string{"write", "format"},
		Handler:      writerHandler,
	})
	if err != nil {
		log.Printf("[demo] Failed to spawn writer: %v", err)
		return
	}

	fmt.Println("📡 Demo: Agents spawned. Starting message loop...")

	// Periodically generate messages to keep the demo alive.
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	round := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			round++
			topic := fmt.Sprintf("Research topic #%d: Benefits of multi-agent systems (round %d)", round, round)

			// Writer sends a request to researcher.
			request := protocol.NewRequest(
				agentB.ID(), agentA.ID(),
				protocol.TextPayload(topic),
			)
			if err := engine.MessageBus.Send(request); err != nil {
				log.Printf("[demo] Failed to send request: %v", err)
			}
		}
	}
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
