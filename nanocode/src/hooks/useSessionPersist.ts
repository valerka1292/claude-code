import { useCallback, useRef } from "react";
import {
  generateSessionName,
  sessionRepository,
  type StoredMessage,
} from "../lib/sessions/index";
import type { Provider } from "../lib/providers";
import { useSession } from "../contexts/SessionContext";

interface StartSessionNameGenerationOptions {
  isFirstMessage: boolean;
  sessionId: string;
  sessionName: string;
  provider: Provider | null;
  firstUserMessage: string;
}

interface PersistTurnOptions {
  projectKey: string | null;
  sessionId: string;
  userInput: string;
  sendTs: number;
  assistantContent: string;
  assistantReasoning: string;
  turnMessages: StoredMessage[];
  isAborted?: boolean;
}

export function useSessionPersist() {
  const {
    getActiveSessionSnapshot,
    updateSession,
    onSessionSaveError,
  } = useSession();
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
    [getActiveSessionSnapshot, updateSession]
  );

  const persistCompletedTurn = useCallback(
    async ({
      projectKey,
      sessionId,
      userInput,
      sendTs,
      assistantContent,
      assistantReasoning,
      turnMessages,
      isAborted = false,
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
        id: crypto.randomUUID(),
        role: "user",
        content: userInput,
        ts: sendTs,
      };

      const latestSession = getActiveSessionSnapshot();
      if (!latestSession || latestSession.id !== sessionId) return;

      let messagesToAdd: StoredMessage[];
      if (turnMessages.length > 0) {
        messagesToAdd = [userStored, ...turnMessages];

        const hasFinalAssistantInTurn = turnMessages.some(
          (msg) =>
            msg.role === "assistant" &&
            !msg.tool_calls &&
            (msg.content ?? "") === assistantContent
        );

        if (!isAborted && assistantContent && !hasFinalAssistantInTurn) {
          messagesToAdd.push({
            id: crypto.randomUUID(),
            role: "assistant",
            content: assistantContent,
            reasoning: assistantReasoning || undefined,
            ts: Date.now(),
          });
        }
      } else if (!isAborted) {
        const assistantStored: StoredMessage = {
          id: crypto.randomUUID(),
          role: "assistant",
          content: assistantContent || null,
          reasoning: assistantReasoning || undefined,
          ts: Date.now(),
        };
        messagesToAdd = [userStored, assistantStored];
      } else {
        messagesToAdd = [userStored];
      }

      const finalSession = {
        ...latestSession,
        name: sessionNameByIdRef.current[sessionId] ?? latestSession.name,
        messages: [...latestSession.messages, ...messagesToAdd],
      };

      try {
        await sessionRepository.save(projectKey, finalSession);
        const latestAfterSave = getActiveSessionSnapshot();
        if (latestAfterSave && latestAfterSave.id === sessionId) {
          updateSession(finalSession);
        }
      } catch (error) {
        onSessionSaveError(error);
      }
    },
    [getActiveSessionSnapshot, onSessionSaveError, updateSession]
  );

  return {
    startSessionNameGeneration,
    persistCompletedTurn,
  };
}
