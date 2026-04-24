/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { X } from "lucide-react";
import { ProvidersPanel } from "./ProvidersPanel";

interface SettingsModalProps {
  open: boolean;
  onClose: () => void;
}

type SettingsTab = "providers";

export function SettingsModal({ open, onClose }: SettingsModalProps) {
  const [tab, setTab] = useState<SettingsTab>("providers");

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            key="backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/60 backdrop-blur-[2px] z-40"
            onClick={onClose}
          />
          <motion.div
            key="modal"
            initial={{ opacity: 0, scale: 0.96, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: 8 }}
            transition={{ duration: 0.18 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4 pointer-events-none"
          >
            <div
              className="
                pointer-events-auto
                w-full max-w-2xl max-h-[80vh]
                bg-[#111] border border-white/[0.09]
                rounded-2xl shadow-[0_24px_64px_rgba(0,0,0,0.8)]
                flex flex-col overflow-hidden
              "
            >
              <div className="flex items-center justify-between px-5 py-4 border-b border-white/[0.06]">
                <div className="flex items-center gap-4">
                  <span className="text-[0.8rem] font-semibold text-white/70 tracking-wide">
                    Settings
                  </span>
                  <div className="flex gap-1">
                    {(["providers"] as SettingsTab[]).map((t) => (
                      <button
                        key={t}
                        onClick={() => setTab(t)}
                        className={`
                          px-3 py-1 rounded-md text-[0.7rem] font-mono capitalize
                          transition-all duration-150
                          ${tab === t
                            ? "bg-white/[0.1] text-white/90"
                            : "text-white/30 hover:text-white/60 hover:bg-white/[0.05]"
                          }
                        `}
                      >
                        {t}
                      </button>
                    ))}
                  </div>
                </div>
                <button
                  onClick={onClose}
                  className="
                    w-7 h-7 flex items-center justify-center rounded-lg
                    text-white/30 hover:text-white/70 hover:bg-white/[0.07]
                    transition-all duration-150
                  "
                >
                  <X size={14} />
                </button>
              </div>

              <div className="flex-1 overflow-y-auto">
                {tab === "providers" && <ProvidersPanel />}
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}