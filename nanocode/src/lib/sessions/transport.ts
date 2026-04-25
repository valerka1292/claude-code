import type { SessionData, SessionMeta } from "../../types/session";

export interface SessionTransport {
  list(projectKey: string): Promise<SessionMeta[]>;
  load(projectKey: string, id: string): Promise<SessionData | null>;
  save(projectKey: string, session: SessionData): Promise<void>;
  delete(projectKey: string, id: string): Promise<void>;
}

export const electronSessionTransport: SessionTransport = {
  async list(projectKey) {
    return (await window.electronAPI?.listSessions(projectKey)) ?? [];
  },
  async load(projectKey, id) {
    return (await window.electronAPI?.loadSession(projectKey, id)) ?? null;
  },
  async save(projectKey, session) {
    await window.electronAPI?.saveSession(projectKey, session);
  },
  async delete(projectKey, id) {
    await window.electronAPI?.deleteSession(projectKey, id);
  },
};
