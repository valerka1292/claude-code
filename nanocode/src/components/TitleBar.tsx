/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import type { CSSProperties } from "react";

export function TitleBar() {
  const isElectron = typeof window !== "undefined" && !!window.electronAPI;

  return (
    <div
      className="
        h-9 flex items-center justify-between
        px-3 bg-[#0a0a0a] border-b border-white/[0.04]
        select-none
      "
      style={{ WebkitAppRegion: "drag" } as CSSProperties}
    >
      <div className="flex items-center gap-2">
        <div className="w-3.5 h-3.5 bg-white rounded-[3px] flex items-center justify-center">
          <div className="w-1.5 h-1.5 bg-black rounded-full" />
        </div>
        <span className="text-[0.65rem] font-bold tracking-[0.25em] text-white/40 uppercase">
          NanoCode
        </span>
      </div>

      {isElectron && (
        <div
          className="flex items-center gap-1"
          style={{ WebkitAppRegion: "no-drag" } as CSSProperties}
        >
          {[
            { label: "−", action: "minimize", hover: "hover:bg-white/[0.08]" },
            { label: "□", action: "maximize", hover: "hover:bg-white/[0.08]" },
            { label: "×", action: "close",    hover: "hover:bg-red-500/60" },
          ].map(({ label, action, hover }) => (
            <button
              key={action}
              onClick={() => window.electronAPI?.[action]?.()}
              className={`
                w-7 h-6 flex items-center justify-center
                text-[0.75rem] text-white/30 hover:text-white/80
                rounded transition-colors duration-100 ${hover}
              `}
            >
              {label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}