// ---------------------------------------------------------------------------
// DashboardBar — top stats bar
// ---------------------------------------------------------------------------

import { useAgentStore, type ConnectionStatus } from "@/stores/agentStore.ts";

const STATUS_COLORS: Record<ConnectionStatus, string> = {
  connected: "#3FB950",
  connecting: "#D29922",
  disconnected: "#F85149",
  error: "#F85149",
};

const STATUS_LABELS: Record<ConnectionStatus, string> = {
  connected: "Connected",
  connecting: "Connecting…",
  disconnected: "Disconnected",
  error: "Error",
};

interface StatItemProps {
  label: string;
  value: string | number;
  color?: string;
}

function StatItem({ label, value, color }: StatItemProps) {
  return (
    <div style={{ textAlign: "center", minWidth: 80 }}>
      <div
        style={{
          fontSize: 28,
          fontWeight: 700,
          fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
          color: color ?? "#E6EDF3",
          lineHeight: 1.2,
        }}
      >
        {value}
      </div>
      <div style={{ fontSize: 11, color: "#8B949E", textTransform: "uppercase", letterSpacing: 1 }}>
        {label}
      </div>
    </div>
  );
}

export function DashboardBar() {
  const agents = useAgentStore((s) => s.agents);
  const messages = useAgentStore((s) => s.messages);
  const connectionStatus = useAgentStore((s) => s.connectionStatus);

  // Compute stats from live agents map.
  const agentList = Array.from(agents.values());
  const agentCount = agentList.length;
  const activeAgents = agentList.filter((a) => a.status === "busy").length;
  const errorAgents = agentList.filter((a) => a.status === "error").length;
  const messageCount = messages.length;

  return (
    <header
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "12px 24px",
        background: "#161B22",
        borderBottom: "1px solid #30363D",
      }}
    >
      {/* Left: Title */}
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <span style={{ fontSize: 18, fontWeight: 700, fontFamily: "'JetBrains Mono', monospace" }}>
          🚀 Mini-Agent
        </span>
        <span
          style={{
            fontSize: 11,
            padding: "2px 8px",
            borderRadius: 9999,
            background: STATUS_COLORS[connectionStatus] + "20",
            color: STATUS_COLORS[connectionStatus],
            fontWeight: 600,
          }}
        >
          {STATUS_LABELS[connectionStatus]}
        </span>
      </div>

      {/* Right: Stats */}
      <div style={{ display: "flex", gap: 32 }}>
        <StatItem label="Agents" value={agentCount} color="#58A6FF" />
        <StatItem label="Active" value={activeAgents} color="#3FB950" />
        <StatItem label="Messages" value={messageCount} color="#4FC3F7" />
        <StatItem label="Errors" value={errorAgents} color={errorAgents > 0 ? "#F85149" : "#8B949E"} />
      </div>
    </header>
  );
}
