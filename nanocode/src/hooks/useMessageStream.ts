import { useCallback, useEffect, useRef, useState } from "react";
import type { Message } from "../components/MessageItem";
import type { SessionData } from "../types/session";
import { storedToUiMessage } from "../lib/converters";

export function useMessageStream(activeSession: SessionData | null) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [usedTokens, setUsedTokens] = useState(0);
  const lastRestoredIdRef = useRef<string | null>(null);

  useEffect(() => {
    if (!activeSession) {
      setMessages([]);
      lastRestoredIdRef.current = null;
      return;
    }

    if (
      lastRestoredIdRef.current === activeSession.id ||
      activeSession.messages.length === 0
    ) {
      return;
    }

    lastRestoredIdRef.current = activeSession.id;
    setMessages(
      activeSession.messages
        .filter((m) => m.role === "user" || m.role === "assistant")
        .map(storedToUiMessage)
    );
  }, [activeSession]);

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
    lastRestoredIdRef.current = null;
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
  };
}
