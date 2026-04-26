import React, { useCallback, useEffect, useState } from 'react';
import Sidebar from './components/Sidebar';
import MessageList from './components/MessageList';
import InputArea from './components/InputArea';
import SettingsModal from './components/SettingsModal';
import Titlebar from './components/Titlebar';
import { AgentMode, Message, ToolCall } from './types';
import { streamChatCompletion } from './services/llmService';
import { MOCK_PROVIDERS } from './constants';
import { useProviders } from './hooks/useProviders';

export default function App() {
  const [activeChatId] = useState('1');
  const [mode, setMode] = useState<AgentMode>('Chat');
  const [messages, setMessages] = useState<Message[]>([]);
  const [llmHistory, setLlmHistory] = useState<{ role: string; content: string }[]>([]);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isAgentRunning, setIsAgentRunning] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [contextTokensUsed, setContextTokensUsed] = useState(0);
  const { activeProvider } = useProviders();

  useEffect(() => {
    setMessages([
      {
        id: 'm1',
        role: 'assistant',
        content: 'Hello! I am your autonomous AI agent. How can I help?',
        timestamp: new Date(),
      },
    ]);
  }, []);

  const handleSend = useCallback(
    (content: string) => {
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

      const provider = activeProvider ?? MOCK_PROVIDERS[0];
      let finalAssistantContent = '';

      streamChatCompletion(provider, nextLlmHistory, {
        onContent: (text) => {
          finalAssistantContent += text;
          setMessages((prev) =>
            prev.map((m) => (m.id === assistantId ? { ...m, content: m.content + text } : m)),
          );
        },
        onThinking: (text) => {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === assistantId ? { ...m, thinking: (m.thinking ?? '') + text } : m,
            ),
          );
        },
        onToolCall: (toolCall) => {
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
        onDone: (usage) => {
          setIsTyping(false);
          if (usage?.total_tokens) {
            setContextTokensUsed(usage.total_tokens);
          }
          setLlmHistory((prev) => [...prev, { role: 'assistant', content: finalAssistantContent }]);
        },
        onError: (error) => {
          setIsTyping(false);
          const errorMessage = error instanceof Error ? error.message : String(error);
          setMessages((prev) =>
            prev.map((m) => (m.id === assistantId ? { ...m, content: `Error: ${errorMessage}` } : m)),
          );
        },
      });
    },
    [activeProvider, llmHistory],
  );

  const handleNewChat = () => {
    setMessages([]);
    setLlmHistory([]);
    setContextTokensUsed(0);
  };

  return (
    <div className="flex h-screen w-full bg-bg-0 overflow-hidden text-text-primary select-none">
      <Sidebar
        activeChatId={activeChatId}
        onNewChat={handleNewChat}
        onSelectChat={() => {}}
        onSettingsOpen={() => setIsSettingsOpen(true)}
      />

      <main className="flex-1 flex flex-col h-full overflow-hidden relative">
        <Titlebar chatTitle="Untitled Chat" onRename={() => {}} />

        <MessageList messages={messages} isTyping={isTyping} />

        <InputArea
          mode={mode}
          onModeChange={setMode}
          onSend={handleSend}
          isAgentRunning={isAgentRunning}
          onToggleAgent={() => setIsAgentRunning(!isAgentRunning)}
          activeModel={(activeProvider ?? MOCK_PROVIDERS[0]).model}
          contextTokensUsed={contextTokensUsed}
          contextWindow={(activeProvider ?? MOCK_PROVIDERS[0]).contextWindowSize}
        />

        <SettingsModal isOpen={isSettingsOpen} onClose={() => setIsSettingsOpen(false)} />
      </main>
    </div>
  );
}
