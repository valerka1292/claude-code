import { useRef, useCallback, useState } from "react";
import type { Mode } from "../types";
import { useProviders } from "../contexts/ProvidersContext";
import { useProject } from "../contexts/ProjectContext";
import { useSession } from "../contexts/SessionContext";
import { runAgentStream } from "../lib/agentLoop";
import { useAbortController } from "./useAbortController";
import { useMessageStream } from "./useMessageStream";
import { buildChatHistory } from "../lib/chatHistory";
import { useSessionPersist } from "./useSessionPersist";

export function useAgent() {
  const { activeProvider } = useProviders();
  const { folderPath, folderName, projectKey } = useProject();
  const {
    activeSession,
    initSession,
    getActiveSessionSnapshot,
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
  const { startSessionNameGeneration, persistCompletedTurn } =
    useSessionPersist();

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

      startSessionNameGeneration({
        isFirstMessage,
        sessionId: session.id,
        sessionName: session.name,
        provider: ap,
        firstUserMessage: value,
      });

      const history = buildChatHistory({
        cwd: fp,
        projectName: folderNameRef.current ?? fp,
        mode: modeRef.current,
        sessionMessages: session.messages,
        userInput: value,
      });

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

            await persistCompletedTurn({
              projectKey: projectKeyRef.current,
              sessionId: session.id,
              userInput: value,
              sendTs,
              assistantContent,
              assistantReasoning,
            });
          },
        },
        fp,
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
      persistCompletedTurn,
      replaceActiveController,
      setUsedTokens,
      setIsTyping,
      startSessionNameGeneration,
      updateMsg,
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
