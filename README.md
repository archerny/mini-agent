# mini-agent

> AI Agent 网络的实时操作系统 — 多 Agent 协调引擎 + 实时可视化作战大屏

一个自研的多 Agent 协调引擎（Go Runtime）+ 实时可视化作战大屏（React），用于观测、理解和**交互控制**动态 Agent 协作网络。支持单二进制部署，前端嵌入 Go Binary。

## Architecture

```
┌─────────────────────────────────────────┐
│            React Frontend               │
│  Network Topology │ Message Timeline    │
│     (60%)         │  (40%, filterable)  │
├─────────────────────────────────────────┤
│  Control Panel (Send Msg / Spawn Agent) │
├─────────────────────────────────────────┤
│  Agent Detail Panel (if selected)       │
└──────────────────┬──────────────────────┘
                   │ WebSocket + REST
┌──────────────────┴──────────────────────┐
│            Go Backend                   │
│  Agent Manager │ MessageBus │ EventStream│
│  Agent A  Agent B  Agent C              │
│  ──── embed.FS (frontend) ────          │
└─────────────────────────────────────────┘
```

**Key design decisions:**
- **Protocol First** — Agent Communication Protocol is the system's constitution
- **Agent = struct + goroutine** — stateful entities that can sleep/resume
- **MessageBus = Single Source of Events** — agents don't know about events
- **Single Binary** — `go:embed` bundles frontend into one deployable binary

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Protocol | JSON-based Agent Communication Protocol |
| Runtime | Go 1.24+, goroutines, channels |
| API | net/http (stdlib), gorilla/websocket |
| Frontend | React 19, TypeScript, Vite |
| Visualization | React Flow (network topology + particle edges) |
| Styling | Inline styles (dark theme, JetBrains Mono) |
| State | Zustand |
| Deployment | `go:embed` — single binary with SPA |

## Project Structure

```
mini-agent/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── protocol/               # Agent Communication Protocol types
│   │   ├── types.go            # States, enums, constants
│   │   ├── message.go          # Message format + validation
│   │   ├── event.go            # Event format + constructors
│   │   ├── agent_card.go       # Agent Card + Topology + Stats
│   │   └── uid/uid.go          # UUID v7 generator
│   ├── agent/                  # Agent lifecycle (M1)
│   ├── runtime/                # MessageBus, EventStream (M1)
│   ├── api/                    # REST + WebSocket (M2, M5)
│   │   ├── handler.go          # Read + Write API handlers
│   │   ├── router.go           # Route registration + CORS + SPA serving
│   │   └── websocket.go        # Real-time event streaming
│   └── demo/                   # Demo scenarios (M4)
│       └── scenario.go         # Research Pipeline — 5-agent collaboration
├── web/                        # React frontend
│   ├── embed.go                # go:embed wrapper (single-binary support)
│   ├── embed_prod.go           # Build tag: embedfrontend → embed dist/
│   ├── embed_dev.go            # Build tag: default → nil (dev mode)
│   └── src/
│       ├── api/client.ts       # Typed REST API client
│       ├── protocol/
│       │   ├── types.ts        # TypeScript protocol types (mirrors Go)
│       │   └── constants.ts    # Shared UI constants (colors, labels)
│       ├── components/
│       │   ├── ControlPanel.tsx # Interactive: send messages, spawn/shutdown agents
│       │   ├── MessageLog.tsx   # Filterable message timeline
│       │   ├── AgentDetailPanel.tsx
│       │   ├── DashboardBar.tsx
│       │   └── topology/       # React Flow visualization
│       ├── hooks/              # useWebSocket, etc.
│       ├── stores/             # Zustand stores
│       ├── App.tsx
│       └── main.tsx
├── Makefile
└── README.md
```

## Getting Started

```bash
# --- Development (two processes) ---

# Backend
make run

# Frontend dev server (separate terminal, with HMR + proxy)
make dev-web

# Or run both at once
make dev

# --- Production (single binary) ---

# Build frontend + embed into Go binary
make build-all

# Run the single binary (serves API + frontend on :8080)
make run-all

# --- Utilities ---

# Run Go tests
make test

# TypeScript type check
make typecheck
```

## REST API

### Read Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/agents` | List all agents |
| `GET` | `/api/agents/{id}` | Get a single agent |
| `GET` | `/api/messages` | Get message history (?limit=N) |
| `GET` | `/api/events` | Get event history (?since_sequence=N&limit=N) |
| `GET` | `/api/topology` | Get network topology (nodes + edges) |
| `GET` | `/api/stats` | Get global statistics |

### Write Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/agents` | Spawn a new agent |
| `DELETE` | `/api/agents/{id}` | Shutdown an agent |
| `POST` | `/api/messages` | Send a message between agents |

### Demo Control

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/demo/pause` | Pause the demo scenario |
| `POST` | `/api/demo/resume` | Resume the demo scenario |
| `GET` | `/api/demo/status` | Get demo running/paused status |

### WebSocket

| Path | Description |
|------|-------------|
| `WS /ws/events` | Real-time event stream (auto-reconnect) |

## Demo Scenario

The built-in demo runs a **Research Pipeline** with 5 agents collaborating:

| Agent | Role | Capabilities |
|-------|------|-------------|
| **coordinator** | Project Coordinator | orchestrate, delegate, aggregate |
| **researcher** | Research Agent | web_search, analyze, summarize |
| **writer** | Writer Agent | write, format, revise |
| **reviewer** | Reviewer Agent | review, critique, approve |
| **publisher** | Publisher Agent | publish, distribute, archive |

**Collaboration flow:**
```
coordinator ──request──► researcher ──response──► coordinator
coordinator ──request──► writer     ──response──► coordinator
coordinator ──request──► reviewer   ──response──► coordinator
coordinator ──request──► publisher  ──response──► coordinator
reviewer    ──request──► writer     (feedback loop, every 3rd round)
researcher  ──message──► writer     (data sharing, every 2nd round)
publisher   ──broadcast──► *        (completion announcement)
```

Every 5 seconds a new research round starts with a rotating topic. The topology view shows all agents, their real-time state changes, and message flow with particle animations.

## Milestones

- [x] **M0: Protocol Spec** — Protocol type definitions (Go + TypeScript)
- [x] **M1: Minimal Runtime** — 2 agents exchanging messages
- [x] **M2: WebSocket Bridge** — Frontend receives real-time events
- [x] **M3a: Basic Dashboard** — Dark theme skeleton + agent list + message log
- [x] **M3b: Advanced Visualization** — Network topology + particle animations
- [x] **M4: Demo Scenario** — Multi-agent collaboration demo
- [x] **M5: Interactive Control + Production Polish** — Write APIs, control panel, message filtering, single-binary deploy

## License

MIT
