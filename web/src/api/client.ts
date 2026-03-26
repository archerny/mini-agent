// ---------------------------------------------------------------------------
// API Client — typed wrappers for the REST API write endpoints
// ---------------------------------------------------------------------------

export interface SendMessageParams {
  type?: string;
  from: string;
  to: string;
  content: string;
}

export interface SpawnAgentParams {
  id?: string;
  name: string;
  role: string;
  capabilities: string[];
}

export interface DemoStatus {
  running: boolean;
  paused: boolean;
  name: string;
}

const API_BASE = "/api";

async function post<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  return res.json();
}

async function del<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { method: "DELETE" });
  return res.json();
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`);
  return res.json();
}

export const apiClient = {
  sendMessage: (params: SendMessageParams) =>
    post<{ id: string; status: string }>("/messages", params),

  spawnAgent: (params: SpawnAgentParams) =>
    post<{ id: string; status: string }>("/agents", params),

  shutdownAgent: (id: string) =>
    del<{ id: string; status: string }>(`/agents/${id}`),

  pauseDemo: () => post<{ status: string }>("/demo/pause"),
  resumeDemo: () => post<{ status: string }>("/demo/resume"),
  getDemoStatus: () => get<DemoStatus>("/demo/status"),
};
