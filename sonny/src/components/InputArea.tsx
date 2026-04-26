import React from 'react';
import { Send, ChevronDown, Play, Square, Info } from 'lucide-react';
import * as Select from '@radix-ui/react-select';
import { AgentMode } from '../types';
import { AGENT_MODES } from '../constants';
import { cn } from '../lib/utils';

interface InputAreaProps {
  mode: AgentMode;
  onModeChange: (mode: AgentMode) => void;
  onSend: (text: string) => void;
  isAgentRunning: boolean;
  onToggleAgent: () => void;
}

export default function InputArea({ mode, onModeChange, onSend, isAgentRunning, onToggleAgent }: InputAreaProps) {
  const [text, setText] = React.useState('');
  const textareaRef = React.useRef<HTMLTextAreaElement>(null);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      if (mode === 'Chat' && text.trim()) {
        e.preventDefault();
        onSend(text);
        setText('');
      }
    }
  };

  React.useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'inherit';
      const scrollHeight = textareaRef.current.scrollHeight;
      textareaRef.current.style.height = `${Math.min(scrollHeight, 200)}px`;
    }
  }, [text]);

  const contextUsed = 4821;
  const contextMax = 128000;
  const contextPercent = (contextUsed / contextMax) * 100;

  return (
    <div className="sticky bottom-0 left-0 right-0 bg-gradient-to-t from-bg-0 via-bg-0 to-transparent pt-4 pb-6">
      <div className="w-full max-w-3xl mx-auto px-4 flex flex-col gap-3">
        
        {/* Mode Selector Top Bar */}
        <div className="flex items-center justify-between mb-0 px-1">
          <Select.Root value={mode} onValueChange={(val) => onModeChange(val as AgentMode)}>
            <Select.Trigger 
              aria-label="Agent Mode" 
              className="flex items-center gap-2 px-3 py-1.5 bg-bg-2 hover:bg-bg-3 border border-border rounded-lg text-sm font-medium text-text-primary transition-colors outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
            >
              <span className="text-text-secondary text-xs font-normal">Mode:</span>
              <Select.Value />
              <Select.Icon>
                <ChevronDown size={14} className="text-text-secondary" />
              </Select.Icon>
            </Select.Trigger>
            <Select.Portal>
              <Select.Content position="popper" side="top" align="start" sideOffset={8} className="bg-bg-1 border border-border shadow-lg rounded-xl overflow-hidden z-50 w-64 animate-in fade-in-0 zoom-in-95">
                <Select.Viewport className="py-1">
                  {AGENT_MODES.map((m) => (
                    <Select.Item key={m.id} value={m.id} className="text-[13px] px-4 py-2 hover:bg-bg-3 cursor-pointer outline-none flex flex-col gap-0.5 select-none data-[highlighted]:bg-bg-3 data-[state=checked]:bg-bg-3">
                      <Select.ItemText>
                        <span className={`font-medium ${mode === m.id ? 'text-text-primary' : 'text-text-secondary'}`}>{m.label}</span>
                      </Select.ItemText>
                      <span className="text-[10px] text-text-secondary block truncate">{m.description}</span>
                    </Select.Item>
                  ))}
                </Select.Viewport>
              </Select.Content>
            </Select.Portal>
          </Select.Root>
        </div>

        {/* Input Field or Agent Controls */}
        <div className="relative bg-bg-1 rounded-xl border border-border focus-within:border-bg-3 focus-within:shadow-[0_0_0_3px_rgba(255,255,255,0.05)] transition-all">
          {mode === 'Chat' ? (
            <div className="flex items-end gap-2 px-4 pb-3">
              <textarea
                ref={textareaRef}
                value={text}
                onChange={(e) => setText(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Message agent..."
                aria-label="Message input"
                className="flex-1 bg-transparent resize-none outline-none text-[15px] pt-3 text-text-primary placeholder:text-text-secondary min-h-[52px] max-h-[200px] leading-relaxed scrollbar-thin"
                rows={1}
              />
              <button
                aria-label="Send message"
                onClick={() => {
                  if (text.trim()) {
                    onSend(text);
                    setText('');
                  }
                }}
                disabled={!text.trim()}
                className="w-9 h-9 flex-shrink-0 rounded-lg flex items-center justify-center transition-all bg-white text-black disabled:bg-bg-3 disabled:text-text-secondary hover:opacity-90 disabled:hover:opacity-100 disabled:cursor-not-allowed outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
              >
                <Send size={16} />
              </button>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-3 py-6">
              {/* Status indicator */}
              <div className="flex items-center gap-2 text-sm">
                <div className={cn(
                  "w-2 h-2 rounded-full transition-colors",
                  isAgentRunning ? "bg-green-500 animate-pulse" : "bg-bg-3"
                )} />
                <span className="text-text-secondary font-medium">
                  {isAgentRunning ? 'Running...' : 'Idle'}
                </span>
              </div>

              {/* Action button */}
              {!isAgentRunning ? (
                <button 
                  onClick={onToggleAgent}
                  className="flex items-center gap-2 px-6 py-2.5 bg-white text-black rounded-lg font-medium hover:bg-white/90 transition-all text-sm shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-white/20"
                >
                  <Play size={16} className="fill-current" />
                  Start {mode} Cycle
                </button>
              ) : (
                <button 
                  onClick={onToggleAgent}
                  className="flex items-center gap-2 px-6 py-2.5 bg-red-500 text-white rounded-lg font-medium hover:bg-red-600 transition-colors text-sm outline-none focus-visible:ring-2 focus-visible:ring-red-500/50"
                >
                  <Square size={14} className="fill-current" />
                  Stop
                </button>
              )}
            </div>
          )}
        </div>

        {/* Bottom Info Bar */}
        <div className="flex items-center justify-between text-xs text-text-secondary px-1 mt-1">
          {/* Left: model + context */}
          <div className="flex items-center gap-3">
            <span>gpt-4o</span>
            <span className="text-border">•</span>
            <div className="flex items-center gap-2">
              <div className="w-20 h-1.5 bg-bg-2 rounded-full overflow-hidden border border-border">
                <div 
                  className={cn(
                    "h-full rounded-full transition-all",
                    contextPercent < 50 ? "bg-green-500" :
                    contextPercent < 80 ? "bg-yellow-500" :
                    "bg-red-500"
                  )}
                  style={{ width: `${contextPercent}%` }}
                />
              </div>
              <span className="tabular-nums">{Math.round(contextPercent)}%</span>
            </div>
          </div>

          {/* Right: keyboard hint */}
          <div className="hidden sm:flex items-center justify-end gap-1.5 text-text-secondary/60">
            <kbd className="px-1.5 py-0.5 bg-bg-2 border border-border rounded text-[10px] font-mono">↵</kbd>
            <span>to send</span>
          </div>
        </div>

      </div>
    </div>
  );
}
