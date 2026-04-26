import React, { useCallback, useEffect, useRef, useState } from 'react';
import Sidebar from './components/Sidebar';
import MessageList from './components/MessageList';
import InputArea from './components/InputArea';
import SettingsModal from './components/SettingsModal';
import Titlebar from './components/Titlebar';
import { AgentMode, Message, ToolCall } from './types';
import { generateChatName, streamChatCompletion } from './services/llmService';
import { useProviders } from './hooks/useProviders';
import { useChats } from './hooks/useChats';

export default function App() {
  const {
    chats,
    activeChatId,
    activeChatIdRef,
    messages,
    setMessages,
    llmHistory,
    setLlmHistory,
    contextTokensUsed,
    setContextTokensUsed,
    isTyping,
    setIsTyping,
    switchChat,
    loadChat,
    newChat,
    createChat,
    renameChat,
    deleteChat,
    persistCurrentChat,
  } = useChats();

  const [mode, setMode] = useState<AgentMode>('Chat');
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isAgentRunning, setIsAgentRunning] = useState(false);
  const { activeProvider } = useProviders();
  const pendingRequestControllerRef = useRef<AbortController | null>(null);

  const cancelPendingRequest = useCallback(() => {
    pendingRequestControllerRef.current?.abort();
    pendingRequestControllerRef.current = null;
  }, []);

  useEffect(() => {
    return () => {
      cancelPendingRequest();
    };
  }, [cancelPendingRequest]);

  const handleSend = useCallback(
    async (content: string) => {
      if (!activeProvider) {
        return;
      }

      cancelPendingRequest();
      const requestController = new AbortController();
      pendingRequestControllerRef.current = requestController;

      let chatId = activeChatIdRef.current;
      if (!chatId) {
        chatId = await createChat();
        await loadChat(chatId);
      }

      const chatIdSnapshot = chatId;
      const isFirstMessage = messages.length === 0;

      const userMsg: Message = {
        id: Date.now().toString(),
        role: 'user',
        content,
        timestamp: new Date(),
      };

      setMessages((prev) => [...prev, userMsg]);

      const nextLlmHistory = llmHistory.concat([{ role: 'user', content }]);
      setLlmHistory(nextLlmHistory);

      const assistantId = (Date.now() + 1).toString();
      const assistantMsg: Message = {
        id: assistantId,
        role: 'assistant',
        content: '',
        timestamp: new Date(),
        thinking: '',
        toolCalls: [],
      };
      setMessages((prev) => [...prev, assistantMsg]);
      setIsTyping(true);

      let finalAssistantContent = '';

      streamChatCompletion(
        activeProvider,
        nextLlmHistory,
        {
          onContent: (text) => {
            if (activeChatIdRef.current !== chatIdSnapshot) {
              return;
            }

            finalAssistantContent += text;
            setMessages((prev) => prev.map((m) => (m.id === assistantId ? { ...m, content: m.content + text } : m)));
          },
          onThinking: (text) => {
            if (activeChatIdRef.current !== chatIdSnapshot) {
              return;
            }

            setMessages((prev) =>
              prev.map((m) => (m.id === assistantId ? { ...m, thinking: (m.thinking ?? '') + text } : m)),
            );
          },
          onToolCall: (toolCall) => {
            if (activeChatIdRef.current !== chatIdSnapshot) {
              return;
            }

            setMessages((prev) =>
              prev.map((m) => {
                if (m.id !== assistantId) {
                  return m;
                }

                const existingToolCalls: ToolCall[] = m.toolCalls ? [...m.toolCalls] : [];
                const idx = existingToolCalls.findIndex((tc) => tc.index === toolCall.index);

                if (idx >= 0) {
                  existingToolCalls[idx] = { ...existingToolCalls[idx], ...toolCall };
                } else {
                  existingToolCalls.push(toolCall);
                }

                return { ...m, toolCalls: existingToolCalls };
              }),
            );
          },
          onDone: async (usage) => {
            if (pendingRequestControllerRef.current === requestController) {
              pendingRequestControllerRef.current = null;
            }

            if (activeChatIdRef.current !== chatIdSnapshot) {
              return;
            }

            setIsTyping(false);
            if (usage?.total_tokens) {
              setContextTokensUsed(usage.total_tokens);
            }
            setLlmHistory((prev) => [...prev, { role: 'assistant', content: finalAssistantContent }]);
            await persistCurrentChat();

            if (isFirstMessage) {
              const title = await generateChatName(activeProvider, content);
              if (activeChatIdRef.current === chatIdSnapshot) {
                await renameChat(chatIdSnapshot, title);
              }
            }
          },
          onError: (error) => {
            if (pendingRequestControllerRef.current === requestController) {
              pendingRequestControllerRef.current = null;
            }

            if (activeChatIdRef.current !== chatIdSnapshot) {
              return;
            }

            setIsTyping(false);
            const errorMessage = error instanceof Error ? error.message : String(error);
            setMessages((prev) =>
              prev.map((m) => (m.id === assistantId ? { ...m, content: `Error: ${errorMessage}` } : m)),
            );
          },
        },
        requestController.signal,
      );
    },
    [
      activeChatIdRef,
      activeProvider,
      cancelPendingRequest,
      createChat,
      llmHistory,
      loadChat,
      messages.length,
      persistCurrentChat,
      renameChat,
      setContextTokensUsed,
      setIsTyping,
      setLlmHistory,
      setMessages,
    ],
  );

  const handleNewChat = useCallback(async () => {
    cancelPendingRequest();
    await newChat();
  }, [cancelPendingRequest, newChat]);

  const handleSwitchChat = useCallback(async (chatId: string) => {
    cancelPendingRequest();
    await switchChat(chatId);
  }, [cancelPendingRequest, switchChat]);

  const handleDeleteChat = useCallback(async (chatId: string) => {
    cancelPendingRequest();
    await deleteChat(chatId);
  }, [cancelPendingRequest, deleteChat]);

  const chatTitle = chats.find((chat) => chat.id === activeChatId)?.title ?? 'Untitled Chat';

  return (
    <div className="flex h-screen w-full select-none overflow-hidden bg-bg-0 text-text-primary">
      <Sidebar
        activeChatId={activeChatId ?? ''}
        chats={chats}
        onNewChat={handleNewChat}
        onSelectChat={handleSwitchChat}
        onDeleteChat={handleDeleteChat}
        onSettingsOpen={() => setIsSettingsOpen(true)}
      />

      <main className="relative flex h-full flex-1 flex-col overflow-hidden">
        <Titlebar chatTitle={chatTitle} onRename={() => {}} />

        <MessageList messages={messages} isTyping={isTyping} />

        <InputArea
          mode={mode}
          onModeChange={setMode}
          onSend={handleSend}
          hasProvider={activeProvider !== null}
          isAgentRunning={isAgentRunning}
          onToggleAgent={() => setIsAgentRunning(!isAgentRunning)}
          contextTokensUsed={contextTokensUsed}
        />

        <SettingsModal isOpen={isSettingsOpen} onClose={() => setIsSettingsOpen(false)} />
      </main>
    </div>
  );
}
