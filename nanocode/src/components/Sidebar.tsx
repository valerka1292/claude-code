/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { Plus, Settings, Trash2 } from "lucide-react";
import { useProject } from "../contexts/ProjectContext";
import type { SessionMeta } from "../lib/sessions";

interface SidebarProps {
  onSettingsClick: () => void;
  onNewSession: () => void;
  onOpenSession: (id: string) => void;
  onDeleteSession: (id: string) => void;
  sessionList: SessionMeta[];
  activeSessionId: string | null;
  isLoadingList: boolean;
}

export function Sidebar({
  onSettingsClick,
  onNewSession,
  onOpenSession,
  onDeleteSession,
  sessionList,
  activeSessionId,
  isLoadingList,
}: SidebarProps) {
  const { folderPath } = useProject();
  const hasProject = !!folderPath;

  return (
    <aside className="w-[210px] min-w-[210px] h-full bg-[#0f0f0f] border-r border-white/[0.05] flex flex-col overflow-hidden">
      {hasProject && (
        <div className="px-3 py-3">
          <button
            onClick={onNewSession}
            className="
              w-full flex items-center gap-2 px-3 py-2
              rounded-lg bg-white/[0.04] border border-white/[0.06]
              text-[0.75rem] text-white/50 hover:text-white/80
              hover:bg-white/[0.07] hover:border-white/[0.1]
              transition-all duration-150 group
            "
          >
            <Plus size={13} className="opacity-50 group-hover:opacity-80" />
            New session
          </button>
        </div>
      )}

      {hasProject && <div className="mx-3 border-t border-white/[0.04]" />}

      <div className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
        {hasProject ? (
          <>
            <div className="px-2 py-1.5">
              <span className="text-[0.6rem] font-semibold text-white/20 uppercase tracking-[0.15em]">
                Recent
              </span>
            </div>

            {isLoadingList && (
              <div className="px-2 py-3 text-[0.65rem] text-white/15 font-mono">
                loading...
              </div>
            )}

            {!isLoadingList && sessionList.length === 0 && (
              <div className="px-2 py-3 text-[0.65rem] text-white/15 font-mono">
                no sessions yet
              </div>
            )}

            {sessionList.map((s) => (
              <SessionRow
                key={s.id}
                session={s}
                isActive={s.id === activeSessionId}
                onOpen={() => onOpenSession(s.id)}
                onDelete={() => onDeleteSession(s.id)}
              />
            ))}
          </>
        ) : (
          <div className="flex flex-col items-center justify-center h-full pb-8 gap-2">
            <span className="text-[0.65rem] font-mono text-white/15 text-center leading-relaxed px-4">
              Select a project<br />to see sessions
            </span>
          </div>
        )}
      </div>

      <div className="px-3 py-3 border-t border-white/[0.04]">
        <button
          onClick={onSettingsClick}
          className="
            w-full flex items-center gap-2 px-2.5 py-2 rounded-lg
            text-[0.72rem] text-white/30 hover:text-white/60
            hover:bg-white/[0.04] transition-all duration-150
          "
        >
          <Settings size={13} />
          Settings
        </button>
      </div>
    </aside>
  );
}

interface SessionRowProps {
  session: SessionMeta;
  isActive: boolean;
  onOpen: () => void;
  onDelete: () => void;
}

function SessionRow({ session, isActive, onOpen, onDelete }: SessionRowProps) {
  const [confirmDelete, setConfirmDelete] = useState(false);

  const timeLabel = formatRelativeTime(session.createdAt);

  return (
    <div
      className={`
        group/row w-full flex items-start gap-1
        px-2.5 py-2 rounded-lg transition-all duration-100 cursor-pointer
        ${
          isActive
            ? "bg-white/[0.07] text-white/90"
            : "text-white/40 hover:bg-white/[0.04] hover:text-white/70"
        }
      `}
      onClick={() => !confirmDelete && onOpen()}
    >
      <div className="flex-1 min-w-0">
        <div className="text-[0.75rem] font-medium truncate leading-snug">
          {session.name}
        </div>
        <div className="text-[0.62rem] text-white/20">{timeLabel}</div>
      </div>

      {confirmDelete ? (
        <div
          className="flex items-center gap-1 flex-shrink-0"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => onDelete()}
            className="text-[0.6rem] text-red-400/80 hover:text-red-400 px-1"
          >
            del
          </button>
          <button
            onClick={() => setConfirmDelete(false)}
            className="text-[0.6rem] text-white/25 hover:text-white/50 px-1"
          >
            ✕
          </button>
        </div>
      ) : (
        <button
          onClick={(e) => {
            e.stopPropagation();
            setConfirmDelete(true);
          }}
          className="
            flex-shrink-0 p-0.5 rounded
            text-white/0 group-hover/row:text-white/20
            hover:!text-white/50 hover:bg-white/[0.06]
            transition-all duration-100
          "
        >
          <Trash2 size={11} />
        </button>
      )}
    </div>
  );
}

function formatRelativeTime(ts: number): string {
  const diff = Date.now() - ts;
  const min = Math.floor(diff / 60_000);
  const hr = Math.floor(diff / 3_600_000);
  const day = Math.floor(diff / 86_400_000);

  if (min < 1) return "just now";
  if (min < 60) return `${min}m ago`;
  if (hr < 24) return `${hr}h ago`;
  if (day < 7) return `${day}d ago`;
  return new Date(ts).toLocaleDateString();
}