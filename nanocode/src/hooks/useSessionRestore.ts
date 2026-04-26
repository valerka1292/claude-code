import { useCallback, useEffect, useRef } from "react";
import type { Message } from "../types/message";
import { turnsToMessages } from "../lib/turnsToMessages";
import type { SessionData } from "../types/session";

export function useSessionRestore(
  activeSession: SessionData | null,
  replaceMessages: (nextMessages: Message[]) => void
) {
  const lastRestoredIdRef = useRef<string | null>(null);

  useEffect(() => {
    if (!activeSession) {
      replaceMessages([]);
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
    replaceMessages(turnsToMessages(activeSession.messages));
  }, [activeSession, replaceMessages]);

  const resetSessionRestore = useCallback(() => {
    lastRestoredIdRef.current = null;
  }, []);

  return { resetSessionRestore };
}
