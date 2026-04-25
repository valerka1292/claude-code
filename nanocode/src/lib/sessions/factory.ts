import type { SessionData } from "../../types/session";

export const DEFAULT_SESSION_NAME = "New session";

function generateId(): string {
  return crypto.randomUUID();
}

export function createNewSession(projectPath: string): SessionData {
  return {
    id: generateId(),
    projectPath,
    name: DEFAULT_SESSION_NAME,
    createdAt: Date.now(),
    messages: [],
  };
}
