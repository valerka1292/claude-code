import { useCallback, useState } from "react";
import type { Message } from "../components/MessageItem";

export function useMessageStream() {
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

  const appendToolCallLabel = useCallback((id: string, name: string) => {
    setMessages((prev) =>
      prev.map((m) =>
        m.id === id
          ? { ...m, content: (m.content ?? "") + `\n[Tool: ${name}]` }
          : m
      )
    );
  }, []);

  const resetMessageStream = useCallback(() => {
    setMessages([]);
    setUsedTokens(0);
  }, []);

  const replaceMessages = useCallback((nextMessages: Message[]) => {
    setMessages(nextMessages);
  }, []);

  return {
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
    replaceMessages,
  };
}
