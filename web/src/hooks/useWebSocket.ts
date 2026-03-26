// ---------------------------------------------------------------------------
// useWebSocket — manages the WebSocket lifecycle
// ---------------------------------------------------------------------------

import { useEffect, useRef } from "react";
import { useAgentStore } from "@/stores/agentStore.ts";
import type { Event } from "@/protocol/types.ts";

const WS_URL = `${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}/ws/events`;

const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempt = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleEvent = useAgentStore((s) => s.handleEvent);
  const loadSnapshot = useAgentStore((s) => s.loadSnapshot);
  const setConnectionStatus = useAgentStore((s) => s.setConnectionStatus);

  useEffect(() => {
    let disposed = false;

    function connect() {
      if (disposed) return;

      setConnectionStatus("connecting");

      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = async () => {
        if (disposed) return;
        reconnectAttempt.current = 0;
        setConnectionStatus("connected");

        // Load full snapshot on (re)connect.
        await loadSnapshot();
      };

      ws.onmessage = (ev) => {
        if (disposed) return;
        try {
          const event: Event = JSON.parse(ev.data);
          handleEvent(event);
        } catch (err) {
          console.error("[ws] Failed to parse event:", err);
        }
      };

      ws.onclose = () => {
        if (disposed) return;
        setConnectionStatus("disconnected");
        scheduleReconnect();
      };

      ws.onerror = () => {
        if (disposed) return;
        setConnectionStatus("error");
        // onclose will fire after onerror, so reconnect is handled there.
      };
    }

    function scheduleReconnect() {
      if (disposed) return;
      const delay = Math.min(
        RECONNECT_BASE_MS * Math.pow(2, reconnectAttempt.current),
        RECONNECT_MAX_MS,
      );
      reconnectAttempt.current++;
      reconnectTimer.current = setTimeout(connect, delay);
    }

    connect();

    return () => {
      disposed = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [handleEvent, loadSnapshot, setConnectionStatus]);
}
