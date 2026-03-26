// ---------------------------------------------------------------------------
// AgentNode — custom React Flow node for an agent in the topology
// ---------------------------------------------------------------------------

import { memo } from "react";
import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import type { AgentState } from "@/protocol/types.ts";
import { STATUS_COLORS } from "@/protocol/constants.ts";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface AgentNodeData {
  name: string;
  role: string;
  status: AgentState;
  capabilities: string[];
  [key: string]: unknown; // React Flow requires this
}

export type AgentNodeType = Node<AgentNodeData, "agent">;

// ---------------------------------------------------------------------------
// Status glow map (only busy/error have glow)
// ---------------------------------------------------------------------------

const STATUS_GLOW: Record<string, string> = {
  busy: "0 0 16px rgba(63, 185, 80, 0.5)",
  error: "0 0 16px rgba(248, 81, 73, 0.5)",
};

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export const AgentNode = memo(function AgentNode({ data }: NodeProps<AgentNodeType>) {
  const color = STATUS_COLORS[data.status] ?? "#8B949E";
  const glow = STATUS_GLOW[data.status] ?? "none";
  const isBusy = data.status === "busy";
  const initial = data.name.charAt(0).toUpperCase();

  return (
    <>
      {/* Invisible target/source handles for edges */}
      <Handle
        type="target"
        position={Position.Left}
        style={{ opacity: 0, width: 1, height: 1 }}
      />
      <Handle
        type="source"
        position={Position.Right}
        style={{ opacity: 0, width: 1, height: 1 }}
      />

      <div style={{ position: "relative", display: "flex", flexDirection: "column", alignItems: "center" }}>
        {/* Pulse ring for busy state */}
        {isBusy && (
          <div
            style={{
              position: "absolute",
              top: -4,
              left: "50%",
              transform: "translateX(-50%)",
              width: 64,
              height: 64,
              borderRadius: "50%",
              border: `2px solid ${color}`,
              animation: "agentPulse 1.5s ease-in-out infinite",
              pointerEvents: "none",
            }}
          />
        )}

        {/* Main circle */}
        <div
          style={{
            width: 56,
            height: 56,
            borderRadius: "50%",
            border: `3px solid ${color}`,
            background: `${color}18`,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: 22,
            fontWeight: 700,
            color: "#E6EDF3",
            fontFamily: "'JetBrains Mono', monospace",
            boxShadow: glow,
            transition: "all 0.3s ease",
            cursor: "pointer",
          }}
        >
          {initial}
        </div>

        {/* Name label */}
        <div
          style={{
            marginTop: 8,
            fontSize: 11,
            fontWeight: 600,
            color: "#E6EDF3",
            fontFamily: "'JetBrains Mono', monospace",
            textAlign: "center",
            maxWidth: 100,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {data.name}
        </div>

        {/* Status label */}
        <div
          style={{
            marginTop: 2,
            fontSize: 9,
            fontWeight: 600,
            color,
            textTransform: "uppercase",
            letterSpacing: 0.5,
          }}
        >
          {data.status}
        </div>
      </div>
    </>
  );
});
