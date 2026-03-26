package protocol

// ---------------------------------------------------------------------------
// Agent Card — agent capability declaration
// ---------------------------------------------------------------------------

// AgentCard describes an agent's identity, role, capabilities, and current state.
// It is the public "business card" of an agent in the network.
type AgentCard struct {
	// ID is the unique identifier for this agent.
	ID string `json:"id"`

	// Name is the human-readable name (e.g., "researcher", "analyzer").
	Name string `json:"name"`

	// Role is a short description of what this agent does.
	Role string `json:"role"`

	// Capabilities lists what this agent can do (e.g., "web_search", "summarize").
	Capabilities []string `json:"capabilities"`

	// Accepts lists the message types this agent can handle.
	Accepts []MessageType `json:"accepts"`

	// Status is the current lifecycle state.
	Status AgentState `json:"status"`

	// Metadata is an extensible key-value bag.
	// Examples: "model": "gpt-4", "max_concurrent_tasks": 3
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Topology types (used by REST API)
// ---------------------------------------------------------------------------

// TopologyEdge represents a communication link between two agents.
type TopologyEdge struct {
	From         string `json:"from"`
	To           string `json:"to"`
	MessageCount int    `json:"message_count"`
}

// Topology represents the current network topology snapshot.
type Topology struct {
	Nodes []AgentCard    `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

// ---------------------------------------------------------------------------
// Stats (used by REST API)
// ---------------------------------------------------------------------------

// Stats represents global runtime statistics.
type Stats struct {
	AgentCount   int     `json:"agent_count"`
	MessageCount int     `json:"message_count"`
	ActiveAgents int     `json:"active_agents"`
	ErrorCount   int     `json:"error_count"`
	Uptime       float64 `json:"uptime"` // seconds
}
