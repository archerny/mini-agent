// ---------------------------------------------------------------------------
// App — M3b Advanced Visualization layout
//
// Architecture:
// ┌─────────────────────────────────────────┐
// │            DashboardBar                 │
// ├───────────────────┬─────────────────────┤
// │  Network Topology │  Message Timeline   │
// │     (60%)         │    (40%)            │
// ├───────────────────┴─────────────────────┤
// │  Agent Detail Panel (if selected)       │
// └─────────────────────────────────────────┘
// ---------------------------------------------------------------------------

import { DashboardBar } from "@/components/DashboardBar.tsx";
import { NetworkTopology } from "@/components/topology/NetworkTopology.tsx";
import { MessageLog } from "@/components/MessageLog.tsx";
import { AgentDetailPanel } from "@/components/AgentDetailPanel.tsx";
import { ControlPanel } from "@/components/ControlPanel.tsx";
import { useWebSocket } from "@/hooks/useWebSocket.ts";
import { useAgentStore } from "@/stores/agentStore.ts";

function App() {
  // Establish WebSocket connection.
  useWebSocket();

  const connectionStatus = useAgentStore((s) => s.connectionStatus);
  const selectedAgentId = useAgentStore((s) => s.selectedAgentId);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        background: "#0D1117",
        color: "#E6EDF3",
        fontFamily: "'Inter', system-ui, sans-serif",
        overflow: "hidden",
      }}
    >
      {/* Disconnected banner */}
      {connectionStatus === "disconnected" && (
        <div
          style={{
            padding: "6px 16px",
            background: "#D2992220",
            color: "#D29922",
            fontSize: 12,
            textAlign: "center",
            borderBottom: "1px solid #D2992240",
            flexShrink: 0,
          }}
        >
          ⚠️ Real-time connection lost. Reconnecting…
        </div>
      )}

      {/* Top: Dashboard Bar */}
      <DashboardBar />

      {/* Middle: Topology (60%) + Messages (40%) */}
      <div
        style={{
          display: "flex",
          flex: 1,
          minHeight: 0,
        }}
      >
        {/* Left: Network Topology (60%) */}
        <div
          style={{
            width: "60%",
            minWidth: 400,
            borderRight: "1px solid #30363D",
            position: "relative",
          }}
          aria-label="Network topology"
        >
          <NetworkTopology />
        </div>

        {/* Right: Message Log (40%) */}
        <main
          style={{
            flex: 1,
            minWidth: 0,
            display: "flex",
            flexDirection: "column",
          }}
          aria-label="Message timeline"
          role="log"
        >
          <MessageLog />
        </main>
      </div>

      {/* Bottom: Control Panel */}
      <ControlPanel />

      {/* Bottom: Agent Detail Panel (when agent is selected) */}
      {selectedAgentId && <AgentDetailPanel />}
    </div>
  );
}

export default App;
