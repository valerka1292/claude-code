import { useRef, useCallback, useState } from "react";
import type { Mode } from "../types";
import { useProviders } from "../contexts/ProvidersContext";
import { useProject } from "../contexts/ProjectContext";
import { useSession } from "../contexts/SessionContext";
import { runAgentStream, type ChatMessage } from "../lib/agentLoop";
import { buildSystemMessages } from "../lib/systemPrompt";
import { generateSessionName, saveSession, type StoredMessage } from "../lib/sessions";
import { storedToChat } from "../lib/converters";
import { useAbortController } from "./useAbortController";
import { useMessageStream } from "./useMessageStream";

export function useAgent() {
  const { activeProvider } = useProviders();
  const { folderPath, folderName, projectKey } = useProject();
  const {
    activeSession,
    initSession,
    updateSession,
    getActiveSessionSnapshot,
    onSessionSaveError,
  } = useSession();

  const [mode, setMode] = useState<Mode>("Ask");
  const {
    messages,
    isTyping,
    setIsTyping,
    usedTokens,
    setUsedTokens,
    addPendingTurn,
    updateMsg,
    appendReasoningChunk,
    appendContentChunk,
    appendToolCallLabel,
    resetMessageStream,
  } = useMessageStream(activeSession);
  const { replaceActiveController, abortActiveRequest, resetAbortController } =
    useAbortController();

  const projectKeyRef = useRef<string | null>(null);
  projectKeyRef.current = projectKey;

  const activeProviderRef = useRef(activeProvider);
  activeProviderRef.current = activeProvider;

  const folderPathRef = useRef(folderPath);
  folderPathRef.current = folderPath;

  const folderNameRef = useRef(folderName);
  folderNameRef.current = folderName;

  const modeRef = useRef(mode);
  modeRef.current = mode;

  const handleSend = useCallback(
    async (value: string) => {
      const fp = folderPathRef.current;
      const pk = projectKeyRef.current;
      const ap = activeProviderRef.current;

      if (!value.trim() || !fp || !pk) return;

      const controller = replaceActiveController();
      const sendTs = Date.now();
      const isFirstMessage = !getActiveSessionSnapshot();
      const session = getActiveSessionSnapshot() ?? initSession(fp);
      const assistantId = addPendingTurn(value, sendTs);
      setIsTyping(true);

      if (!ap) {
        setTimeout(() => {
          updateMsg(assistantId, {
            content: "⚠ No provider configured. Please add one in Settings.",
            isStreaming: false,
          });
          setIsTyping(false);
        }, 300);
        return;
      }

      let sessionName = session.name;
      if (isFirstMessage) {
        generateSessionName(ap, value)
          .then((name) => {
            sessionName = name;
            const current = getActiveSessionSnapshot();
            if (!current || current.id !== session.id) return;
            updateSession({ ...current, name });
          })
          .catch((error) => {
            console.error("[agent] generateSessionName error:", error);
          });
      }

      const history: ChatMessage[] = [
        ...buildSystemMessages({
          cwd: fp,
          projectName: folderNameRef.current ?? fp,
          mode: modeRef.current,
        }),
        ...session.messages.map(storedToChat),
        { role: "user", content: value },
      ];

      let assistantContent = "";
      let assistantReasoning = "";

      runAgentStream(
        ap,
        history,
        {
          onReasoningChunk: (chunk) => {
            assistantReasoning += chunk;
            appendReasoningChunk(assistantId, chunk);
          },
          onContentChunk: (chunk) => {
            assistantContent += chunk;
            appendContentChunk(assistantId, chunk);
          },
          onToolCallStart: (_id, name) => {
            appendToolCallLabel(assistantId, name);
          },
          onToolCallDone: () => {},
          onUsage: (prompt, completion) => setUsedTokens(prompt + completion),
          onError: (err) => {
            updateMsg(assistantId, {
              content: `⚠ Error: ${err.message}`,
              isStreaming: false,
              isReasoningStreaming: false,
            });
            setIsTyping(false);
          },
          onDone: async () => {
            updateMsg(assistantId, {
              isStreaming: false,
              isReasoningStreaming: false,
            });
            setIsTyping(false);

            const currentPk = projectKeyRef.current;
            if (!currentPk) return;

            const userStored: StoredMessage = {
              role: "user",
              content: value,
              ts: sendTs,
            };

            const assistantStored: StoredMessage = {
              role: "assistant",
              content: assistantContent || null,
              reasoning: assistantReasoning || undefined,
              ts: Date.now(),
            };

            const latestSession = getActiveSessionSnapshot();
            if (!latestSession || latestSession.id !== session.id) return;

            const finalSession = {
              ...latestSession,
              name: sessionName,
              messages: [...latestSession.messages, userStored, assistantStored],
            };

            try {
              await saveSession(currentPk, finalSession);
              updateSession(finalSession);
            } catch (error) {
              onSessionSaveError(error);
            }
          },
        },
        controller.signal
      );
    },
    [
      addPendingTurn,
      appendContentChunk,
      appendReasoningChunk,
      appendToolCallLabel,
      getActiveSessionSnapshot,
      initSession,
      onSessionSaveError,
      replaceActiveController,
      setUsedTokens,
      setIsTyping,
      updateMsg,
      updateSession,
    ]
  );

  const resetAgentUi = useCallback(() => {
    resetAbortController();
    resetMessageStream();
  }, [resetAbortController, resetMessageStream]);

  return {
    mode,
    setMode,
    messages,
    isTyping,
    usedTokens,
    handleSend,
    resetAgentUi,
    abortActiveRequest,
  };
}
