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
    list: (projectKey) => transport.list(projectKey),
    load: (projectKey, id) => transport.load(projectKey, id),
    save: (projectKey, session) => transport.save(projectKey, session),
    delete: (projectKey, id) => transport.delete(projectKey, id),
  };
}

export const sessionRepository = createSessionRepository(
  electronSessionTransport
);
