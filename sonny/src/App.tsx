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
    isLoading,
    activeChatId,
    activeChatIdRef,
    messages,
    setMessages,
    llmHistory,
    setLlmHistory,
    contextTokensUsed,
    setContextTokensUsed,
    messagesRef,
    llmHistoryRef,
    contextTokensUsedRef,
    isTyping,
    setIsTyping,
    switchChat,
    loadChat,
    newChat,
    createChat,
    renameChat,
    deleteChat,
    persistChatData,
    scheduleAutoSave,
  } = useChats();

  const [mode, setMode] = useState<AgentMode>('Chat');
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isAgentRunning, setIsAgentRunning] = useState(false);
  const { activeProvider } = useProviders();
  const pendingRequestControllerRef = useRef<AbortController | null>(null);

  const cancelPendingRequest = useCallback(() => {
    console.log('[App] Cancelling pending request');
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
      console.log(`[App] handleSend triggered: "${content.slice(0, 50)}..."`);
      if (!activeProvider) {
        console.warn('[App] No active provider, aborting send');
        return;
      }

      cancelPendingRequest();
      const requestController = new AbortController();
      pendingRequestControllerRef.current = requestController;

      let chatId = activeChatIdRef.current;
      if (!chatId) {
        console.log('[App] No active chat, creating new one');
        chatId = await createChat();
      }

      const chatIdSnapshot = chatId;
      const existingMessages = [...messagesRef.current];
      const isFirstMessage = existingMessages.length === 0;

      const userMsg: Message = {
        id: Date.now().toString(),
        role: 'user',
        content,
        timestamp: new Date(),
      };

      console.log(`[App] Adding user message to chat ${chatIdSnapshot}`);
      const nextMessagesBase = [...existingMessages, userMsg];
      setMessages(nextMessagesBase);

      const nextLlmHistory = [...llmHistoryRef.current, { role: 'user', content }];
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

      let workingMessages: Message[] = [...nextMessagesBase, assistantMsg];
      setMessages(workingMessages);
      setIsTyping(true);

      let finalAssistantContent = '';
      let finalAssistantThinking = '';
      const finalToolCalls = new Map<number, ToolCall>();

      console.log('[App] Starting LLM stream...');
      await streamChatCompletion(
        activeProvider,
        nextLlmHistory,
        {
          onContent: (text) => {
            if (activeChatIdRef.current !== chatIdSnapshot) {
              console.warn('[App] onContent ignored: chat changed');
              return;
            }

            finalAssistantContent += text;
            workingMessages = workingMessages.map((m) =>
              m.id === assistantId ? { ...m, content: m.content + text } : m,
            );
            setMessages(workingMessages);
            scheduleAutoSave();
          },
          onThinking: (text) => {
            if (activeChatIdRef.current !== chatIdSnapshot) return;

            finalAssistantThinking += text;
            workingMessages = workingMessages.map((m) =>
              m.id === assistantId ? { ...m, thinking: (m.thinking ?? '') + text } : m,
            );
            setMessages(workingMessages);
            scheduleAutoSave();
          },
          onToolCall: (toolCall) => {
            if (activeChatIdRef.current !== chatIdSnapshot) return;

            console.log(`[App] Tool call update for index ${toolCall.index}`);
            finalToolCalls.set(toolCall.index, { ...(finalToolCalls.get(toolCall.index) ?? {}), ...toolCall });

            workingMessages = workingMessages.map((m) => {
              if (m.id !== assistantId) return m;

              const existingToolCalls: ToolCall[] = m.toolCalls ? [...m.toolCalls] : [];
              const idx = existingToolCalls.findIndex((tc) => tc.index === toolCall.index);

              if (idx >= 0) {
                existingToolCalls[idx] = { ...existingToolCalls[idx], ...toolCall };
              } else {
                existingToolCalls.push(toolCall);
              }

              return { ...m, toolCalls: existingToolCalls };
            });
            setMessages(workingMessages);
            scheduleAutoSave();
          },
          onDone: async (usage) => {
            console.log('[App] Stream done');
            if (pendingRequestControllerRef.current === requestController) {
              pendingRequestControllerRef.current = null;
            }

            if (activeChatIdRef.current !== chatIdSnapshot) {
              console.warn('[App] onDone ignored: chat changed');
              return;
            }

            setIsTyping(false);
            const finalTokens = usage?.total_tokens ?? contextTokensUsedRef.current;
            if (usage?.total_tokens) {
              setContextTokensUsed(usage.total_tokens);
            }

            const finalAssistantMsg: Message = {
              ...assistantMsg,
              content: finalAssistantContent,
              thinking: finalAssistantThinking,
              toolCalls: Array.from(finalToolCalls.values()),
            };

            const finalMessages = [...nextMessagesBase, finalAssistantMsg];
            const finalLlmHistory = [...nextLlmHistory, { role: 'assistant', content: finalAssistantContent }];

            setMessages(finalMessages);
            setLlmHistory(finalLlmHistory);
            
            console.log(`[App] Persisting final chat data for ${chatIdSnapshot}`);
            await persistChatData(chatIdSnapshot, finalMessages, finalLlmHistory, finalTokens);

            if (isFirstMessage) {
              console.log('[App] Generating title for new chat');
              const title = await generateChatName(activeProvider, content, requestController.signal);
              if (activeChatIdRef.current === chatIdSnapshot) {
                await renameChat(chatIdSnapshot, title);
              }
            }
          },
          onError: async (error) => {
            console.error('[App] Stream error:', error);
            if (pendingRequestControllerRef.current === requestController) {
              pendingRequestControllerRef.current = null;
            }

            if (activeChatIdRef.current !== chatIdSnapshot) return;

            setIsTyping(false);
            const errorMessage = error instanceof Error ? error.message : String(error);
            const errorAssistantMsg: Message = {
              ...assistantMsg,
              content: finalAssistantContent || `Error: ${errorMessage}`,
              thinking: finalAssistantThinking,
              toolCalls: Array.from(finalToolCalls.values()),
            };
            const erroredMessages = [...nextMessagesBase, errorAssistantMsg];
            setMessages(erroredMessages);
            await persistChatData(chatIdSnapshot, erroredMessages, nextLlmHistory, contextTokensUsedRef.current);
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
      persistChatData,
      renameChat,
      scheduleAutoSave,
      messagesRef,
      llmHistoryRef,
      contextTokensUsedRef,
      setContextTokensUsed,
      setIsTyping,
      setLlmHistory,
      setMessages,
    ],
  );

  const handleNewChat = useCallback(async () => {
    console.log('[App] handleNewChat');
    cancelPendingRequest();
    await newChat();
  }, [cancelPendingRequest, newChat]);

  const handleSwitchChat = useCallback(async (chatId: string) => {
    console.log(`[App] handleSwitchChat: ${chatId}`);
    cancelPendingRequest();
    await switchChat(chatId);
  }, [cancelPendingRequest, switchChat]);

  const handleDeleteChat = useCallback(async (chatId: string) => {
    console.log(`[App] handleDeleteChat: ${chatId}`);
    cancelPendingRequest();
    await deleteChat(chatId);
  }, [cancelPendingRequest, deleteChat]);


  const handleRenameChat = useCallback(async () => {
    if (!activeChatId) return;

    const currentTitle = chats.find((chat) => chat.id === activeChatId)?.title ?? 'Untitled Chat';
    const nextTitle = window.prompt('Rename chat', currentTitle);
    if (nextTitle === null) return;

    console.log(`[App] handleRenameChat: ${nextTitle}`);
    await renameChat(activeChatId, nextTitle);
  }, [activeChatId, chats, renameChat]);

  const chatTitle = chats.find((chat) => chat.id === activeChatId)?.title ?? 'Untitled Chat';

  return (
    <div className="flex h-screen w-full select-none overflow-hidden bg-bg-0 text-text-primary">
      <Sidebar
        activeChatId={activeChatId ?? ''}
        chats={chats}
        isLoading={isLoading}
        onNewChat={handleNewChat}
        onSelectChat={handleSwitchChat}
        onDeleteChat={handleDeleteChat}
        onSettingsOpen={() => setIsSettingsOpen(true)}
      />

      <main className="relative flex h-full flex-1 flex-col overflow-hidden">
        <Titlebar chatTitle={chatTitle} onRename={handleRenameChat} />

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
