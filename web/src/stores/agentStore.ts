// ---------------------------------------------------------------------------
// Zustand Store — single source of truth for the dashboard
// ---------------------------------------------------------------------------

import { create } from "zustand";
import type {
  AgentCard,
  Message,
  Event,
  Stats,
  Topology,
  EventType,
  AgentSpawnedData,
  AgentStateChangedData,
  AgentMessageSentData,
} from "@/protocol/types.ts";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type ConnectionStatus = "connecting" | "connected" | "disconnected" | "error";

interface AgentStore {
  // Data
  agents: Map<string, AgentCard>;
  messages: Message[];
  events: Array<Event>;
  stats: Stats;
  topology: Topology;
  lastSequence: number;

  // Connection
  connectionStatus: ConnectionStatus;

  // UI state
  selectedAgentId: string | null;

  // Actions — data
  handleEvent: (event: Event) => void;
  loadSnapshot: () => Promise<void>;
  setConnectionStatus: (status: ConnectionStatus) => void;

  // Actions — UI
  selectAgent: (id: string | null) => void;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const MAX_MESSAGES = 1000;
const MAX_EVENTS = 500;

const EMPTY_STATS: Stats = {
  agent_count: 0,
  message_count: 0,
  active_agents: 0,
  error_count: 0,
  uptime: 0,
};

const EMPTY_TOPOLOGY: Topology = { nodes: [], edges: [] };

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

export const useAgentStore = create<AgentStore>((set, get) => ({
  // Initial state
  agents: new Map(),
  messages: [],
  events: [],
  stats: EMPTY_STATS,
  topology: EMPTY_TOPOLOGY,
  lastSequence: 0,
  connectionStatus: "connecting",
  selectedAgentId: null,

  // -------------------------------------------------------------------------
  // handleEvent — process a single WebSocket event
  // -------------------------------------------------------------------------
  handleEvent: (event: Event) => {
    const state = get();
    const type = event.type as EventType;

    // Always update lastSequence and push to events history.
    const newEvents = [...state.events, event];
    if (newEvents.length > MAX_EVENTS) {
      newEvents.splice(0, newEvents.length - MAX_EVENTS);
    }

    const updates: Partial<AgentStore> = {
      events: newEvents,
      lastSequence: event.sequence,
    };

    switch (type) {
      case "agent.spawned": {
        const data = event.data as AgentSpawnedData;
        const agentId = event.agent_id ?? "";
        const newAgents = new Map(state.agents);
        newAgents.set(agentId, {
          id: agentId,
          name: data.name,
          role: data.role,
          capabilities: data.capabilities,
          accepts: ["agent.message", "agent.request", "agent.response", "agent.broadcast"],
          status: "idle", // just spawned → will quickly become idle
        });
        updates.agents = newAgents;
        break;
      }

      case "agent.state_changed": {
        const data = event.data as AgentStateChangedData;
        const agentId = event.agent_id ?? "";
        const newAgents = new Map(state.agents);
        const existing = newAgents.get(agentId);
        if (existing) {
          newAgents.set(agentId, { ...existing, status: data.new_state });
          updates.agents = newAgents;
        }
        break;
      }

      case "agent.message_sent": {
        const data = event.data as AgentMessageSentData;
        const newMessages = [...state.messages, data.message];
        if (newMessages.length > MAX_MESSAGES) {
          newMessages.splice(0, newMessages.length - MAX_MESSAGES);
        }
        updates.messages = newMessages;
        break;
      }

      case "agent.shutdown": {
        const agentId = event.agent_id ?? "";
        const newAgents = new Map(state.agents);
        const existing = newAgents.get(agentId);
        if (existing) {
          newAgents.set(agentId, { ...existing, status: "shutdown" });
          updates.agents = newAgents;
        }
        break;
      }

      case "agent.error": {
        // Error events are captured in events array, no special handling for M3a.
        break;
      }

      default:
        break;
    }

    set(updates);
  },

  // -------------------------------------------------------------------------
  // loadSnapshot — fetch REST API for initial state
  // -------------------------------------------------------------------------
  loadSnapshot: async () => {
    try {
      const [agentsRes, statsRes, topologyRes, messagesRes] = await Promise.all([
        fetch("/api/agents"),
        fetch("/api/stats"),
        fetch("/api/topology"),
        fetch("/api/messages?limit=100"),
      ]);

      const agentCards: AgentCard[] = await agentsRes.json();
      const stats: Stats = await statsRes.json();
      const topology: Topology = await topologyRes.json();
      const messages: Message[] = await messagesRes.json();

      const agentsMap = new Map<string, AgentCard>();
      for (const card of agentCards) {
        agentsMap.set(card.id, card);
      }

      set({
        agents: agentsMap,
        stats,
        topology,
        messages: messages ?? [],
      });
    } catch (err) {
      console.error("[store] Failed to load snapshot:", err);
    }
  },

  setConnectionStatus: (status) => set({ connectionStatus: status }),

  selectAgent: (id) => set({ selectedAgentId: id }),
}));
