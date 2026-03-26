// ---------------------------------------------------------------------------
// MessageLog — right panel showing message timeline as text stream
// ---------------------------------------------------------------------------

import { useEffect, useRef } from "react";
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
  // "agent-researcher" → "researcher"
  if (id.startsWith("agent-")) return id.slice(6);
  return id;
}

function truncate(s: string, max: number): string {
  if (s.length <= max) return s;
  return s.slice(0, max - 1) + "…";
}

export function MessageLog() {
  const messages = useAgentStore((s) => s.messages);
  const containerRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  // Auto-scroll to bottom when new messages arrive.
  useEffect(() => {
    const el = containerRef.current;
    if (el && shouldAutoScroll.current) {
      el.scrollTop = el.scrollHeight;
    }
  }, [messages]);

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
        }}
      >
        Messages ({messages.length})
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
        {messages.length === 0 ? (
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
            No messages yet…
          </div>
        ) : (
          messages.map((msg) => <MessageRow key={msg.id} msg={msg} />)
        )}
      </div>
    </div>
  );
}

function MessageRow({ msg }: { msg: Message }) {
  const color = MSG_COLORS[msg.type] ?? "#8B949E";
  const label = MSG_LABELS[msg.type] ?? "???";

  return (
    <div
      style={{
        padding: "3px 12px",
        display: "flex",
        gap: 8,
        alignItems: "flex-start",
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
        {truncate(msg.payload.content, 120)}
      </span>
    </div>
  );
}
