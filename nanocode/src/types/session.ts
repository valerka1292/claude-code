export interface SessionMeta {
  id: string;
  name: string;
  createdAt: number;
  projectPath: string;
}

export interface SessionData extends SessionMeta {
  messages: StoredMessage[];
}

export interface StoredMessage {
  role: "user" | "assistant" | "tool";
  content: string | null;
  tool_calls?: unknown[];
  tool_call_id?: string;
  name?: string;
  reasoning?: string;
  ts: number;
}
