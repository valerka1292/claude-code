import { useCallback, useRef } from "react";
import {
  generateSessionName,
  sessionRepository,
  type StoredMessage,
} from "../lib/sessions";
import type { Provider } from "../lib/providers";
import type { SessionData } from "../types/session";

interface StartSessionNameGenerationOptions {
  isFirstMessage: boolean;
  sessionId: string;
  sessionName: string;
  provider: Provider | null;
  firstUserMessage: string;
  getActiveSessionSnapshot: () => SessionData | null;
  updateSession: (session: SessionData) => void;
}

interface PersistTurnOptions {
  projectKey: string | null;
  sessionId: string;
  userInput: string;
  sendTs: number;
  assistantContent: string;
  assistantReasoning: string;
  getActiveSessionSnapshot: () => SessionData | null;
  updateSession: (session: SessionData) => void;
  onSessionSaveError: (error: unknown) => void;
}

export function useSessionPersist() {
  const sessionNameByIdRef = useRef<Record<string, string>>({});
  const pendingNameGenerationByIdRef = useRef<Record<string, Promise<string>>>(
    {}
  );

  const startSessionNameGeneration = useCallback(
    ({
      isFirstMessage,
      sessionId,
      sessionName,
      provider,
      firstUserMessage,
      getActiveSessionSnapshot,
      updateSession,
    }: StartSessionNameGenerationOptions) => {
      sessionNameByIdRef.current[sessionId] = sessionName;

      if (!isFirstMessage || !provider) {
        return;
      }

      const pendingGeneration = generateSessionName(provider, firstUserMessage)
        .then((name) => {
          sessionNameByIdRef.current[sessionId] = name;
          const current = getActiveSessionSnapshot();
          if (current && current.id === sessionId) {
            updateSession({ ...current, name });
          }
          return name;
        })
        .catch((error) => {
          console.error("[agent] generateSessionName error:", error);
          return sessionNameByIdRef.current[sessionId] ?? sessionName;
        })
        .finally(() => {
          delete pendingNameGenerationByIdRef.current[sessionId];
        });

      pendingNameGenerationByIdRef.current[sessionId] = pendingGeneration;
    },
    []
  );

  const persistCompletedTurn = useCallback(
    async ({
      projectKey,
      sessionId,
      userInput,
      sendTs,
      assistantContent,
      assistantReasoning,
      getActiveSessionSnapshot,
      updateSession,
      onSessionSaveError,
    }: PersistTurnOptions) => {
      if (!projectKey) {
        return;
      }

      const pendingSessionName =
        pendingNameGenerationByIdRef.current[sessionId];
      if (pendingSessionName) {
        await pendingSessionName;
      }

      const userStored: StoredMessage = {
        role: "user",
        content: userInput,
        ts: sendTs,
      };

      const assistantStored: StoredMessage = {
        role: "assistant",
        content: assistantContent || null,
        reasoning: assistantReasoning || undefined,
        ts: Date.now(),
      };

      const latestSession = getActiveSessionSnapshot();
      if (!latestSession || latestSession.id !== sessionId) return;

      const finalSession = {
        ...latestSession,
        name: sessionNameByIdRef.current[sessionId] ?? latestSession.name,
        messages: [...latestSession.messages, userStored, assistantStored],
      };

      try {
        await sessionRepository.save(projectKey, finalSession);
        updateSession(finalSession);
      } catch (error) {
        onSessionSaveError(error);
      }
    },
    []
  );

  return {
    startSessionNameGeneration,
    persistCompletedTurn,
  };
}
