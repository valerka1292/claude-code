import { useCallback, useEffect, useRef } from "react";
import type { Message } from "../types/message";
import { turnsToMessages } from "../lib/turnsToMessages";
import type { SessionData } from "../types/session";

function normalizeForComparison(message: Message) {
  const textFromBlocks = (message.blocks ?? [])
    .filter((block) => block.type === "text")
    .map((block) => block.content)
    .join("");
  const reasoningFromBlocks = (message.blocks ?? [])
    .filter((block) => block.type === "reasoning")
    .map((block) => block.content)
    .join("");

  return {
    role: message.role,
    content: textFromBlocks || message.content,
    reasoning: reasoningFromBlocks || message.reasoning || "",
    toolCallId: message.toolCallId ?? "",
    toolName: message.toolName ?? "",
    blocks: (message.blocks ?? []).map((block) => {
      if (block.type === "reasoning" || block.type === "text") {
        return { type: block.type, content: block.content };
      }
      if (block.type === "tool_call") {
        return {
          type: "tool_call" as const,
          id: block.call.id,
          name: block.call.name,
          arguments: block.call.arguments,
          status: block.call.status,
        };
      }
      return {
        type: "tool_result" as const,
        callId: block.callId,
        status: block.status,
        result: block.result ?? "",
      };
    }),
  };
}

export function areMessageListsEquivalent(
  currentMessages: Message[],
  nextMessages: Message[]
) {
  if (currentMessages.length !== nextMessages.length) {
    return false;
  }

  for (let index = 0; index < currentMessages.length; index += 1) {
    const current = normalizeForComparison(currentMessages[index]);
    const next = normalizeForComparison(nextMessages[index]);
    if (JSON.stringify(current) !== JSON.stringify(next)) {
      return false;
    }
  }

  return true;
}

export function useSessionRestore(
  activeSession: SessionData | null,
  replaceArchiveMessages: (nextMessages: Message[]) => void,
  shouldSkipRestore?: () => boolean,
  getCurrentArchiveMessages?: () => Message[]
) {
  const lastRestoredSignatureRef = useRef<string | null>(null);

  useEffect(() => {
    if (!activeSession) {
      replaceArchiveMessages([]);
      lastRestoredSignatureRef.current = null;
      return;
    }

    if (activeSession.messages.length === 0) {
      replaceArchiveMessages([]);
      lastRestoredSignatureRef.current = `${activeSession.id}:0`;
      return;
    }

    const signature = `${activeSession.id}:${activeSession.messages.length}`;
    if (lastRestoredSignatureRef.current === signature) {
      return;
    }
    if (shouldSkipRestore?.()) {
      return;
    }

    lastRestoredSignatureRef.current = signature;
    const restoredMessages = turnsToMessages(activeSession.messages);
    const currentArchiveMessages = getCurrentArchiveMessages?.() ?? [];

    if (areMessageListsEquivalent(currentArchiveMessages, restoredMessages)) {
      return;
    }

    replaceArchiveMessages(restoredMessages);
  }, [
    activeSession,
    activeSession?.messages.length,
    getCurrentArchiveMessages,
    replaceArchiveMessages,
    shouldSkipRestore,
  ]);

  const resetSessionRestore = useCallback(() => {
    lastRestoredSignatureRef.current = null;
  }, []);

  return { resetSessionRestore };
}
