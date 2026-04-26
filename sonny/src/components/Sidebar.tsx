import React from 'react';
import { Plus, MessageSquare, Trash2, Edit2, Settings } from 'lucide-react';
import { ChatSession } from '../types';
import { MOCK_CHATS } from '../constants';
import { cn } from '../lib/utils';

interface SidebarProps {
  activeChatId: string;
  onNewChat: () => void;
  onSettingsOpen: () => void;
}

export default function Sidebar({ activeChatId, onNewChat, onSettingsOpen }: SidebarProps) {
  const [chats] = React.useState<ChatSession[]>(MOCK_CHATS);

  const groupedChats = React.useMemo(() => {
    const now = new Date();
    const today = new Date(now.setHours(0,0,0,0));
    const yesterday = new Date(today.getTime() - 86400000);
    const week = new Date(today.getTime() - 7 * 86400000);
    
    return {
      today: chats.filter((c) => new Date(c.updatedAt) >= today),
      yesterday: chats.filter((c) => new Date(c.updatedAt) >= yesterday && new Date(c.updatedAt) < today),
      week: chats.filter((c) => new Date(c.updatedAt) >= week && new Date(c.updatedAt) < yesterday),
      older: chats.filter((c) => new Date(c.updatedAt) < week)
    };
  }, [chats]);

  const renderChatGroup = (title: string, groupChats: ChatSession[]) => {
    if (groupChats.length === 0) return null;
    return (
      <div className="mb-2">
        <div className="px-3 pt-4 pb-2 first:pt-0 text-xs font-medium text-text-secondary">
          {title}
        </div>
        {groupChats.map((chat) => (
          <button
            key={chat.id}
            onClick={() => {}}
            className={cn(
              "group relative flex items-center w-full text-left gap-3 px-3 py-2 rounded-lg text-[13px] transition-colors outline-none",
              "focus-visible:ring-2 focus-visible:ring-white/20",
              activeChatId === chat.id 
                ? 'bg-bg-3 text-text-primary' 
                : 'hover:bg-bg-2 text-text-secondary hover:text-text-primary'
            )}
          >
            <MessageSquare size={14} className="flex-shrink-0 opacity-60" />
            <span className="truncate flex-1 min-w-0">{chat.title}</span>
            
            {/* Actions strictly on hover */}
            <div className="hidden group-hover:flex items-center gap-0.5 ml-auto relative z-10">
              <button 
                onClick={(e) => { e.stopPropagation(); }}
                className="p-1.5 hover:bg-bg-3 rounded text-text-secondary hover:text-text-primary transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/20"
                aria-label="Rename chat"
              >
                <Edit2 size={14} />
              </button>
              <button 
                onClick={(e) => { e.stopPropagation(); }}
                className="p-1.5 hover:bg-red-500/10 rounded text-text-secondary hover:text-red-400 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/20"
                aria-label="Delete chat"
              >
                <Trash2 size={14} />
              </button>
            </div>
          </button>
        ))}
      </div>
    );
  };

  return (
    <aside className="w-[260px] bg-bg-1 flex flex-col h-full flex-shrink-0 border-r border-border">
      {/* Titlebar extension for sidebar */}
      <div className="h-11 flex-shrink-0 flex items-center px-4 border-b border-border titlebar-drag">
        <span className="text-[13px] font-medium text-text-secondary">Agent Workspace</span>
      </div>

      {/* New Chat Button */}
      <div className="p-3 border-b border-border flex-shrink-0">
        <button
          onClick={onNewChat}
          className="w-full flex items-center gap-2.5 py-2.5 px-3 bg-bg-2 hover:bg-bg-3 border border-border rounded-lg text-[13px] font-medium transition-colors no-drag text-text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/20"
        >
          <Plus size={16} strokeWidth={2} />
          New chat
        </button>
      </div>

      {/* History */}
      <div className="flex-1 overflow-y-auto px-2 py-4 flex flex-col gap-0.5">
        {renderChatGroup('Today', groupedChats.today)}
        {renderChatGroup('Yesterday', groupedChats.yesterday)}
        {renderChatGroup('Previous 7 Days', groupedChats.week)}
        {renderChatGroup('Older', groupedChats.older)}
      </div>

      {/* Bottom Actions */}
      <div className="p-3 border-t border-border flex flex-col gap-1 flex-shrink-0">
        <button
          onClick={onSettingsOpen}
          className="w-full flex items-center gap-3 py-2 px-3 hover:bg-bg-3 rounded-lg text-sm text-text-primary transition-colors no-drag focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/20"
        >
          <Settings size={16} className="text-text-secondary" />
          Settings
        </button>
        <div className="px-3 pt-1 pb-1 text-[10px] text-text-secondary">
          v0.0.1
        </div>
      </div>
    </aside>
  );
}
