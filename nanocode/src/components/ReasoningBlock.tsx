/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { ChevronDown, Brain } from "lucide-react";

interface ReasoningBlockProps {
  content: string;
  isStreaming?: boolean;
}

export function ReasoningBlock({ content, isStreaming = false }: ReasoningBlockProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="mb-2">
      <button
        onClick={() => setExpanded((v) => !v)}
        className="
          flex items-center gap-2 px-2 py-1 rounded-lg
          text-[0.67rem] font-mono text-white/25
          hover:text-white/50 hover:bg-white/[0.04]
          transition-all duration-150
        "
      >
        <Brain
          size={11}
          className={
            isStreaming ? "animate-pulse text-violet-400/60" : "text-white/20"
          }
        />
        <span>{isStreaming ? "Thinking..." : "Reasoning"}</span>
        <ChevronDown
          size={10}
          className={`opacity-50 transition-transform duration-200 ${expanded ? "rotate-180" : ""}`}
        />
      </button>

      <AnimatePresence>
        {expanded && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.18 }}
            className="overflow-hidden"
          >
            <div
              className="
                mt-1 ml-2 pl-3 border-l border-violet-500/20
                text-[0.75rem] text-white/30 font-mono leading-relaxed
                whitespace-pre-wrap break-words max-h-60 overflow-y-auto
              "
            >
              {content}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}