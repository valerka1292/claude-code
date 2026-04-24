import { useState, useRef, useEffect, useCallback } from "react";
import type { Message } from "../components/MessageItem";
import type { Mode } from "../types";
import { useProviders } from "../contexts/ProvidersContext";
import { useProject } from "../contexts/ProjectContext";
import { useSession } from "../contexts/SessionContext";
import { runAgentStream, type ChatMessage } from "../lib/agentLoop";
import { buildSystemMessages } from "../lib/systemPrompt";
import { generateSessionName, saveSession, type StoredMessage } from "../lib/sessions";
import { storedToChat, storedToUiMessage } from "../lib/converters";

export function useAgent() {
  const { activeProvider } = useProviders();
  const { folderPath, folderName, projectKey } = useProject();
  const {
    activeSession,
    initSession,
    updateSession,
    getActiveSessionSnapshot,
    onSessionSaveError,
  } = useSession();

  const [mode, setMode] = useState<Mode>("Ask");
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [usedTokens, setUsedTokens] = useState(0);

  const abortRef = useRef<AbortController | null>(null);
  const lastRestoredIdRef = useRef<string | null>(null);

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

  const updateMsg = useCallback(
    (id: string, patch: Partial<Message>) =>
      setMessages((prev) =>
        prev.map((m) => (m.id === id ? { ...m, ...patch } : m))
      ),
    []
  );

  const handleSend = useCallback(
    async (value: string) => {
      const fp = folderPathRef.current;
      const pk = projectKeyRef.current;
      const ap = activeProviderRef.current;

      if (!value.trim() || !fp || !pk) return;

      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      const sendTs = Date.now();
      const isFirstMessage = !getActiveSessionSnapshot();

      const session = getActiveSessionSnapshot() ?? initSession(fp);

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

      let sessionName = session.name;
      if (isFirstMessage) {
        generateSessionName(ap, value)
          .then((name) => {
            sessionName = name;
            const current = getActiveSessionSnapshot();
            if (!current || current.id !== session.id) return;
            updateSession({ ...current, name });
          })
          .catch((error) => {
            console.error("[agent] generateSessionName error:", error);
          });
      }

      const fn = folderNameRef.current;
      const md = modeRef.current;

      const history: ChatMessage[] = [
        ...buildSystemMessages({
          cwd: fp,
          projectName: fn ?? fp,
          mode: md,
        }),
        ...session.messages.map(storedToChat),
        { role: "user", content: value },
      ];

      let assistantContent = "";
      let assistantReasoning = "";

      runAgentStream(
        ap,
        history,
        {
          onReasoningChunk: (chunk) => {
            assistantReasoning += chunk;
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantId
                  ? {
                      ...m,
                      reasoning: (m.reasoning ?? "") + chunk,
                      isReasoningStreaming: true,
                    }
                  : m
              )
            );
          },

          onContentChunk: (chunk) => {
            assistantContent += chunk;
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantId
                  ? {
                      ...m,
                      content: (m.content ?? "") + chunk,
                      isReasoningStreaming: false,
                      isStreaming: true,
                    }
                  : m
              )
            );
          },

          onToolCallStart: (_id, name) => {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantId
                  ? { ...m, content: (m.content ?? "") + `\n[Tool: ${name}]` }
                  : m
              )
            );
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

            const currentPk = projectKeyRef.current;
            if (!currentPk) return;

            const userStored: StoredMessage = {
              role: "user",
              content: value,
              ts: sendTs,
            };

            const assistantStored: StoredMessage = {
              role: "assistant",
              content: assistantContent || null,
              reasoning: assistantReasoning || undefined,
              ts: Date.now(),
            };

            const latestSession = getActiveSessionSnapshot();
            if (!latestSession || latestSession.id !== session.id) return;

            const finalSession = {
              ...latestSession,
              name: sessionName,
              messages: [...latestSession.messages, userStored, assistantStored],
            };

            try {
              await saveSession(currentPk, finalSession);
              updateSession(finalSession);
            } catch (error) {
              onSessionSaveError(error);
            }
          },
        },
        controller.signal
      );
    },
    [getActiveSessionSnapshot, initSession, onSessionSaveError, updateMsg, updateSession]
  );

  const resetAgentUi = useCallback(() => {
    abortRef.current?.abort();
    setMessages([]);
    setUsedTokens(0);
    lastRestoredIdRef.current = null;
  }, []);

  return {
    mode,
    setMode,
    messages,
    isTyping,
    usedTokens,
    handleSend,
    resetAgentUi,
    abortActiveRequest: () => abortRef.current?.abort(),
  };
}
