import type { SessionData } from "../types/session";

interface StartNewSessionOptions {
  isTurnActive: boolean;
  currentSession: SessionData | null;
  projectKey: string | null;
  clearActiveSession: () => void;
  saveSession: (projectKey: string, session: SessionData) => Promise<void>;
  onSessionSaveError: (error: unknown) => void;
}

export async function startNewSessionWithGuard({
  isTurnActive,
  currentSession,
  projectKey,
  clearActiveSession,
  saveSession,
  onSessionSaveError,
}: StartNewSessionOptions): Promise<boolean> {
  if (isTurnActive) {
    return false;
  }

  clearActiveSession();

  if (currentSession && projectKey && currentSession.messages.length > 0) {
    try {
      await saveSession(projectKey, currentSession);
    } catch (error) {
      onSessionSaveError(error);
    }
  }

  return true;
}
