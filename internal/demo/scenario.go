// Package demo provides demonstration scenarios for the mini-agent runtime.
//
// Each scenario spawns a set of agents and generates inter-agent traffic to
// showcase the real-time visualization dashboard.
package demo

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/archerny/mini-agent/internal/agent"
	"github.com/archerny/mini-agent/internal/protocol"
	"github.com/archerny/mini-agent/internal/runtime"
)

// ---------------------------------------------------------------------------
// Scenario — pluggable demo scenario interface
// ---------------------------------------------------------------------------

// Scenario defines a demo scenario that can be run against the runtime engine.
type Scenario interface {
	// Name returns the scenario display name.
	Name() string
	// Run starts the scenario. It should block until ctx is cancelled.
	Run(ctx context.Context, engine *runtime.Engine) error
}

// ---------------------------------------------------------------------------
// ResearchPipeline — multi-agent collaboration demo
//
// 5 agents collaborate on a research report:
//
//   coordinator ──request──► researcher ──response──► coordinator
//   coordinator ──request──► writer     ──response──► coordinator
//   coordinator ──request──► reviewer   ──response──► coordinator
//   coordinator ──request──► publisher  ──response──► coordinator
//   publisher   ──broadcast──► * (completion notification)
//
// The coordinator drives the pipeline in phases:
//   Phase 1: Research   — coordinator asks researcher for data
//   Phase 2: Writing    — coordinator sends research to writer
//   Phase 3: Review     — coordinator sends draft to reviewer
//   Phase 4: Revision   — if reviewer rejects, back to writer (feedback loop)
//   Phase 5: Publishing — coordinator sends final draft to publisher
//   Phase 6: Broadcast  — publisher announces completion to all
// ---------------------------------------------------------------------------

// ResearchPipeline implements a multi-agent research report workflow.
type ResearchPipeline struct{}

// Name returns the scenario name.
func (rp *ResearchPipeline) Name() string {
	return "Research Pipeline — Multi-Agent Collaboration"
}

// Agent IDs (stable, for topology visualization).
const (
	idCoordinator = "agent-coordinator"
	idResearcher  = "agent-researcher"
	idWriter      = "agent-writer"
	idReviewer    = "agent-reviewer"
	idPublisher   = "agent-publisher"
)

// Research topics for variety.
var topics = []string{
	"The Future of Multi-Agent Systems in Enterprise Software",
	"Emergent Behavior in Decentralized Agent Networks",
	"Fault Tolerance Patterns for Autonomous Agent Swarms",
	"Real-Time Visualization of Agent Communication Topologies",
	"Scalability Challenges in Agent-to-Agent Messaging",
	"Self-Organizing Agent Hierarchies for Complex Task Decomposition",
	"Event-Driven Architectures for Agent Runtime Systems",
	"Observability and Debugging in Multi-Agent Environments",
}

// Run starts the research pipeline scenario.
func (rp *ResearchPipeline) Run(ctx context.Context, engine *runtime.Engine) error {
	// Wait for HTTP server to start.
	time.Sleep(500 * time.Millisecond)

	fmt.Println("📡 Demo: Research Pipeline — spawning 5 agents...")

	// -----------------------------------------------------------------------
	// Spawn agents
	// -----------------------------------------------------------------------

	agents, err := spawnAgents(ctx, engine)
	if err != nil {
		return fmt.Errorf("spawn agents: %w", err)
	}

	fmt.Println("📡 Demo: All agents online. Starting research pipeline...")
	fmt.Println()

	// -----------------------------------------------------------------------
	// Main loop — run a new research round every 5 seconds
	// -----------------------------------------------------------------------

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	round := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			round++
			topic := topics[round%len(topics)]
			go runResearchRound(ctx, engine, agents, round, topic)
		}
	}
}

// agentSet holds references to all demo agents.
type agentSet struct {
	coordinator *agent.Agent
	researcher  *agent.Agent
	writer      *agent.Agent
	reviewer    *agent.Agent
	publisher   *agent.Agent
}

func spawnAgents(ctx context.Context, engine *runtime.Engine) (*agentSet, error) {
	coordinator, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           idCoordinator,
		Name:         "coordinator",
		Role:         "Project Coordinator — orchestrates the research pipeline",
		Capabilities: []string{"orchestrate", "delegate", "aggregate"},
		Handler:      coordinatorHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn coordinator: %w", err)
	}

	researcher, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           idResearcher,
		Name:         "researcher",
		Role:         "Research Agent — gathers and analyzes information",
		Capabilities: []string{"web_search", "analyze", "summarize"},
		Handler:      researcherHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn researcher: %w", err)
	}

	writer, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           idWriter,
		Name:         "writer",
		Role:         "Writer Agent — drafts reports from research data",
		Capabilities: []string{"write", "format", "revise"},
		Handler:      writerHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn writer: %w", err)
	}

	reviewer, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           idReviewer,
		Name:         "reviewer",
		Role:         "Reviewer Agent — quality assurance and feedback",
		Capabilities: []string{"review", "critique", "approve"},
		Handler:      reviewerHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn reviewer: %w", err)
	}

	publisher, err := engine.AgentManager.Spawn(ctx, agent.Config{
		ID:           idPublisher,
		Name:         "publisher",
		Role:         "Publisher Agent — final formatting and distribution",
		Capabilities: []string{"publish", "distribute", "archive"},
		Handler:      publisherHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn publisher: %w", err)
	}

	return &agentSet{
		coordinator: coordinator,
		researcher:  researcher,
		writer:      writer,
		reviewer:    reviewer,
		publisher:   publisher,
	}, nil
}

// ---------------------------------------------------------------------------
// Research round — one complete pipeline execution
// ---------------------------------------------------------------------------

func runResearchRound(ctx context.Context, engine *runtime.Engine, agents *agentSet, round int, topic string) {
	log.Printf("[demo] ═══ Round %d: %s ═══", round, topic)

	// Phase 1: Coordinator → Researcher (request research)
	sendWithDelay(ctx, engine, protocol.NewRequest(
		idCoordinator, idResearcher,
		protocol.TextPayload(fmt.Sprintf("[Round %d] Please research: %s", round, topic)),
	), randomDelay(200, 500))

	// Phase 2: Coordinator → Writer (after small delay, simulating receiving research)
	sendWithDelay(ctx, engine, protocol.NewRequest(
		idCoordinator, idWriter,
		protocol.TextPayload(fmt.Sprintf("[Round %d] Please draft a report section on: %s", round, topic)),
	), randomDelay(800, 1500))

	// Phase 3: Coordinator → Reviewer (send draft for review)
	sendWithDelay(ctx, engine, protocol.NewRequest(
		idCoordinator, idReviewer,
		protocol.TextPayload(fmt.Sprintf("[Round %d] Please review the draft report on: %s", round, topic)),
	), randomDelay(1500, 2500))

	// Phase 4: Reviewer feedback loop — reviewer sometimes sends feedback to writer
	if round%3 == 0 {
		// Every 3rd round: reviewer requests revision from writer
		sendWithDelay(ctx, engine, protocol.NewRequest(
			idReviewer, idWriter,
			protocol.TextPayload(fmt.Sprintf("[Round %d] Revision needed: please strengthen the conclusion on: %s", round, topic)),
		), randomDelay(2000, 3000))
	}

	// Phase 5: Coordinator → Publisher
	sendWithDelay(ctx, engine, protocol.NewRequest(
		idCoordinator, idPublisher,
		protocol.TextPayload(fmt.Sprintf("[Round %d] Approved for publishing: %s", round, topic)),
	), randomDelay(3000, 4000))

	// Phase 6: Cross-agent communication — researcher shares findings with writer
	if round%2 == 0 {
		sendWithDelay(ctx, engine, protocol.NewMessage(
			protocol.TypeMessage,
			idResearcher, idWriter,
			protocol.TextPayload(fmt.Sprintf("[Round %d] FYI — additional data point discovered for: %s", round, topic)),
		), randomDelay(1000, 2000))
	}

	// Phase 7: Broadcast — publisher announces completion
	sendWithDelay(ctx, engine, protocol.NewBroadcast(
		idPublisher,
		protocol.TextPayload(fmt.Sprintf("[Round %d] 📢 Report published: \"%s\"", round, topic)),
	), randomDelay(4000, 4800))
}

// ---------------------------------------------------------------------------
// Agent Handlers
// ---------------------------------------------------------------------------

func coordinatorHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s", a.Name(), msg.Type, msg.From)

	switch msg.Type {
	case protocol.TypeResponse:
		// Acknowledge responses from other agents.
		time.Sleep(randomDelay(30, 80))
		ack := protocol.NewMessage(
			protocol.TypeMessage,
			a.ID(), msg.From,
			protocol.TextPayload("Acknowledged. Updating project status."),
		)
		return []*protocol.Message{ack}, nil

	case protocol.TypeMessage:
		// Broadcast announcements are just noted.
		log.Printf("[%s] noted: %s", a.Name(), msg.Payload.Content)
		return nil, nil

	default:
		return nil, nil
	}
}

func researcherHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s", a.Name(), msg.Type, msg.From)

	switch msg.Type {
	case protocol.TypeRequest:
		// Simulate research work.
		time.Sleep(randomDelay(80, 200))

		findings := []string{
			"Key finding: distributed agent systems show 3x throughput improvement",
			"Analysis reveals emergent coordination patterns in agent swarms",
			"Data suggests event-driven messaging reduces latency by 40%",
			"Research indicates hierarchical agent topologies outperform flat ones for complex tasks",
			"Studies confirm that real-time visualization improves agent system debugging efficiency by 60%",
		}

		response := protocol.NewResponse(
			a.ID(), msg.From, msg.ID,
			protocol.TextPayload(fmt.Sprintf("Research complete. %s. Confidence: %.0f%%.",
				findings[rand.Intn(len(findings))],
				70+rand.Float64()*30,
			)),
		)
		return []*protocol.Message{response}, nil

	case protocol.TypeMessage:
		log.Printf("[%s] noted message from %s", a.Name(), msg.From)
		return nil, nil

	default:
		return nil, nil
	}
}

func writerHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s", a.Name(), msg.Type, msg.From)

	switch msg.Type {
	case protocol.TypeRequest:
		// Simulate writing work.
		time.Sleep(randomDelay(100, 250))

		sections := []string{
			"Draft section complete: Introduction and methodology outlined with 3 key arguments.",
			"Report section drafted: comparative analysis with 5 case studies included.",
			"Content prepared: executive summary with key metrics and visual data references.",
			"Section written: detailed findings with supporting evidence from 8 sources.",
		}

		response := protocol.NewResponse(
			a.ID(), msg.From, msg.ID,
			protocol.TextPayload(sections[rand.Intn(len(sections))]),
		)
		return []*protocol.Message{response}, nil

	case protocol.TypeMessage:
		// Acknowledge informational messages from researcher.
		if msg.From == idResearcher {
			time.Sleep(randomDelay(20, 50))
			ack := protocol.NewMessage(
				protocol.TypeMessage,
				a.ID(), msg.From,
				protocol.TextPayload("Thanks for the additional data. Incorporating into the draft."),
			)
			return []*protocol.Message{ack}, nil
		}
		return nil, nil

	default:
		return nil, nil
	}
}

func reviewerHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s", a.Name(), msg.Type, msg.From)

	switch msg.Type {
	case protocol.TypeRequest:
		// Simulate review work.
		time.Sleep(randomDelay(60, 150))

		verdicts := []string{
			"Review passed. Quality score: 9.2/10. Minor suggestions noted in metadata.",
			"Approved with comments. Strong analysis, recommend expanding the conclusion.",
			"Review complete. Excellent research depth. Ready for publishing.",
			"Approved. Methodology is sound. Data presentation could be improved slightly.",
		}

		response := protocol.NewResponse(
			a.ID(), msg.From, msg.ID,
			protocol.TextPayload(verdicts[rand.Intn(len(verdicts))]),
		)
		return []*protocol.Message{response}, nil

	case protocol.TypeMessage:
		log.Printf("[%s] noted: %s", a.Name(), msg.Payload.Content)
		return nil, nil

	default:
		return nil, nil
	}
}

func publisherHandler(ctx context.Context, a *agent.Agent, msg *protocol.Message) ([]*protocol.Message, error) {
	log.Printf("[%s] received %s from %s", a.Name(), msg.Type, msg.From)

	switch msg.Type {
	case protocol.TypeRequest:
		// Simulate publishing work.
		time.Sleep(randomDelay(50, 120))

		response := protocol.NewResponse(
			a.ID(), msg.From, msg.ID,
			protocol.TextPayload("Published successfully. Distribution: internal + external channels. Archive ID generated."),
		)
		return []*protocol.Message{response}, nil

	case protocol.TypeMessage:
		log.Printf("[%s] noted: %s", a.Name(), msg.Payload.Content)
		return nil, nil

	default:
		return nil, nil
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func sendWithDelay(ctx context.Context, engine *runtime.Engine, msg *protocol.Message, delay time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(delay):
	}

	if err := engine.MessageBus.Send(msg); err != nil {
		log.Printf("[demo] failed to send message: %v", err)
	}
}

func randomDelay(minMs, maxMs int) time.Duration {
	return time.Duration(minMs+rand.Intn(maxMs-minMs)) * time.Millisecond
}
