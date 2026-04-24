/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useRef, useEffect, useCallback } from "react";
import { TitleBar } from "./components/TitleBar";
import { Sidebar } from "./components/Sidebar";
import { MainArea } from "./components/MainArea";
import { InputContainer, type InputContainerHandle } from "./components/InputContainer";
import { type Mode } from "./components/ModeDropdown";
import { type Message, MessageItem } from "./components/MessageItem";
import { SettingsModal } from "./components/settings/SettingsModal";
import { motion } from "motion/react";
import { useProviders } from "./contexts/ProvidersContext";
import { useProject } from "./contexts/ProjectContext";
import { useSession } from "./contexts/SessionContext";
import { runAgentStream, type ChatMessage } from "./lib/agentLoop";
import { buildSystemMessages } from "./lib/systemPrompt";
import {
  generateSessionName,
  saveSession,
  storedToChat,
  type StoredMessage,
  type SessionData,
} from "./lib/sessions";

export default function App() {
  const { activeProvider } = useProviders();
  const { folderPath, folderName, projectKey } = useProject();
  const {
    activeSession,
    initSession,
    updateSession,
    startNewSession,
    openSession,
    sessionList,
    isLoadingList,
    removeSession,
  } = useSession();

  const [mode, setMode] = useState<Mode>("Ask");
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [usedTokens, setUsedTokens] = useState(0);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const inputRef = useRef<InputContainerHandle>(null);
  const lastRestoredIdRef = useRef<string | null>(null);

  const activeSessionRef = useRef<SessionData | null>(null);
  activeSessionRef.current = activeSession;

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
    const uiMessages: Message[] = activeSession.messages
      .filter((m) => m.role === "user" || m.role === "assistant")
      .map((m) => ({
        id: `${m.ts}-${m.role}`,
        role: m.role as "user" | "assistant",
        content: m.content ?? "",
        reasoning: (m as StoredMessage).reasoning,
      }));
    setMessages(uiMessages);
  }, [activeSession?.id]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isTyping]);

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
      const isFirstMessage = !activeSessionRef.current;

      const session: SessionData =
        activeSessionRef.current ?? initSession(fp);

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

      const sessionNameRef = { current: session.name };
      if (isFirstMessage) {
        generateSessionName(ap, value).then((name) => {
          sessionNameRef.current = name;
          updateSession({ ...session, name });
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

      runAgentStream(ap, history, {
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

          const finalSession: SessionData = {
            ...session,
            name: sessionNameRef.current,
            messages: [...session.messages, userStored, assistantStored],
          };

          await saveSession(currentPk, finalSession);
          updateSession(finalSession);
        },
      }, controller.signal);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [initSession, updateSession, updateMsg]
  );

  const handleNewSession = useCallback(() => {
    abortRef.current?.abort();
    startNewSession();
    setMessages([]);
    setUsedTokens(0);
    lastRestoredIdRef.current = null;
  }, [startNewSession]);

  const handleOpenSession = useCallback(
    async (id: string) => {
      abortRef.current?.abort();
      setMessages([]);
      setUsedTokens(0);
      lastRestoredIdRef.current = null;
      await openSession(id);
    },
    [openSession]
  );

  const inputNode = (
    <InputContainer
      ref={inputRef}
      mode={mode}
      setMode={setMode}
      onSend={handleSend}
      usedTokens={usedTokens}
    />
  );

  return (
    <div className="flex flex-col h-screen w-screen bg-[#0d0d0d] text-white overflow-hidden select-none font-mono">
      <TitleBar />

      <div className="flex flex-1 overflow-hidden">
        <Sidebar
          onSettingsClick={() => setSettingsOpen(true)}
          onNewSession={handleNewSession}
          onOpenSession={handleOpenSession}
          onDeleteSession={removeSession}
          sessionList={sessionList}
          activeSessionId={activeSession?.id ?? null}
          isLoadingList={isLoadingList}
        />

        <div className="flex flex-col flex-1 overflow-hidden">
          <MainArea
            hasMessages={messages.length > 0}
            input={inputNode}
            inputRef={inputRef}
          >
            {messages.map((msg) => (
              <MessageItem key={msg.id} message={msg} />
            ))}

            {isTyping && messages.at(-1)?.role !== "assistant" && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="flex gap-3 py-3 px-5"
              >
                <div className="w-5 h-5 rounded-md bg-white flex items-center justify-center flex-shrink-0 mt-0.5">
                  <div className="w-2 h-2 bg-black rounded-full" />
                </div>
                <div className="flex-1">
                  <div className="text-[0.68rem] font-mono text-white/20 mb-1 uppercase tracking-wider">
                    nanocode
                  </div>
                  <div className="flex gap-1 items-center h-5">
                    {[0, 0.15, 0.3].map((delay, i) => (
                      <span
                        key={i}
                        className="w-1 h-1 rounded-full bg-white/30 animate-pulse"
                        style={{ animationDelay: `${delay}s` }}
                      />
                    ))}
                  </div>
                </div>
              </motion.div>
            )}

            <div ref={messagesEndRef} />
          </MainArea>
        </div>
      </div>

      <SettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </div>
  );
}