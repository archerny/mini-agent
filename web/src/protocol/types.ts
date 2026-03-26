// ---------------------------------------------------------------------------
// Agent Communication Protocol — TypeScript Type Definitions
//
// These types mirror the Go definitions in internal/protocol/*.go.
// Keep them in sync manually (MVP). Future: consider code generation.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Agent States
// ---------------------------------------------------------------------------

export type AgentState =
  | "spawning"
  | "ready"
  | "idle"
  | "busy"
  | "completed"
  | "error"
  | "shutdown";

/** Complete state transition table — matches Go ValidTransitions. */
export const VALID_TRANSITIONS: Record<AgentState, AgentState[]> = {
  spawning: ["ready", "error"],
  ready: ["idle"],
  idle: ["busy", "error", "shutdown"],
  busy: ["completed", "error"],
  completed: ["idle"],
  error: ["idle", "shutdown"],
  shutdown: [], // terminal
};

/** Check whether a state transition is valid. */
export function canTransition(from: AgentState, to: AgentState): boolean {
  return VALID_TRANSITIONS[from]?.includes(to) ?? false;
}

/** Terminal states — no further transitions. */
export function isTerminal(state: AgentState): boolean {
  return state === "shutdown";
}

/** Transient states — auto-transition to next state. */
export function isTransient(state: AgentState): boolean {
  return state === "ready" || state === "completed";
}

// ---------------------------------------------------------------------------
// Message Types
// ---------------------------------------------------------------------------

export type MessageType =
  | "agent.message"
  | "agent.request"
  | "agent.response"
  | "agent.broadcast";

// ---------------------------------------------------------------------------
// Content Types
// ---------------------------------------------------------------------------

export type ContentType = "text" | "json" | "tool_call" | "tool_result";

// ---------------------------------------------------------------------------
// Event Types
// ---------------------------------------------------------------------------

export type EventType =
  | "agent.spawned"
  | "agent.state_changed"
  | "agent.message_sent"
  | "agent.message_received"
  | "agent.error"
  | "agent.shutdown"
  | "network.topology_changed";

// ---------------------------------------------------------------------------
// Error Kinds
// ---------------------------------------------------------------------------

export type ErrorKind =
  | "undeliverable"
  | "backpressure"
  | "invalid_message"
  | "payload_too_large"
  | "panic"
  | "timeout"
  | "internal";

// ---------------------------------------------------------------------------
// Topology Change Types
// ---------------------------------------------------------------------------

export type TopologyChangeType =
  | "agent_joined"
  | "agent_left"
  | "link_created"
  | "link_removed";

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** Broadcast target sentinel value. */
export const BROADCAST_TARGET = "*";

/** Maximum payload size in bytes (1MB). */
export const MAX_PAYLOAD_SIZE = 1 << 20;

/** Maximum broadcasts per agent per second. */
export const BROADCAST_RATE_LIMIT = 10;

// ---------------------------------------------------------------------------
// Message
// ---------------------------------------------------------------------------

export interface Payload {
  content_type: ContentType;
  content: string;
}

export interface Message {
  id: string;
  type: MessageType;
  from: string;
  to: string;
  correlation_id?: string;
  timestamp: string; // ISO 8601
  payload: Payload;
  metadata?: Record<string, unknown>;
}

// ---------------------------------------------------------------------------
// Event
// ---------------------------------------------------------------------------

export interface Event<T = unknown> {
  id: string;
  type: EventType;
  sequence: number;
  agent_id?: string;
  timestamp: string; // ISO 8601
  data: T;
  metadata?: Record<string, unknown>;
}

// ---------------------------------------------------------------------------
// Event Data Types — one per EventType
// ---------------------------------------------------------------------------

export interface AgentSpawnedData {
  name: string;
  role: string;
  capabilities: string[];
}

export interface AgentStateChangedData {
  prev_state: AgentState;
  new_state: AgentState;
  reason?: string;
}

export interface AgentMessageSentData {
  message: Message;
}

export interface AgentMessageReceivedData {
  message: Message;
}

export interface AgentErrorData {
  error: string;
  kind: ErrorKind;
  recoverable: boolean;
  stack?: string;
}

export interface AgentShutdownData {
  reason: string;
  exit_code: number;
}

export interface TopologyChangedData {
  change_type: TopologyChangeType;
  details?: Record<string, unknown>;
}

// ---------------------------------------------------------------------------
// Typed Event Aliases (for convenience)
// ---------------------------------------------------------------------------

export type AgentSpawnedEvent = Event<AgentSpawnedData>;
export type AgentStateChangedEvent = Event<AgentStateChangedData>;
export type AgentMessageSentEvent = Event<AgentMessageSentData>;
export type AgentMessageReceivedEvent = Event<AgentMessageReceivedData>;
export type AgentErrorEvent = Event<AgentErrorData>;
export type AgentShutdownEvent = Event<AgentShutdownData>;
export type TopologyChangedEvent = Event<TopologyChangedData>;

// ---------------------------------------------------------------------------
// Agent Card
// ---------------------------------------------------------------------------

export interface AgentCard {
  id: string;
  name: string;
  role: string;
  capabilities: string[];
  accepts: MessageType[];
  status: AgentState;
  metadata?: Record<string, unknown>;
}

// ---------------------------------------------------------------------------
// Topology (REST API response)
// ---------------------------------------------------------------------------

export interface TopologyEdge {
  from: string;
  to: string;
  message_count: number;
}

export interface Topology {
  nodes: AgentCard[];
  edges: TopologyEdge[];
}

// ---------------------------------------------------------------------------
// Stats (REST API response)
// ---------------------------------------------------------------------------

export interface Stats {
  agent_count: number;
  message_count: number;
  active_agents: number;
  error_count: number;
  uptime: number; // seconds
}
