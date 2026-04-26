/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  useState,
  useRef,
  useEffect,
  type KeyboardEvent,
  forwardRef,
  useImperativeHandle,
} from "react";
import { FolderIcon } from "./Icons";
import { ModeDropdown } from "./ModeDropdown";
import type { Mode } from "../types";
import { ContextIndicator } from "./ContextIndicator";
import { CornerDownLeft, Square } from "lucide-react";
import { useProject } from "../contexts/ProjectContext";

interface InputContainerProps {
  mode: Mode;
  setMode: (m: Mode) => void;
  onSend: (value: string) => void;
  onStop?: () => void;
  isGenerating?: boolean;
  usedTokens?: number;
}

export interface InputContainerHandle {
  setValue: (text: string) => void;
}

export const InputContainer = forwardRef<
  InputContainerHandle,
  InputContainerProps
>(function InputContainer(
  { mode, setMode, onSend, onStop, isGenerating = false, usedTokens = 0 },
  ref
) {
  const [inputValue, setInputValue] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { folderPath, folderName: _folderName, selectFolder } = useProject();

  const canSend = !!folderPath && !!inputValue.trim() && !isGenerating;

  useImperativeHandle(ref, () => ({
    setValue: (text: string) => {
      setInputValue(text);
      setTimeout(() => textareaRef.current?.focus(), 0);
    },
  }));

  useEffect(() => {
    const ta = textareaRef.current;
    if (!ta) return;
    ta.style.height = "auto";
    ta.style.height = Math.min(ta.scrollHeight, 140) + "px";
  }, [inputValue]);

  const handleSend = () => {
    if (!canSend) return;
    onSend(inputValue);
    setInputValue("");
    if (textareaRef.current) textareaRef.current.style.height = "auto";
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const folderLabel = folderPath ?? "no folder selected";

  return (
    <div className="px-5 pb-5 pt-0 w-full flex flex-col items-center">
      <div
        className="
          w-full max-w-[760px]
          rounded-xl overflow-hidden
          bg-[#121212]
          border border-white/[0.07]
          shadow-[0_8px_32px_rgba(0,0,0,0.6)]
          focus-within:border-white/[0.14]
          transition-all duration-200
        "
      >
        <div className="flex items-center gap-2 px-3 py-1.5 border-b border-white/[0.04]">
          <button
            onClick={selectFolder}
            title={folderPath ? folderPath : "Select project folder"}
            className="
              flex items-center gap-1.5 px-2 py-0.5 rounded-md
              text-[0.65rem] font-mono text-white/25
              hover:text-white/50 hover:bg-white/[0.04]
              transition-all duration-150
            "
          >
            <FolderIcon />
            <span className="truncate max-w-[340px]">{folderLabel}</span>
          </button>
        </div>

        <div className="flex items-start gap-2 px-3 pt-2.5 pb-2">
          <span className="text-[0.8rem] font-mono text-white/20 mt-0.5 flex-shrink-0 select-none">
            &gt;
          </span>
          <textarea
            ref={textareaRef}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={!folderPath}
            placeholder={
              folderPath ? "describe a task..." : "select a project folder first..."
            }
            rows={1}
            className="
              flex-1 resize-none bg-transparent
              text-[0.88rem] font-mono text-white/80
              placeholder-white/15
              outline-none leading-relaxed
              min-h-[26px] max-h-[140px]
              disabled:cursor-not-allowed
            "
          />
          <div className="relative group/send self-start mt-0.5">
            {isGenerating ? (
              <button
                onClick={onStop}
                className="
                  flex-shrink-0 flex items-center gap-1
                  px-2 py-1 rounded-md
                  text-[0.65rem] font-mono text-white/40
                  hover:text-white/80 hover:bg-white/[0.08]
                  border border-white/[0.08]
                  transition-all duration-150
                "
              >
                <Square size={11} />
              </button>
            ) : (
              <button
                onClick={handleSend}
                disabled={!canSend}
                className="
                  flex-shrink-0 flex items-center gap-1
                  px-2 py-1 rounded-md
                  text-[0.65rem] font-mono text-white/25
                  hover:text-white/60 hover:bg-white/[0.06]
                  disabled:opacity-20 disabled:cursor-not-allowed
                  border border-transparent hover:border-white/[0.08]
                  disabled:hover:bg-transparent disabled:hover:border-transparent
                  transition-all duration-150
                "
              >
                <CornerDownLeft size={11} />
              </button>
            )}
            {!folderPath && (
              <div
                className="
                  pointer-events-none absolute bottom-full right-0 mb-1.5
                  px-2 py-1 rounded-md
                  bg-[#1e1e1e] border border-white/[0.08]
                  text-[0.65rem] font-mono text-white/50
                  whitespace-nowrap
                  opacity-0 group-hover/send:opacity-100
                  transition-opacity duration-150
                  z-50
                "
              >
                Select a project folder first
              </div>
            )}
          </div>
        </div>

        <div className="flex items-center justify-between px-3 py-1.5 border-t border-white/[0.04] bg-black/10">
          <ModeDropdown mode={mode} setMode={setMode} />
          <ContextIndicator usedTokens={usedTokens} />
        </div>
      </div>
    </div>
  );
});
