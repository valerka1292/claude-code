export {};

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

declare global {
  interface Window {
    electronAPI?: {
      minimize: () => void;
      maximize: () => void;
      close: () => void;
      platform: string;

      selectFolder: () => Promise<string | null>;
      resolvePath: (p: string) => Promise<string | null>;

      loadSettings: () => Promise<{ providers: Record<string, unknown>; activeProviderId: string | null }>;
      saveSettings: (data: unknown) => Promise<boolean>;

      listSessions: (projectKey: string) => Promise<SessionMeta[]>;
      loadSession: (projectKey: string, id: string) => Promise<SessionData | null>;
      saveSession: (projectKey: string, session: SessionData) => Promise<boolean>;
      deleteSession: (projectKey: string, id: string) => Promise<boolean>;
    };
  }
}