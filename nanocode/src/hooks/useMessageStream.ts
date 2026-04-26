import { useCallback, useState } from "react";
import type { Message, ToolCallDisplay, ContentBlock } from "../types/message";
import type { SessionData } from "../types/session";
import { useSessionRestore } from "./useSessionRestore";

export function useMessageStream(activeSession: SessionData | null) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [usedTokens, setUsedTokens] = useState(0);

  const addPendingTurn = useCallback((value: string, sendTs: number) => {
    const userMsgId = `${sendTs}-user`;
    const assistantId = `${sendTs + 1}-assistant`;

    setMessages((prev) => [
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
      setMessages((prev) =>
        prev.map((m) => (m.id === id ? { ...m, ...patch } : m))
      ),
    []
  );

  const appendReasoningChunk = useCallback((id: string, chunk: string) => {
    setMessages((prev) =>
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
  }, []);

  const appendContentChunk = useCallback((id: string, chunk: string) => {
    setMessages((prev) =>
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
  }, []);

  const addToolCall = useCallback(
    (msgId: string, call: ToolCallDisplay) => {
      setMessages((prev) =>
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
    []
  );

  const updateToolCallStatus = useCallback(
    (
      msgId: string,
      callId: string,
      status: ToolCallDisplay["status"],
      result?: string
    ) => {
      setMessages((prev) =>
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
    []
  );

  const appendBlock = useCallback((msgId: string, block: ContentBlock) => {
    setMessages((prev) =>
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
  }, []);

  const replaceMessages = useCallback((nextMessages: Message[]) => {
    setMessages(nextMessages);
  }, []);
  const { resetSessionRestore } = useSessionRestore(activeSession, replaceMessages);

  const resetMessageStream = useCallback(() => {
    resetSessionRestore();
    setMessages([]);
    setUsedTokens(0);
  }, [resetSessionRestore]);

  return {
    messages,
    setMessages,
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
    resetMessageStream,
  };
}