// ---------------------------------------------------------------------------
// App — M3a Basic Dashboard layout
// ---------------------------------------------------------------------------

import { DashboardBar } from "@/components/DashboardBar.tsx";
import { AgentList } from "@/components/AgentList.tsx";
import { MessageLog } from "@/components/MessageLog.tsx";
import { useWebSocket } from "@/hooks/useWebSocket.ts";
import { useAgentStore } from "@/stores/agentStore.ts";

function App() {
  // Establish WebSocket connection.
  useWebSocket();

  const connectionStatus = useAgentStore((s) => s.connectionStatus);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        background: "#0D1117",
        color: "#E6EDF3",
        fontFamily: "'Inter', system-ui, sans-serif",
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
          }}
        >
          ⚠️ Real-time connection lost. Reconnecting…
        </div>
      )}

      {/* Top: Dashboard Bar */}
      <DashboardBar />

      {/* Main: Left (Agents) + Right (Messages) */}
      <div
        style={{
          display: "flex",
          flex: 1,
          minHeight: 0, // allow flex children to shrink
        }}
      >
        {/* Left panel: Agent list (40%) */}
        <aside
          style={{
            width: "40%",
            minWidth: 280,
            maxWidth: 480,
            borderRight: "1px solid #30363D",
            overflowY: "auto",
            background: "#0D1117",
          }}
          aria-label="Agent list"
        >
          <AgentList />
        </aside>

        {/* Right panel: Message log (60%) */}
        <main
          style={{
            flex: 1,
            minWidth: 0,
            background: "#0D1117",
          }}
          aria-label="Message timeline"
          role="log"
        >
          <MessageLog />
        </main>
      </div>
    </div>
  );
}

export default App;
