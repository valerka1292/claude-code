import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Message, ToolCallDisplay, ContentBlock } from "../types/message";
import type { SessionData } from "../types/session";
import { useSessionRestore } from "./useSessionRestore";

export function finalizeLiveTurnMessages(liveTurn: Message[]): Message[] {
  return liveTurn.map((m) => {
    const blocks = m.blocks ?? [];
    const updatedBlocks = blocks.map((b) => {
      if (b.type === "reasoning") return { ...b, streaming: false };
      if (b.type === "text") return { ...b, streaming: false };
      return b;
    });
    return {
      ...m,
      blocks: updatedBlocks,
      isStreaming: false,
      isReasoningStreaming: false,
    };
  });
}

export function useMessageStream(activeSession: SessionData | null) {
  const [archiveMessages, setArchiveMessages] = useState<Message[]>([]);
  const [liveTurn, setLiveTurn] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [usedTokens, setUsedTokens] = useState(0);
  const sessionRestoreSuspendedRef = useRef(false);
  const archiveMessagesRef = useRef<Message[]>([]);
  const messages = useMemo(
    () => [...archiveMessages, ...liveTurn],
    [archiveMessages, liveTurn]
  );
  useEffect(() => {
    archiveMessagesRef.current = archiveMessages;
  }, [archiveMessages]);

  const updateLiveTurn = useCallback(
    (updater: (prev: Message[]) => Message[]) => {
      setLiveTurn((prev) => updater(prev));
    },
    []
  );

  const addPendingTurn = useCallback((value: string, sendTs: number) => {
    const userMsgId = `${sendTs}-user`;
    const assistantId = `${sendTs + 1}-assistant`;

    setLiveTurn((prev) => [
      ...prev,
      { id: userMsgId, role: "user", content: value },
      {
        id: assistantId,
        role: "assistant",
        content: "",
        reasoning: "",
        isStreaming: true,
        isReasoningStreaming: false,
        blocks: [],
      },
    ]);

    return assistantId;
  }, []);

  const updateMsg = useCallback(
    (id: string, patch: Partial<Message>) =>
      updateLiveTurn((prev) =>
        prev.map((m) => (m.id === id ? { ...m, ...patch } : m))
      ),
    [updateLiveTurn]
  );

  const appendReasoningChunk = useCallback((id: string, chunk: string) => {
    updateLiveTurn((prev) =>
      prev.map((m) =>
        m.id === id
          ? {
              ...m,
              reasoning: (m.reasoning ?? "") + chunk,
              isReasoningStreaming: true,
            }
          : m
      )
    );
  }, [updateLiveTurn]);

  const appendContentChunk = useCallback((id: string, chunk: string) => {
    updateLiveTurn((prev) =>
      prev.map((m) =>
        m.id === id
          ? {
              ...m,
              content: (m.content ?? "") + chunk,
              isReasoningStreaming: false,
              isStreaming: true,
            }
          : m
      )
    );
  }, [updateLiveTurn]);

  const addToolCall = useCallback(
    (msgId: string, call: ToolCallDisplay) => {
      updateLiveTurn((prev) =>
        prev.map((m) =>
          m.id === msgId
            ? {
                ...m,
                toolCalls: [...(m.toolCalls ?? []), call],
              }
            : m
        )
      );
    },
    [updateLiveTurn]
  );

  const updateToolCallStatus = useCallback(
    (
      msgId: string,
      callId: string,
      status: ToolCallDisplay["status"],
      result?: string
    ) => {
      updateLiveTurn((prev) =>
        prev.map((m) => {
          if (m.id !== msgId || !m.toolCalls) return m;
          return {
            ...m,
            toolCalls: m.toolCalls.map((tc) =>
              tc.id === callId ? { ...tc, status, result } : tc
            ),
          };
        })
      );
    },
    [updateLiveTurn]
  );

  const appendBlock = useCallback((msgId: string, block: ContentBlock) => {
    updateLiveTurn((prev) =>
      prev.map((m) => {
        if (m.id !== msgId) return m;
        const blocks = m.blocks ?? [];

        if (block.type === "reasoning") {
          const last = blocks[blocks.length - 1];
          if (last && last.type === "reasoning" && last.streaming) {
            return {
              ...m,
              blocks: [
                ...blocks.slice(0, -1),
                { ...last, content: last.content + block.content },
              ],
            };
          }
          return { ...m, blocks: [...blocks, block] };
        }

        if (block.type === "text") {
          const last = blocks[blocks.length - 1];
          if (last && last.type === "text" && last.streaming) {
            return {
              ...m,
              blocks: [
                ...blocks.slice(0, -1),
                { ...last, content: last.content + block.content },
              ],
            };
          }
          return { ...m, blocks: [...blocks, block] };
        }

        return { ...m, blocks: [...blocks, block] };
      })
    );
  }, [updateLiveTurn]);

  const replaceArchiveMessages = useCallback((nextMessages: Message[]) => {
    setArchiveMessages(nextMessages);
  }, []);
  const { resetSessionRestore } = useSessionRestore(
    activeSession,
    replaceArchiveMessages,
    () => sessionRestoreSuspendedRef.current,
    () => archiveMessagesRef.current
  );

  const finalizeAndCommitLiveTurn = useCallback(() => {
    let finalizedTurn: Message[] = [];
    setLiveTurn((prev) => {
      finalizedTurn = finalizeLiveTurnMessages(prev);
      setArchiveMessages((archive) => [...archive, ...finalizedTurn]);
      return [];
    });
    return finalizedTurn;
  }, []);

  const resetMessageStream = useCallback(() => {
    resetSessionRestore();
    setArchiveMessages([]);
    archiveMessagesRef.current = [];
    setLiveTurn([]);
    setUsedTokens(0);
  }, [resetSessionRestore]);

  const commitLiveTurnAndPersist = useCallback(
    async (persist: () => Promise<void>) => {
      sessionRestoreSuspendedRef.current = true;
      try {
        const finalizedTurn = finalizeAndCommitLiveTurn();
        await persist();
        return finalizedTurn;
      } finally {
        sessionRestoreSuspendedRef.current = false;
      }
    },
    [finalizeAndCommitLiveTurn]
  );

  return {
    messages,
    archiveMessages,
    liveTurn,
    updateLiveTurn,
    isTyping,
    setIsTyping,
    usedTokens,
    setUsedTokens,
    addPendingTurn,
    updateMsg,
    appendReasoningChunk,
    appendContentChunk,
    addToolCall,
    updateToolCallStatus,
    appendBlock,
    finalizeAndCommitLiveTurn,
    commitLiveTurnAndPersist,
    resetMessageStream,
  };
}
