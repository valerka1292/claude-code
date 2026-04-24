/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { ChevronDown, MessageSquare, Code } from "lucide-react";

import type { Mode } from "../types";

interface ModeDropdownProps {
  mode: Mode;
  setMode: (m: Mode) => void;
}

export function ModeDropdown({ mode, setMode }: ModeDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const modes: Mode[] = ["Ask", "Code"];

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="
          flex items-center gap-1.5
          px-2 py-1 rounded-md
          bg-white/[0.03] border border-white/[0.05]
          text-[0.72rem] font-medium text-neutral-400
          hover:bg-white/[0.06] hover:text-neutral-200
          transition-all duration-150
        "
      >
        {mode === "Ask" ? (
          <MessageSquare size={12} className="opacity-70" />
        ) : (
          <Code size={12} className="opacity-70" />
        )}
        <span>{mode}</span>
        <ChevronDown 
          size={10} 
          className={`opacity-50 transition-transform duration-200 ${isOpen ? "rotate-180" : ""}`} 
        />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: 4, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 4, scale: 0.95 }}
            className="
              absolute bottom-full left-0 mb-1.5
              w-28 bg-[#1a1a1a] border border-[#2e2e2e]
              rounded-lg shadow-2xl p-1 z-50
            "
          >
            {modes.map((m) => (
              <button
                key={m}
                onClick={() => {
                  setMode(m);
                  setIsOpen(false);
                }}
                className={`
                  w-full flex items-center gap-2 px-2 py-1.5 rounded-md 
                  text-[0.72rem] transition-colors
                  ${mode === m 
                    ? "bg-white/[0.08] text-white" 
                    : "text-neutral-400 hover:bg-white/[0.04] hover:text-white/80"
                  }
                `}
              >
                {m === "Ask" ? <MessageSquare size={12} /> : <Code size={12} />}
                {m}
              </button>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
