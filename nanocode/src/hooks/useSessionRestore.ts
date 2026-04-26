import { useCallback, useEffect, useRef } from "react";
import type { Message } from "../types/message";
import { turnsToMessages } from "../lib/turnsToMessages";
import type { SessionData } from "../types/session";

export function useSessionRestore(
  activeSession: SessionData | null,
  replaceMessages: (nextMessages: Message[]) => void
) {
  const lastRestoredSignatureRef = useRef<string | null>(null);

  useEffect(() => {
    if (!activeSession) {
      replaceMessages([]);
      lastRestoredSignatureRef.current = null;
      return;
    }

    if (activeSession.messages.length === 0) {
      lastRestoredSignatureRef.current = `${activeSession.id}:0`;
      return;
    }

    const signature = `${activeSession.id}:${activeSession.messages.length}`;
    if (lastRestoredSignatureRef.current === signature) {
      return;
    }

    lastRestoredSignatureRef.current = signature;
    replaceMessages(turnsToMessages(activeSession.messages));
  }, [activeSession, activeSession?.messages.length, replaceMessages]);

  const resetSessionRestore = useCallback(() => {
    lastRestoredSignatureRef.current = null;
  }, []);

  return { resetSessionRestore };
}
