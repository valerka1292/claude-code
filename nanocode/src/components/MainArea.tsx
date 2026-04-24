/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { type ReactNode, type RefObject } from "react";
import { Terminal } from "lucide-react";
import { useProject } from "../contexts/ProjectContext";
import type { InputContainerHandle } from "./InputContainer";

interface MainAreaProps {
  children?: ReactNode;
  hasMessages?: boolean;
  input?: ReactNode;
  inputRef?: RefObject<InputContainerHandle>;
}

export function MainArea({ children, hasMessages, input, inputRef }: MainAreaProps) {
  const { folderName } = useProject();

  const displayName = folderName ?? "no project";

  return (
    <main className="flex flex-col flex-1 h-full overflow-hidden relative">

      <div className="
        h-10 flex-shrink-0 flex items-center justify-between
        px-5 border-b border-white/[0.05] bg-[#0d0d0d]
      ">
        <div className="flex items-center gap-3">
          <span className="text-[0.72rem] text-white/20 font-mono">
            {displayName}
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 rounded-full bg-emerald-400/60" />
          <span className="text-[0.65rem] text-white/25 font-mono">ready</span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto relative flex flex-col scroll-smooth min-h-0">
        {!hasMessages ? (
          <EmptyState inputRef={inputRef} />
        ) : (
          <div className="w-full max-w-[780px] mx-auto flex flex-col pb-2">
            {children}
          </div>
        )}
      </div>

      {input && (
        <div className="flex-shrink-0 bg-gradient-to-t from-[#0d0d0d] via-[#0d0d0d]/95 to-transparent pt-3">
          {input}
        </div>
      )}
    </main>
  );
}

interface EmptyStateProps {
  inputRef?: RefObject<InputContainerHandle>;
}

const QUICK_PROMPTS = [
  "Explain this codebase",
  "Fix TypeScript errors",
  "Write unit tests",
  "Refactor for readability",
];

function EmptyState({ inputRef }: EmptyStateProps) {
  const handleQuickPrompt = (prompt: string) => {
    inputRef?.current?.setValue(prompt);
  };

  return (
    <div className="flex-1 flex flex-col items-center justify-center gap-6 pb-16 px-8">

      <div className="flex flex-col items-center gap-3">
        <div className="
          w-10 h-10 rounded-xl bg-white/[0.05] border border-white/[0.08]
          flex items-center justify-center
        ">
          <Terminal size={18} className="text-white/40" />
        </div>
        <div className="text-center">
          <h1 className="text-[1.5rem] font-semibold text-white/80 leading-tight tracking-tight">
            What are we building?
          </h1>
          <p className="mt-1.5 text-[0.82rem] text-white/25">
            Describe a task, ask a question, or paste code
          </p>
        </div>
      </div>

      <div className="flex flex-wrap gap-2 justify-center max-w-sm">
        {QUICK_PROMPTS.map((prompt) => (
          <button
            key={prompt}
            onClick={() => handleQuickPrompt(prompt)}
            className="
              px-3 py-1.5 rounded-lg
              bg-white/[0.03] border border-white/[0.06]
              text-[0.72rem] text-white/35
              hover:bg-white/[0.06] hover:text-white/60 hover:border-white/[0.1]
              transition-all duration-150
            "
          >
            {prompt}
          </button>
        ))}
      </div>
    </div>
  );
}