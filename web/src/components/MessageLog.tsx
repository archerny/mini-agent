// ---------------------------------------------------------------------------
// MessageLog — right panel showing message timeline with filtering
// ---------------------------------------------------------------------------

import { useEffect, useRef, useState, useMemo } from "react";
import { useAgentStore } from "@/stores/agentStore.ts";
import type { Message, MessageType } from "@/protocol/types.ts";

const MSG_COLORS: Record<MessageType, string> = {
  "agent.message": "#4FC3F7",
  "agent.request": "#FFB74D",
  "agent.response": "#81C784",
  "agent.broadcast": "#CE93D8",
};

const MSG_LABELS: Record<MessageType, string> = {
  "agent.message": "MSG",
  "agent.request": "REQ",
  "agent.response": "RES",
  "agent.broadcast": "BCT",
};

function formatTime(ts: string): string {
  try {
    const d = new Date(ts);
    return d.toLocaleTimeString("en-US", { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" });
  } catch {
    return "??:??:??";
  }
}

function shortenId(id: string): string {
  if (id.startsWith("agent-")) return id.slice(6);
  return id;
}

export function MessageLog() {
  const messages = useAgentStore((s) => s.messages);
  const agents = useAgentStore((s) => s.agents);
  const containerRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  // Filter state
  const [filterAgent, setFilterAgent] = useState<string>("");
  const [filterType, setFilterType] = useState<string>("");
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const agentList = useMemo(() => Array.from(agents.values()), [agents]);

  // Filtered messages
  const filtered = useMemo(() => {
    return messages.filter((msg) => {
      if (filterAgent && msg.from !== filterAgent && msg.to !== filterAgent) return false;
      if (filterType && msg.type !== filterType) return false;
      return true;
    });
  }, [messages, filterAgent, filterType]);

  // Auto-scroll to bottom when new messages arrive.
  useEffect(() => {
    const el = containerRef.current;
    if (el && shouldAutoScroll.current) {
      el.scrollTop = el.scrollHeight;
    }
  }, [filtered]);

  // Detect if user has scrolled up (disable auto-scroll).
  function onScroll() {
    const el = containerRef.current;
    if (!el) return;
    const isAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
    shouldAutoScroll.current = isAtBottom;
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {/* Header */}
      <div
        style={{
          padding: "8px 12px",
          fontSize: 13,
          fontWeight: 600,
          color: "#8B949E",
          textTransform: "uppercase",
          letterSpacing: 1,
          borderBottom: "1px solid #30363D",
          flexShrink: 0,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <span>
          Messages ({filtered.length}{filtered.length !== messages.length ? ` / ${messages.length}` : ""})
        </span>
      </div>

      {/* Filter bar */}
      <div
        style={{
          display: "flex",
          gap: 6,
          padding: "6px 12px",
          borderBottom: "1px solid #21262D",
          flexShrink: 0,
          alignItems: "center",
        }}
      >
        <span style={{ fontSize: 9, color: "#484F58", textTransform: "uppercase" }}>Filter</span>
        <select
          value={filterAgent}
          onChange={(e) => setFilterAgent(e.target.value)}
          style={filterSelectStyle}
        >
          <option value="">All Agents</option>
          {agentList.map((a) => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value)}
          style={filterSelectStyle}
        >
          <option value="">All Types</option>
          <option value="agent.message">Message</option>
          <option value="agent.request">Request</option>
          <option value="agent.response">Response</option>
          <option value="agent.broadcast">Broadcast</option>
        </select>
        {(filterAgent || filterType) && (
          <button
            onClick={() => { setFilterAgent(""); setFilterType(""); }}
            style={{
              padding: "1px 6px",
              fontSize: 10,
              border: "1px solid #30363D",
              borderRadius: 4,
              background: "#21262D",
              color: "#8B949E",
              cursor: "pointer",
            }}
          >
            Clear
          </button>
        )}
      </div>

      {/* Message list */}
      <div
        ref={containerRef}
        onScroll={onScroll}
        style={{
          flex: 1,
          overflowY: "auto",
          padding: "8px 0",
          fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
          fontSize: 12,
          lineHeight: 1.6,
        }}
      >
        {filtered.length === 0 ? (
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
            {messages.length === 0 ? "No messages yet…" : "No messages match filter"}
          </div>
        ) : (
          filtered.map((msg) => (
            <MessageRow
              key={msg.id}
              msg={msg}
              expanded={expandedId === msg.id}
              onToggle={() => setExpandedId(expandedId === msg.id ? null : msg.id)}
            />
          ))
        )}
      </div>
    </div>
  );
}

function MessageRow({
  msg,
  expanded,
  onToggle,
}: {
  msg: Message;
  expanded: boolean;
  onToggle: () => void;
}) {
  const color = MSG_COLORS[msg.type] ?? "#8B949E";
  const label = MSG_LABELS[msg.type] ?? "???";

  return (
    <div>
      <div
        onClick={onToggle}
        style={{
          padding: "3px 12px",
          display: "flex",
          gap: 8,
          alignItems: "flex-start",
          cursor: "pointer",
          background: expanded ? "#21262D" : "transparent",
        }}
      >
        {/* Timestamp */}
        <span style={{ color: "#484F58", flexShrink: 0 }}>{formatTime(msg.timestamp)}</span>

        {/* Type badge */}
        <span
          style={{
            color,
            fontWeight: 700,
            fontSize: 10,
            padding: "1px 4px",
            borderRadius: 3,
            background: color + "15",
            flexShrink: 0,
            minWidth: 28,
            textAlign: "center",
          }}
        >
          {label}
        </span>

        {/* From → To */}
        <span style={{ color: "#8B949E", flexShrink: 0 }}>
          {shortenId(msg.from)}
          <span style={{ color: "#484F58" }}> → </span>
          {msg.to === "*" ? "ALL" : shortenId(msg.to)}
        </span>

        {/* Content */}
        <span style={{ color: "#E6EDF3", wordBreak: "break-word" }}>
          {expanded ? msg.payload.content : truncate(msg.payload.content, 120)}
        </span>
      </div>

      {/* Expanded detail */}
      {expanded && (
        <div
          style={{
            padding: "4px 12px 8px 54px",
            background: "#21262D",
            fontSize: 10,
            color: "#8B949E",
            display: "flex",
            gap: 16,
          }}
        >
          <span>ID: {msg.id.slice(0, 12)}…</span>
          <span>Type: {msg.type}</span>
          {msg.correlation_id && <span>Corr: {msg.correlation_id.slice(0, 12)}…</span>}
          <span>Content-Type: {msg.payload.content_type}</span>
          <span>{msg.payload.content.length} chars</span>
        </div>
      )}
    </div>
  );
}

function truncate(s: string, max: number): string {
  if (s.length <= max) return s;
  return s.slice(0, max - 1) + "…";
}

const filterSelectStyle: React.CSSProperties = {
  padding: "2px 6px",
  fontSize: 10,
  background: "#0D1117",
  border: "1px solid #30363D",
  borderRadius: 4,
  color: "#E6EDF3",
  fontFamily: "'JetBrains Mono', monospace",
  outline: "none",
};
