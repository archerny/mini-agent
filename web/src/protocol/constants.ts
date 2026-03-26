// ---------------------------------------------------------------------------
// Shared UI constants — status colors, labels, etc.
// Single source of truth for all components.
// ---------------------------------------------------------------------------

import type { AgentState } from "@/protocol/types.ts";

export const STATUS_COLORS: Record<AgentState, string> = {
  spawning: "#8B949E",
  ready: "#8B949E",
  idle: "#58A6FF",
  busy: "#3FB950",
  completed: "#A371F7",
  error: "#F85149",
  shutdown: "#484F58",
};

export const STATUS_LABELS: Record<AgentState, string> = {
  spawning: "Spawning",
  ready: "Ready",
  idle: "Idle",
  busy: "Active",
  completed: "Completed",
  error: "Error",
  shutdown: "Shutdown",
};
