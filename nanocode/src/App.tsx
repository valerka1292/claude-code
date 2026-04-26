/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useRef, useEffect, useCallback, useState } from "react";
import { TitleBar } from "./components/TitleBar";
import { Sidebar } from "./components/Sidebar";
import { MainArea } from "./components/MainArea";
import { InputContainer, type InputContainerHandle } from "./components/InputContainer";
import { MessageItem } from "./components/MessageItem";
import { SettingsModal } from "./components/settings/SettingsModal";
import { motion } from "motion/react";
import { useSession } from "./contexts/SessionContext";
import { useAgent } from "./hooks/useAgent";
import { useProviders } from "./contexts/ProvidersContext";
import { Loader2 } from "lucide-react";

export default function App() {
  const { isLoading: isProvidersLoading } = useProviders();

  const {
    activeSession,
    startNewSession,
    openSession,
    sessionList,
    isLoadingList,
    removeSession,
  } = useSession();

  const {
    mode,
    setMode,
    messages,
    isTyping,
    usedTokens,
    handleSend,
    resetAgentUi,
  } = useAgent();

  const [settingsOpen, setSettingsOpen] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<InputContainerHandle>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length, isTyping, messages.at(-1)?.content?.length]);

  const handleNewSession = useCallback(() => {
    resetAgentUi();
    startNewSession();
  }, [resetAgentUi, startNewSession]);

  const handleOpenSession = useCallback(
    async (id: string) => {
      resetAgentUi();
      await openSession(id);
    },
    [openSession, resetAgentUi]
  );

  if (isProvidersLoading) {
    return (
      <div className="flex items-center justify-center h-screen w-screen bg-[#0d0d0d] text-white font-mono">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="animate-spin text-white/20" size={24} />
          <span className="text-[0.7rem] text-white/20 uppercase tracking-widest">Loading Environment</span>
        </div>
      </div>
    );
  }

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
