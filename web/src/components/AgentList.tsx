// ---------------------------------------------------------------------------
// AgentList — left panel showing agent cards with status
// ---------------------------------------------------------------------------

import { useAgentStore } from "@/stores/agentStore.ts";
import type { AgentState } from "@/protocol/types.ts";

const STATUS_COLORS: Record<AgentState, string> = {
  spawning: "#8B949E",
  ready: "#8B949E",
  idle: "#58A6FF",
  busy: "#3FB950",
  completed: "#A371F7",
  error: "#F85149",
  shutdown: "#484F58",
};

const STATUS_LABELS: Record<AgentState, string> = {
  spawning: "Spawning",
  ready: "Ready",
  idle: "Idle",
  busy: "Active",
  completed: "Completed",
  error: "Error",
  shutdown: "Shutdown",
};

export function AgentList() {
  const agents = useAgentStore((s) => s.agents);
  const selectedAgentId = useAgentStore((s) => s.selectedAgentId);
  const selectAgent = useAgentStore((s) => s.selectAgent);

  const agentList = Array.from(agents.values());

  if (agentList.length === 0) {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          height: "100%",
          color: "#484F58",
          fontSize: 13,
          fontStyle: "italic",
        }}
      >
        Waiting for agents…
      </div>
    );
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 8, padding: 12 }}>
      <div
        style={{
          fontSize: 13,
          fontWeight: 600,
          color: "#8B949E",
          textTransform: "uppercase",
          letterSpacing: 1,
          marginBottom: 4,
        }}
      >
        Agents ({agentList.length})
      </div>
      {agentList.map((agent) => {
        const isSelected = selectedAgentId === agent.id;
        const statusColor = STATUS_COLORS[agent.status] ?? "#8B949E";

        return (
          <div
            key={agent.id}
            onClick={() => selectAgent(isSelected ? null : agent.id)}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 12,
              padding: "10px 12px",
              borderRadius: 8,
              border: `1px solid ${isSelected ? statusColor : "#30363D"}`,
              background: isSelected ? statusColor + "10" : "#161B22",
              cursor: "pointer",
              transition: "all 0.15s",
            }}
          >
            {/* Status dot */}
            <div
              style={{
                width: 36,
                height: 36,
                borderRadius: "50%",
                border: `2px solid ${statusColor}`,
                background: statusColor + "20",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontSize: 14,
                fontWeight: 700,
                color: "#E6EDF3",
                fontFamily: "'JetBrains Mono', monospace",
                flexShrink: 0,
              }}
            >
              {agent.name.charAt(0).toUpperCase()}
            </div>

            {/* Info */}
            <div style={{ flex: 1, minWidth: 0 }}>
              <div
                style={{
                  fontSize: 14,
                  fontWeight: 600,
                  color: "#E6EDF3",
                  marginBottom: 2,
                  fontFamily: "'JetBrains Mono', monospace",
                }}
              >
                {agent.name}
              </div>
              <div
                style={{
                  fontSize: 11,
                  color: "#8B949E",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {agent.role}
              </div>
            </div>

            {/* Status badge */}
            <span
              style={{
                fontSize: 10,
                fontWeight: 600,
                padding: "2px 8px",
                borderRadius: 9999,
                background: statusColor + "20",
                color: statusColor,
                textTransform: "uppercase",
                letterSpacing: 0.5,
                flexShrink: 0,
              }}
            >
              {STATUS_LABELS[agent.status] ?? agent.status}
            </span>
          </div>
        );
      })}

      {/* Agent detail panel (if selected) */}
      {selectedAgentId && agents.has(selectedAgentId) && (
        <AgentDetail agent={agents.get(selectedAgentId)!} />
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// AgentDetail — expanded detail for a selected agent
// ---------------------------------------------------------------------------

import type { AgentCard } from "@/protocol/types.ts";

function AgentDetail({ agent }: { agent: AgentCard }) {
  const statusColor = STATUS_COLORS[agent.status] ?? "#8B949E";

  return (
    <div
      style={{
        padding: 12,
        borderRadius: 8,
        border: `1px solid ${statusColor}40`,
        background: "#0D1117",
        fontSize: 12,
        color: "#8B949E",
      }}
    >
      <div style={{ marginBottom: 8, color: "#E6EDF3", fontWeight: 600, fontSize: 13 }}>
        {agent.name} — Detail
      </div>
      <div style={{ marginBottom: 4 }}>
        <strong>ID:</strong>{" "}
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 11 }}>{agent.id}</span>
      </div>
      <div style={{ marginBottom: 4 }}>
        <strong>Role:</strong> {agent.role}
      </div>
      <div style={{ marginBottom: 4 }}>
        <strong>Capabilities:</strong>{" "}
        {(agent.capabilities ?? []).map((cap) => (
          <span
            key={cap}
            style={{
              display: "inline-block",
              padding: "1px 6px",
              borderRadius: 4,
              background: "#21262D",
              marginRight: 4,
              fontSize: 11,
              color: "#58A6FF",
            }}
          >
            {cap}
          </span>
        ))}
      </div>
      <div>
        <strong>Status:</strong>{" "}
        <span style={{ color: statusColor, fontWeight: 600 }}>
          {STATUS_LABELS[agent.status] ?? agent.status}
        </span>
      </div>
    </div>
  );
}
