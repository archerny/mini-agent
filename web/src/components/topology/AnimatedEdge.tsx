// ---------------------------------------------------------------------------
// AnimatedEdge — custom React Flow edge with flowing particle animation
// ---------------------------------------------------------------------------

import { memo } from "react";
import {
  BaseEdge,
  getBezierPath,
  type EdgeProps,
  type Edge,
} from "@xyflow/react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface AnimatedEdgeData {
  messageCount: number;
  active: boolean; // true when a message was recently sent on this edge
  [key: string]: unknown;
}

export type AnimatedEdgeType = Edge<AnimatedEdgeData, "animated">;

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export const AnimatedEdge = memo(function AnimatedEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
}: EdgeProps<AnimatedEdgeType>) {
  const messageCount = data?.messageCount ?? 0;
  const active = data?.active ?? false;

  // Thicker edge for more messages, but clamp at 4px
  const strokeWidth = Math.min(1 + Math.log2(Math.max(messageCount, 1)) * 0.5, 4);

  // Color: bright when active, muted otherwise
  const edgeColor = active ? "#58A6FF" : "#30363D";
  const particleColor = active ? "#58A6FF" : "#484F58";

  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
  });

  return (
    <>
      {/* Base edge line */}
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: edgeColor,
          strokeWidth,
          transition: "stroke 0.3s ease, stroke-width 0.3s ease",
        }}
      />

      {/* Particle animation along the path */}
      {messageCount > 0 && (
        <>
          <circle r={3} fill={particleColor} filter="url(#particleGlow)">
            <animateMotion
              dur={active ? "1.2s" : "3s"}
              repeatCount="indefinite"
              path={edgePath}
            />
          </circle>

          {/* Second particle offset for active edges */}
          {active && (
            <circle r={2.5} fill={particleColor} opacity={0.6} filter="url(#particleGlow)">
              <animateMotion
                dur="1.2s"
                repeatCount="indefinite"
                path={edgePath}
                begin="0.6s"
              />
            </circle>
          )}
        </>
      )}

      {/* Message count label at midpoint */}
      {messageCount > 0 && (
        <text
          x={(sourceX + targetX) / 2}
          y={(sourceY + targetY) / 2 - 10}
          textAnchor="middle"
          style={{
            fontSize: 9,
            fill: "#8B949E",
            fontFamily: "'JetBrains Mono', monospace",
            fontWeight: 600,
          }}
        >
          {messageCount}
        </text>
      )}
    </>
  );
});
