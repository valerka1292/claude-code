/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Minus, Square, X, Edit3 } from 'lucide-react';
import Sidebar from './components/Sidebar';
import MessageList from './components/MessageList';
import InputArea from './components/InputArea';
import SettingsModal from './components/SettingsModal';
import Titlebar from './components/Titlebar';
import { AgentMode, Message } from './types';

export default function App() {
  const [activeChatId, setActiveChatId] = React.useState('1');
  const [mode, setMode] = React.useState<AgentMode>('Chat');
  const [messagesByChatId, setMessagesByChatId] = React.useState<Record<string, Message[]>>({});
  const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);
  const [isAgentRunning, setIsAgentRunning] = React.useState(false);
  const [isTyping, setIsTyping] = React.useState(false);

  // Simulated initial conversation
  React.useEffect(() => {
    setMessagesByChatId({
      '1': [
        {
          id: 'm1',
          role: 'assistant',
          content: 'Hello! I am your autonomous AI agent. I can help you write code, manage files, or even run investigative loops in autonomous mode.\n\nChoose a mode below to get started.',
          timestamp: new Date()
        }
      ]
    });
  }, []);

  const handleSend = (content: string) => {
    const chatId = activeChatId;
    const userMsg: Message = { id: Date.now().toString(), role: 'user', content, timestamp: new Date() };
    setMessagesByChatId((prev) => ({
      ...prev,
      [chatId]: [...(prev[chatId] ?? []), userMsg]
    }));
    setIsTyping(true);
    
    // Simulate assistant reply
    setTimeout(() => {
      const assistantMsg: Message = { 
        id: (Date.now() + 1).toString(), 
        role: 'assistant', 
        content: `I received your message: "${content}". How else can I help?`, 
        timestamp: new Date() 
      };
      setMessagesByChatId((prev) => ({
        ...prev,
        [chatId]: [...(prev[chatId] ?? []), assistantMsg]
      }));
      setIsTyping(false);
    }, 1500);
  };

  const handleNewChat = () => {
    const newChatId = Date.now().toString();
    setActiveChatId(newChatId);
    setMessagesByChatId((prev) => ({
      ...prev,
      [newChatId]: []
    }));
  };

  return (
    <div className="flex h-screen w-full bg-bg-0 overflow-hidden text-text-primary select-none">
      <Sidebar 
        activeChatId={activeChatId} 
        onNewChat={handleNewChat} 
        onSelectChat={setActiveChatId}
        onSettingsOpen={() => setIsSettingsOpen(true)}
      />

      <main className="flex-1 flex flex-col h-full overflow-hidden relative">
        {/* Electron Style Titlebar */}
        <Titlebar chatTitle="Untitled Chat" onRename={() => {}} />

        {/* Content area */}
        <MessageList messages={messagesByChatId[activeChatId] ?? []} isTyping={isTyping} />

        {/* Bottom Input */}
        <InputArea 
          mode={mode} 
          onModeChange={setMode} 
          onSend={handleSend}
          isAgentRunning={isAgentRunning}
          onToggleAgent={() => setIsAgentRunning(!isAgentRunning)}
        />

        {/* Modals */}
        <SettingsModal isOpen={isSettingsOpen} onClose={() => setIsSettingsOpen(false)} />
      </main>
    </div>
  );
}
