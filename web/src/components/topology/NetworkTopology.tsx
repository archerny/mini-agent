// ---------------------------------------------------------------------------
// NetworkTopology — main topology visualization with React Flow
// ---------------------------------------------------------------------------

import { useCallback, useEffect, useMemo, useRef } from "react";
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { useAgentStore } from "@/stores/agentStore.ts";
import { AgentNode, type AgentNodeData } from "./AgentNode.tsx";
import { AnimatedEdge, type AnimatedEdgeData } from "./AnimatedEdge.tsx";

// ---------------------------------------------------------------------------
// Node / Edge type registrations
// ---------------------------------------------------------------------------

const nodeTypes = { agent: AgentNode };
const edgeTypes = { animated: AnimatedEdge };

// ---------------------------------------------------------------------------
// Layout helpers — circular layout
// ---------------------------------------------------------------------------

function circularLayout(count: number, centerX: number, centerY: number, radius: number) {
  const positions: Array<{ x: number; y: number }> = [];
  for (let i = 0; i < count; i++) {
    const angle = (2 * Math.PI * i) / count - Math.PI / 2; // start from top
    positions.push({
      x: centerX + radius * Math.cos(angle),
      y: centerY + radius * Math.sin(angle),
    });
  }
  return positions;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function NetworkTopology() {
  const agents = useAgentStore((s) => s.agents);
  const topology = useAgentStore((s) => s.topology);
  const messages = useAgentStore((s) => s.messages);
  const selectedAgentId = useAgentStore((s) => s.selectedAgentId);
  const selectAgent = useAgentStore((s) => s.selectAgent);

  const [nodes, setNodes, onNodesChangeBase] = useNodesState<Node<AgentNodeData>>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge<AnimatedEdgeData>>([]);

  // Track recently active edges (last 3 seconds)
  const recentMessages = useRef<Map<string, number>>(new Map());

  // Track which edges had messages recently
  useEffect(() => {
    if (messages.length === 0) return;
    const latest = messages[messages.length - 1];
    if (!latest) return;

    const edgeKey = `${latest.from}-${latest.to}`;
    recentMessages.current.set(edgeKey, Date.now());

    // Clean up old entries
    const now = Date.now();
    for (const [key, ts] of recentMessages.current.entries()) {
      if (now - ts > 3000) {
        recentMessages.current.delete(key);
      }
    }
  }, [messages]);

  // -------------------------------------------------------------------------
  // Convert store data → React Flow nodes & edges
  // -------------------------------------------------------------------------
  const agentList = useMemo(() => Array.from(agents.values()), [agents]);

  // Track which nodes have been positioned (to avoid resetting dragged positions).
  const nodePositions = useRef<Map<string, { x: number; y: number }>>(new Map());
  const prevAgentCount = useRef(0);

  const updateGraph = useCallback(() => {
    if (agentList.length === 0) {
      setNodes([]);
      setEdges([]);
      nodePositions.current.clear();
      prevAgentCount.current = 0;
      return;
    }

    // Recalculate layout only when agent count changes (new agent added/removed).
    const needsLayout = agentList.length !== prevAgentCount.current;
    prevAgentCount.current = agentList.length;

    if (needsLayout) {
      // Compute circular layout for all agents, but only apply to NEW nodes.
      const positions = circularLayout(agentList.length, 300, 250, Math.max(150, agentList.length * 50));
      agentList.forEach((agent, i) => {
        if (!nodePositions.current.has(agent.id)) {
          nodePositions.current.set(agent.id, positions[i] ?? { x: 0, y: 0 });
        }
      });
      // Remove positions for agents that no longer exist.
      const currentIds = new Set(agentList.map((a) => a.id));
      for (const id of nodePositions.current.keys()) {
        if (!currentIds.has(id)) nodePositions.current.delete(id);
      }
    }

    const newNodes: Node<AgentNodeData>[] = agentList.map((agent) => ({
      id: agent.id,
      type: "agent",
      position: nodePositions.current.get(agent.id) ?? { x: 0, y: 0 },
      data: {
        name: agent.name,
        role: agent.role,
        status: agent.status,
        capabilities: agent.capabilities ?? [],
      },
      selected: agent.id === selectedAgentId,
    }));

    // Build edge map from topology + messages
    const edgeMap = new Map<string, { from: string; to: string; messageCount: number }>();

    // From topology API
    for (const e of topology.edges) {
      const key = `${e.from}-${e.to}`;
      edgeMap.set(key, { from: e.from, to: e.to, messageCount: e.message_count });
    }

    // Also infer from messages if topology is empty
    if (topology.edges.length === 0) {
      for (const msg of messages) {
        if (msg.to === "*") continue; // skip broadcasts
        const key = `${msg.from}-${msg.to}`;
        const existing = edgeMap.get(key);
        if (existing) {
          existing.messageCount++;
        } else {
          edgeMap.set(key, { from: msg.from, to: msg.to, messageCount: 1 });
        }
      }
    }

    const now = Date.now();
    const newEdges: Edge<AnimatedEdgeData>[] = Array.from(edgeMap.values()).map((e) => {
      const key = `${e.from}-${e.to}`;
      const lastActive = recentMessages.current.get(key) ?? 0;
      const active = now - lastActive < 3000;

      return {
        id: `e-${e.from}-${e.to}`,
        source: e.from,
        target: e.to,
        type: "animated",
        data: {
          messageCount: e.messageCount,
          active,
        },
      };
    });

    setNodes(newNodes);
    setEdges(newEdges);
  }, [agentList, topology, messages, selectedAgentId, setNodes, setEdges]);

  useEffect(() => {
    updateGraph();
  }, [updateGraph]);

  // Refresh active state periodically
  useEffect(() => {
    const timer = setInterval(() => {
      updateGraph();
    }, 1000);
    return () => clearInterval(timer);
  }, [updateGraph]);

  // Wrap onNodesChange to persist dragged positions.
  const onNodesChange = useCallback(
    (changes: Parameters<typeof onNodesChangeBase>[0]) => {
      onNodesChangeBase(changes);
      for (const change of changes) {
        if (change.type === "position" && change.position && change.dragging) {
          nodePositions.current.set(change.id, { x: change.position.x, y: change.position.y });
        }
      }
    },
    [onNodesChangeBase],
  );

  // Handle node click → select agent
  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      selectAgent(selectedAgentId === node.id ? null : node.id);
    },
    [selectAgent, selectedAgentId],
  );

  const onPaneClick = useCallback(() => {
    selectAgent(null);
  }, [selectAgent]);

  if (agentList.length === 0) {
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          height: "100%",
          gap: 16,
          color: "#484F58",
        }}
      >
        <div style={{ fontSize: 48, opacity: 0.3 }}>🕸️</div>
        <div style={{ fontSize: 13, fontStyle: "italic" }}>
          Waiting for agents to join the network…
        </div>
      </div>
    );
  }

  return (
    <div style={{ width: "100%", height: "100%", position: "relative" }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        minZoom={0.3}
        maxZoom={2}
        proOptions={{ hideAttribution: true }}
        style={{ background: "transparent" }}
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={24}
          size={1}
          color="#21262D"
        />

        {/* SVG defs for glow effects */}
        <svg style={{ position: "absolute", width: 0, height: 0 }}>
          <defs>
            <filter id="particleGlow" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="2" />
            </filter>
          </defs>
        </svg>
      </ReactFlow>

      {/* CSS animations */}
      <style>{`
        @keyframes agentPulse {
          0% { opacity: 0.7; transform: translateX(-50%) scale(1); }
          50% { opacity: 0; transform: translateX(-50%) scale(1.4); }
          100% { opacity: 0; transform: translateX(-50%) scale(1.4); }
        }
        .react-flow__node {
          cursor: pointer !important;
        }
        .react-flow__node.selected > div {
          filter: brightness(1.2);
        }
        .react-flow__controls {
          background: #161B22 !important;
          border: 1px solid #30363D !important;
          border-radius: 8px !important;
        }
        .react-flow__controls button {
          background: #161B22 !important;
          border-color: #30363D !important;
          color: #E6EDF3 !important;
        }
        .react-flow__controls button:hover {
          background: #21262D !important;
        }
        .react-flow__minimap {
          background: #0D1117 !important;
          border: 1px solid #30363D !important;
          border-radius: 8px !important;
        }
      `}</style>
    </div>
  );
}
