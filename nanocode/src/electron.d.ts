import type { SessionData, SessionMeta } from "./types/session";

export {};

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

      glob?: (pattern: string, options: {
        cwd: string;
        absolute?: boolean;
        nodir?: boolean;
        dot?: boolean;
        follow?: boolean;
      }) => Promise<string[]>;

      stat?: (filePath: string, cwd?: string) => Promise<{
        mtimeMs: number;
        isDirectory: boolean;
      }>;
    };
  }
}
