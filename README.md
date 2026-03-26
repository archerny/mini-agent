# mini-agent

> AI Agent 网络的实时操作系统 — 多 Agent 协调引擎 + 实时可视化作战大屏

一个自研的多 Agent 协调引擎（Go Runtime）+ 实时可视化作战大屏（React），用于观测、理解和管理动态 Agent 协作网络。

## Architecture

```
┌─────────────────────────────────────────┐
│            React Frontend               │
│  Network Topology │ Message Timeline    │
│     (60%)         │    (40%)            │
├─────────────────────────────────────────┤
│  Agent Detail Panel                     │
└──────────────────┬──────────────────────┘
                   │ WebSocket + REST
┌──────────────────┴──────────────────────┐
│            Go Backend                   │
│  Agent Manager │ MessageBus │ EventStream│
│  Agent A  Agent B  Agent C              │
└─────────────────────────────────────────┘
```

**Key design decisions:**
- **Protocol First** — Agent Communication Protocol is the system's constitution
- **Agent = struct + goroutine** — stateful entities that can sleep/resume
- **MessageBus = Single Source of Events** — agents don't know about events

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Protocol | JSON-based Agent Communication Protocol |
| Runtime | Go 1.24+, goroutines, channels |
| API | net/http (stdlib), gorilla/websocket |
| Frontend | React 19, TypeScript, Vite |
| Visualization | React Flow / D3-force |
| Styling | Tailwind CSS |
| State | Zustand |

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
│   ├── api/                    # REST + WebSocket (M2)
│   └── demo/                   # Demo scenarios (M4)
├── web/                        # React frontend
│   └── src/
│       ├── protocol/types.ts   # TypeScript protocol types (mirrors Go)
│       ├── App.tsx
│       └── main.tsx
├── Makefile
└── README.md
```

## Getting Started

```bash
# Build & run backend
make run

# Run frontend dev server (separate terminal)
make dev-web

# Run both (development)
make dev

# Run tests
make test

# TypeScript type check
make typecheck
```

## Milestones

- [x] **M0: Protocol Spec** — Protocol type definitions (Go + TypeScript)
- [ ] **M1: Minimal Runtime** — 2 agents exchanging messages
- [ ] **M2: WebSocket Bridge** — Frontend receives real-time events
- [ ] **M3a: Basic Dashboard** — Dark theme skeleton + agent list + message log
- [ ] **M3b: Advanced Visualization** — Network topology + particle animations
- [ ] **M4: Demo Scenario** — Multi-agent collaboration demo

## License

MIT
