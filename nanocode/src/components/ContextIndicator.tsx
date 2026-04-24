/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useProviders } from "../contexts/ProvidersContext";

interface ContextIndicatorProps {
  usedTokens?: number;
}

export function ContextIndicator({ usedTokens = 0 }: ContextIndicatorProps) {
  const { activeProvider } = useProviders();

  const contextWindow = activeProvider?.contextWindow ?? 128000;
  const modelName     = activeProvider?.model ?? "no model";

  const raw        = (usedTokens / contextWindow) * 100;
  const percentage = Math.min(100, raw);
  const percentDisplay = percentage.toFixed(2);

  const radius       = 7;
  const circumference = 2 * Math.PI * radius;
  const offset       = circumference - (percentage / 100) * circumference;

  return (
    <div className="flex items-center gap-3">
      <div className="flex items-center gap-2">
        <div
          className="relative w-5 h-5 flex items-center justify-center group/ctx cursor-default"
          title={`Context: ${percentDisplay}% used (${usedTokens.toLocaleString()} / ${contextWindow.toLocaleString()} tokens)`}
        >
          <svg className="w-full h-full -rotate-90">
            <circle
              cx="10"
              cy="10"
              r={radius}
              stroke="white"
              strokeWidth="1.5"
              fill="transparent"
              strokeOpacity="0.1"
            />
            <circle
              cx="10"
              cy="10"
              r={radius}
              stroke="white"
              strokeWidth="1.5"
              fill="transparent"
              strokeDasharray={circumference}
              strokeDashoffset={offset}
              strokeOpacity="0.8"
              strokeLinecap="round"
            />
          </svg>
          <div
            className="
              pointer-events-none absolute bottom-full right-0 mb-1.5
              px-2 py-1 rounded-md
              bg-[#1e1e1e] border border-white/[0.08]
              text-[0.65rem] font-mono text-white/50
              whitespace-nowrap
              opacity-0 group-hover/ctx:opacity-100
              transition-opacity duration-150
              z-50
            "
          >
            {percentDisplay}% · {usedTokens.toLocaleString()} / {contextWindow.toLocaleString()} tokens
          </div>
        </div>

        <span
          className="text-[0.75rem] font-medium text-neutral-500 uppercase tracking-wider truncate max-w-[140px]"
          title={modelName}
        >
          {modelName}
        </span>
      </div>
    </div>
  );
}