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
import type { StoredMessage } from "../types/session";
import { applyToolExecutionResultToMessages } from "./toolExecutionUpdate";

export function useAgent() {
  const { activeProvider } = useProviders();
  const { folderPath, folderName, projectKey } = useProject();
  const {
    activeSession,
    initSession,
    getActiveSessionSnapshot,
    setTurnActive,
  } = useSession();

  const [mode, setMode] = useState<Mode>("Ask");
  const {
    messages,
    updateLiveTurn,
    isTyping,
    setIsTyping,
    usedTokens,
    setUsedTokens,
    addPendingTurn,
    updateMsg,
    appendReasoningChunk,
    appendContentChunk,
    appendBlock,
    finalizeAndCommitLiveTurn,
    resetMessageStream,
    suspendSessionRestore,
    resumeSessionRestore,
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
  const isProcessingRef = useRef(false);

  const handleSend = useCallback(
    async (value: string) => {
      const fp = folderPathRef.current;
      const pk = projectKeyRef.current;
      const ap = activeProviderRef.current;
      const normalizedValue = value.trim();

      if (!normalizedValue || !fp || !pk || isProcessingRef.current) return;
      isProcessingRef.current = true;
      setTurnActive(true);

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
            isProcessingRef.current = false;
            setTurnActive(false);
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
      const turnMessages: StoredMessage[] = [];

      try {
        await runAgentStream(
          ap,
          history,
          {
          onReasoningChunk: (chunk) => {
            assistantReasoning += chunk;
            appendReasoningChunk(assistantId, chunk);
            appendBlock(assistantId, { type: "reasoning", content: chunk, streaming: true });
          },
          onContentChunk: (chunk) => {
            assistantContent += chunk;
            appendContentChunk(assistantId, chunk);
            appendBlock(assistantId, { type: "text", content: chunk, streaming: true });
          },
          onToolCallStart: (id, name) => {
            const toolCall = {
              id,
              name,
              arguments: {},
              status: "pending" as const,
            };
            updateLiveTurn((prev) =>
              prev.map((m) => {
                if (m.id !== assistantId) return m;
                return {
                  ...m,
                  toolCalls: [...(m.toolCalls ?? []), toolCall],
                  blocks: [...(m.blocks ?? []), { type: "tool_call", call: toolCall }],
                };
              })
            );
          },
          onToolCallDone: (id, args) => {
            updateLiveTurn((prev) =>
              prev.map((m) => {
                if (m.id !== assistantId) return m;

                let parsedArgs: Record<string, unknown> = {};
                try {
                  parsedArgs = JSON.parse(args);
                } catch {
                  parsedArgs = { raw: args };
                }

                const updatedToolCalls = m.toolCalls?.map((tc) =>
                  tc.id === id ? { ...tc, arguments: parsedArgs } : tc
                );

                const updatedBlocks = m.blocks?.map((b) => {
                  if (b.type === "tool_call" && b.call.id === id) {
                    return { ...b, call: { ...b.call, arguments: parsedArgs } };
                  }
                  return b;
                });

                return { ...m, toolCalls: updatedToolCalls, blocks: updatedBlocks };
              })
            );
          },
          onAssistantMessageWithTools: (content, toolCalls) => {
            turnMessages.push({
              id: crypto.randomUUID(),
              role: "assistant",
              content,
              tool_calls: toolCalls,
              reasoning: assistantReasoning || undefined,
              ts: Date.now(),
            });
          },
          onToolExecutionStart: (id) => {
            updateLiveTurn((prev) =>
              prev.map((m) => {
                if (m.id !== assistantId) return m;
                return {
                  ...m,
                  toolCalls: m.toolCalls?.map((tc) =>
                    tc.id === id ? { ...tc, status: "running" } : tc
                  ),
                  blocks: [...(m.blocks ?? []), { type: "tool_result", callId: id, status: "running" }],
                };
              })
            );
          },
          onToolExecutionDone: (id, result) => {
            updateLiveTurn((prev) =>
              applyToolExecutionResultToMessages(
                prev,
                assistantId,
                id,
                "success",
                result
              )
            );

            const toolCall = turnMessages
              .flatMap((m) => m.tool_calls ?? [])
              .find((tc) => tc.id === id);

            if (!toolCall) {
              return;
            }

            turnMessages.push({
              id: crypto.randomUUID(),
              role: "tool",
              content: result,
              tool_call_id: id,
              name: toolCall.function.name,
              ts: Date.now(),
            });
          },
          onToolExecutionError: (id, error) => {
            updateLiveTurn((prev) =>
              applyToolExecutionResultToMessages(
                prev,
                assistantId,
                id,
                "error",
                error
              )
            );

            const toolCall = turnMessages
              .flatMap((m) => m.tool_calls ?? [])
              .find((tc) => tc.id === id);

            if (!toolCall) {
              return;
            }

            turnMessages.push({
              id: crypto.randomUUID(),
              role: "tool",
              content: error,
              tool_call_id: id,
              name: toolCall.function.name,
              ts: Date.now(),
            });
          },
          onUsage: (prompt, completion) => setUsedTokens(prompt + completion),
          onError: (err) => {
            updateMsg(assistantId, {
              content: `⚠ Error: ${err.message}`,
              isStreaming: false,
              isReasoningStreaming: false,
            });
            setIsTyping(false);
            isProcessingRef.current = false;
            setTurnActive(false);
          },
          onDone: async () => {
            const hasFinalAssistantInTurn = turnMessages.some(
              (msg) =>
                msg.role === "assistant" &&
                !msg.tool_calls &&
                (msg.content ?? "") === assistantContent
            );
            if (assistantContent && !hasFinalAssistantInTurn) {
              turnMessages.push({
                id: crypto.randomUUID(),
                role: "assistant",
                content: assistantContent,
                reasoning: assistantReasoning || undefined,
                ts: Date.now(),
              });
            }

            suspendSessionRestore();
            try {
              finalizeAndCommitLiveTurn();
              setIsTyping(false);

              await persistCompletedTurn({
                projectKey: projectKeyRef.current,
                sessionId: session.id,
                userInput: value,
                sendTs,
                assistantContent,
                assistantReasoning,
                turnMessages,
              });
            } finally {
              resumeSessionRestore();
            }
            isProcessingRef.current = false;
            setTurnActive(false);
          },
        },
          fp,
          controller.signal
        );
      } finally {
        isProcessingRef.current = false;
        setTurnActive(false);
      }
    },
    [
      addPendingTurn,
      appendContentChunk,
      appendBlock,
      appendReasoningChunk,
      getActiveSessionSnapshot,
      initSession,
      persistCompletedTurn,
      replaceActiveController,
      setUsedTokens,
      setIsTyping,
      updateLiveTurn,
      finalizeAndCommitLiveTurn,
      startSessionNameGeneration,
      suspendSessionRestore,
      resumeSessionRestore,
      setTurnActive,
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
