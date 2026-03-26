// ---------------------------------------------------------------------------
// AgentDetailPanel — bottom slide-up panel showing selected agent details
// ---------------------------------------------------------------------------

import { useAgentStore } from "@/stores/agentStore.ts";
import { STATUS_COLORS, STATUS_LABELS } from "@/protocol/constants.ts";

export function AgentDetailPanel() {
  const agents = useAgentStore((s) => s.agents);
  const selectedAgentId = useAgentStore((s) => s.selectedAgentId);
  const selectAgent = useAgentStore((s) => s.selectAgent);
  const messages = useAgentStore((s) => s.messages);

  if (!selectedAgentId || !agents.has(selectedAgentId)) return null;

  const agent = agents.get(selectedAgentId)!;
  const statusColor = STATUS_COLORS[agent.status] ?? "#8B949E";

  // Count messages involving this agent
  const agentMessages = messages.filter(
    (m) => m.from === agent.id || m.to === agent.id,
  );
  const sentCount = messages.filter((m) => m.from === agent.id).length;
  const receivedCount = messages.filter((m) => m.to === agent.id).length;

  return (
    <div
      style={{
        borderTop: "1px solid #30363D",
        background: "#161B22",
        padding: "12px 24px",
        flexShrink: 0,
        display: "flex",
        gap: 32,
        alignItems: "center",
        animation: "slideUp 0.2s ease-out",
      }}
    >
      {/* Agent identity */}
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        {/* Avatar */}
        <div
          style={{
            width: 44,
            height: 44,
            borderRadius: "50%",
            border: `3px solid ${statusColor}`,
            background: `${statusColor}18`,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: 20,
            fontWeight: 700,
            color: "#E6EDF3",
            fontFamily: "'JetBrains Mono', monospace",
          }}
        >
          {agent.name.charAt(0).toUpperCase()}
        </div>

        <div>
          <div
            style={{
              fontSize: 15,
              fontWeight: 700,
              color: "#E6EDF3",
              fontFamily: "'JetBrains Mono', monospace",
            }}
          >
            {agent.name}
          </div>
          <div style={{ fontSize: 11, color: "#8B949E" }}>{agent.role}</div>
        </div>
      </div>

      {/* Status */}
      <div>
        <div style={{ fontSize: 10, color: "#8B949E", textTransform: "uppercase", marginBottom: 2 }}>
          Status
        </div>
        <span
          style={{
            fontSize: 11,
            fontWeight: 700,
            padding: "3px 10px",
            borderRadius: 9999,
            background: `${statusColor}20`,
            color: statusColor,
            textTransform: "uppercase",
            letterSpacing: 0.5,
          }}
        >
          {STATUS_LABELS[agent.status] ?? agent.status}
        </span>
      </div>

      {/* Message stats */}
      <div style={{ display: "flex", gap: 16 }}>
        <StatMini label="Total" value={agentMessages.length} />
        <StatMini label="Sent" value={sentCount} color="#4FC3F7" />
        <StatMini label="Recv" value={receivedCount} color="#81C784" />
      </div>

      {/* Capabilities */}
      <div style={{ flex: 1 }}>
        <div style={{ fontSize: 10, color: "#8B949E", textTransform: "uppercase", marginBottom: 4 }}>
          Capabilities
        </div>
        <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
          {(agent.capabilities ?? []).map((cap) => (
            <span
              key={cap}
              style={{
                display: "inline-block",
                padding: "2px 8px",
                borderRadius: 4,
                background: "#21262D",
                fontSize: 11,
                color: "#58A6FF",
                fontFamily: "'JetBrains Mono', monospace",
              }}
            >
              {cap}
            </span>
          ))}
        </div>
      </div>

      {/* ID */}
      <div>
        <div style={{ fontSize: 10, color: "#8B949E", textTransform: "uppercase", marginBottom: 2 }}>
          ID
        </div>
        <div
          style={{
            fontSize: 10,
            fontFamily: "'JetBrains Mono', monospace",
            color: "#484F58",
            maxWidth: 160,
            overflow: "hidden",
            textOverflow: "ellipsis",
          }}
        >
          {agent.id}
        </div>
      </div>

      {/* Close button */}
      <button
        onClick={() => selectAgent(null)}
        style={{
          background: "none",
          border: "1px solid #30363D",
          borderRadius: 6,
          color: "#8B949E",
          cursor: "pointer",
          padding: "4px 8px",
          fontSize: 12,
        }}
        aria-label="Close agent detail"
      >
        ✕
      </button>

      <style>{`
        @keyframes slideUp {
          from { transform: translateY(100%); opacity: 0; }
          to { transform: translateY(0); opacity: 1; }
        }
      `}</style>
    </div>
  );
}

// ---------------------------------------------------------------------------
// StatMini — small stat display for the detail panel
// ---------------------------------------------------------------------------

function StatMini({ label, value, color }: { label: string; value: number; color?: string }) {
  return (
    <div style={{ textAlign: "center" }}>
      <div
        style={{
          fontSize: 16,
          fontWeight: 700,
          fontFamily: "'JetBrains Mono', monospace",
          color: color ?? "#E6EDF3",
        }}
      >
        {value}
      </div>
      <div style={{ fontSize: 9, color: "#8B949E", textTransform: "uppercase" }}>{label}</div>
    </div>
  );
}
