import type { SessionData, SessionMeta } from "../../types/session";
import {
  electronSessionTransport,
  type SessionTransport,
} from "./transport";

export interface SessionRepository {
  list(projectKey: string): Promise<SessionMeta[]>;
  load(projectKey: string, id: string): Promise<SessionData | null>;
  save(projectKey: string, session: SessionData): Promise<void>;
  delete(projectKey: string, id: string): Promise<void>;
}

export function createSessionRepository(
  transport: SessionTransport
): SessionRepository {
  return {
    async list(projectKey) {
      if (!projectKey) {
        throw new Error("projectKey is required");
      }
      const sessions = await transport.list(projectKey);
      return [...sessions].sort((a, b) => b.createdAt - a.createdAt);
    },
    async load(projectKey, id) {
      if (!projectKey) {
        throw new Error("projectKey is required");
      }
      if (!id) {
        throw new Error("session id is required");
      }
      return transport.load(projectKey, id);
    },
    async save(projectKey, session) {
      if (!projectKey) {
        throw new Error("projectKey is required");
      }
      if (!session.id) {
        throw new Error("session id is required");
      }
      return transport.save(projectKey, session);
    },
    async delete(projectKey, id) {
      if (!projectKey) {
        throw new Error("projectKey is required");
      }
      if (!id) {
        throw new Error("session id is required");
      }
      return transport.delete(projectKey, id);
    },
  };
}

export const sessionRepository = createSessionRepository(
  electronSessionTransport
);
