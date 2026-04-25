import { useCallback, useEffect, useRef } from "react";
import type { Message } from "../types/message";
import { storedToUiMessage } from "../lib/converters";
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
    replaceMessages(
      activeSession.messages
        .filter((m) => m.role === "user" || m.role === "assistant")
        .map((m, index) => {
          const normalizedId =
            typeof m.id === "string" && m.id.length > 0
              ? m.id
              : `${m.ts}-${m.role}-${index}`;

          return storedToUiMessage({
            ...m,
            id: normalizedId,
          });
        })
    );
  }, [activeSession, replaceMessages]);

  const resetSessionRestore = useCallback(() => {
    lastRestoredIdRef.current = null;
  }, []);

  return { resetSessionRestore };
}
