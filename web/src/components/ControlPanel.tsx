// ---------------------------------------------------------------------------
// ControlPanel — interactive control panel for sending messages, spawning
// agents, and controlling the demo scenario.
// ---------------------------------------------------------------------------

import { useState, useEffect, useCallback } from "react";
import { useAgentStore } from "@/stores/agentStore.ts";
import { apiClient, type DemoStatus } from "@/api/client.ts";

// ---------------------------------------------------------------------------
// Main ControlPanel
// ---------------------------------------------------------------------------

export function ControlPanel() {
  const [activeTab, setActiveTab] = useState<"message" | "spawn">("message");

  const tabs: Array<{ key: "message" | "spawn"; label: string }> = [
    { key: "message", label: "Send Message" },
    { key: "spawn", label: "Spawn Agent" },
  ];

  return (
    <div
      style={{
        borderTop: "1px solid #30363D",
        background: "#161B22",
        flexShrink: 0,
      }}
    >
      {/* Tab bar + Demo controls */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "0 12px",
          borderBottom: "1px solid #21262D",
        }}
      >
        <div style={{ display: "flex", gap: 0 }}>
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              style={{
                padding: "8px 16px",
                fontSize: 11,
                fontWeight: 600,
                color: activeTab === tab.key ? "#58A6FF" : "#8B949E",
                background: "none",
                border: "none",
                borderBottom: activeTab === tab.key ? "2px solid #58A6FF" : "2px solid transparent",
                cursor: "pointer",
                textTransform: "uppercase",
                letterSpacing: 0.5,
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <DemoControls />
      </div>

      {/* Tab content */}
      <div style={{ padding: "10px 12px" }}>
        {activeTab === "message" ? <SendMessageForm /> : <SpawnAgentForm />}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// DemoControls — pause/resume demo scenario
// ---------------------------------------------------------------------------

function DemoControls() {
  const [demoStatus, setDemoStatus] = useState<DemoStatus | null>(null);

  const fetchStatus = useCallback(() => {
    apiClient.getDemoStatus().then(setDemoStatus).catch(console.error);
  }, []);

  useEffect(() => {
    fetchStatus();
    const timer = setInterval(fetchStatus, 5000);
    return () => clearInterval(timer);
  }, [fetchStatus]);

  const toggleDemo = async () => {
    if (!demoStatus) return;
    if (demoStatus.paused) {
      await apiClient.resumeDemo();
    } else {
      await apiClient.pauseDemo();
    }
    fetchStatus();
  };

  if (!demoStatus?.running) return null;

  return (
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <span style={{ fontSize: 10, color: "#8B949E", textTransform: "uppercase" }}>Demo</span>
      <button
        onClick={toggleDemo}
        style={{
          padding: "3px 10px",
          fontSize: 11,
          fontWeight: 600,
          borderRadius: 6,
          border: "1px solid #30363D",
          background: demoStatus.paused ? "#238636" : "#21262D",
          color: demoStatus.paused ? "#FFFFFF" : "#8B949E",
          cursor: "pointer",
        }}
      >
        {demoStatus.paused ? "▶ Resume" : "⏸ Pause"}
      </button>
    </div>
  );
}

// ---------------------------------------------------------------------------
// SendMessageForm
// ---------------------------------------------------------------------------

function SendMessageForm() {
  const agents = useAgentStore((s) => s.agents);
  const agentList = Array.from(agents.values()).filter((a) => a.status !== "shutdown");

  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [msgType, setMsgType] = useState("agent.message");
  const [content, setContent] = useState("");
  const [status, setStatus] = useState<string | null>(null);

  const needsTo = msgType !== "agent.broadcast";
  const canSend = !!from && !!content && (!needsTo || !!to);

  const send = async () => {
    if (!canSend) return;
    setStatus("sending...");
    try {
      const res = await apiClient.sendMessage({ type: msgType, from, to: needsTo ? to : "*", content });
      setStatus(`✓ Sent (${res.id.slice(0, 8)})`);
      setContent("");
      setTimeout(() => setStatus(null), 3000);
    } catch (err) {
      setStatus(`✗ ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  return (
    <div style={{ display: "flex", gap: 8, alignItems: "flex-end" }}>
      <Field label="Type" width={130}>
        <select value={msgType} onChange={(e) => setMsgType(e.target.value)} style={selectStyle}>
          <option value="agent.message">Message</option>
          <option value="agent.request">Request</option>
          <option value="agent.broadcast">Broadcast</option>
        </select>
      </Field>
      <Field label="From" width={130}>
        <select value={from} onChange={(e) => setFrom(e.target.value)} style={selectStyle}>
          <option value="">— select —</option>
          {agentList.map((a) => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
      </Field>
      {msgType !== "agent.broadcast" && (
        <Field label="To" width={130}>
          <select value={to} onChange={(e) => setTo(e.target.value)} style={selectStyle}>
            <option value="">— select —</option>
            {agentList.filter((a) => a.id !== from).map((a) => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>
        </Field>
      )}
      <Field label="Content" width={0} flex={1}>
        <input
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && send()}
          placeholder="Type a message..."
          style={inputStyle}
        />
      </Field>
      <button onClick={send} disabled={!canSend} style={btnPrimary(!canSend)}>
        Send
      </button>
      {status && (
        <span style={{ fontSize: 10, color: status.startsWith("✓") ? "#3FB950" : "#F85149", whiteSpace: "nowrap" }}>
          {status}
        </span>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// SpawnAgentForm
// ---------------------------------------------------------------------------

function SpawnAgentForm() {
  const agents = useAgentStore((s) => s.agents);
  const agentList = Array.from(agents.values());

  const [name, setName] = useState("");
  const [role, setRole] = useState("");
  const [caps, setCaps] = useState("");
  const [status, setStatus] = useState<string | null>(null);

  const spawn = async () => {
    if (!name) return;
    setStatus("spawning...");
    try {
      const res = await apiClient.spawnAgent({
        name,
        role: role || `${name} agent`,
        capabilities: caps ? caps.split(",").map((s) => s.trim()) : [],
      });
      setStatus(`✓ Spawned (${res.id})`);
      setName("");
      setRole("");
      setCaps("");
      setTimeout(() => setStatus(null), 3000);
    } catch (err) {
      setStatus(`✗ ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const shutdown = async (id: string) => {
    try {
      await apiClient.shutdownAgent(id);
    } catch (err) {
      console.error("Shutdown failed:", err instanceof Error ? err.message : err);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
      {/* Spawn form */}
      <div style={{ display: "flex", gap: 8, alignItems: "flex-end" }}>
        <Field label="Name" width={120}>
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="agent name" style={inputStyle} />
        </Field>
        <Field label="Role" width={0} flex={1}>
          <input value={role} onChange={(e) => setRole(e.target.value)} placeholder="role description" style={inputStyle} />
        </Field>
        <Field label="Capabilities" width={160}>
          <input value={caps} onChange={(e) => setCaps(e.target.value)} placeholder="cap1, cap2" style={inputStyle} />
        </Field>
        <button onClick={spawn} disabled={!name} style={btnPrimary(!name)}>
          Spawn
        </button>
        {status && (
          <span style={{ fontSize: 10, color: status.startsWith("✓") ? "#3FB950" : "#F85149", whiteSpace: "nowrap" }}>
            {status}
          </span>
        )}
      </div>

      {/* Quick shutdown buttons */}
      {agentList.filter((a) => a.status !== "shutdown").length > 0 && (
        <div style={{ display: "flex", gap: 4, flexWrap: "wrap", alignItems: "center" }}>
          <span style={{ fontSize: 10, color: "#8B949E", textTransform: "uppercase", marginRight: 4 }}>
            Shutdown:
          </span>
          {agentList
            .filter((a) => a.status !== "shutdown")
            .map((a) => (
              <button
                key={a.id}
                onClick={() => shutdown(a.id)}
                style={{
                  padding: "2px 8px",
                  fontSize: 10,
                  borderRadius: 4,
                  border: "1px solid #30363D",
                  background: "#21262D",
                  color: "#F85149",
                  cursor: "pointer",
                  fontFamily: "'JetBrains Mono', monospace",
                }}
              >
                ✕ {a.name}
              </button>
            ))}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Shared UI helpers
// ---------------------------------------------------------------------------

function Field({
  label,
  width,
  flex,
  children,
}: {
  label: string;
  width: number;
  flex?: number;
  children: React.ReactNode;
}) {
  return (
    <div style={{ width: width || undefined, flex: flex || undefined }}>
      <div style={{ fontSize: 9, color: "#8B949E", textTransform: "uppercase", marginBottom: 2 }}>{label}</div>
      {children}
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "5px 8px",
  fontSize: 12,
  fontFamily: "'JetBrains Mono', monospace",
  background: "#0D1117",
  border: "1px solid #30363D",
  borderRadius: 6,
  color: "#E6EDF3",
  outline: "none",
  boxSizing: "border-box",
};

const selectStyle: React.CSSProperties = {
  ...inputStyle,
  appearance: "none",
  paddingRight: 20,
  backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%238B949E' d='M3 5l3 3 3-3'/%3E%3C/svg%3E")`,
  backgroundRepeat: "no-repeat",
  backgroundPosition: "right 6px center",
};

function btnPrimary(disabled: boolean): React.CSSProperties {
  return {
    padding: "5px 16px",
    fontSize: 12,
    fontWeight: 600,
    borderRadius: 6,
    border: "1px solid #238636",
    background: disabled ? "#21262D" : "#238636",
    color: disabled ? "#484F58" : "#FFFFFF",
    cursor: disabled ? "not-allowed" : "pointer",
    whiteSpace: "nowrap",
    flexShrink: 0,
  };
}
