/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
  type ReactNode,
} from "react";
import type { SessionData, SessionMeta } from "../lib/sessions";
import {
  listSessions,
  loadSession,
  saveSession,
  createNewSession,
  deleteSession,
} from "../lib/sessions";
import { useProject } from "./ProjectContext";

interface SessionContextValue {
  activeSession: SessionData | null;
  sessionList: SessionMeta[];
  isLoadingList: boolean;
  startNewSession: () => void;
  openSession: (id: string) => Promise<void>;
  updateSession: (updated: SessionData) => void;
  initSession: (projectPath: string) => SessionData;
  reloadList: () => Promise<void>;
  removeSession: (id: string) => Promise<void>;
}

const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const { projectKey } = useProject();

  const [activeSession, setActiveSession] = useState<SessionData | null>(null);
  const [sessionList, setSessionList] = useState<SessionMeta[]>([]);
  const [isLoadingList, setIsLoadingList] = useState(false);

  const activeSessionRef = useRef<SessionData | null>(null);
  activeSessionRef.current = activeSession;

  const projectKeyRef = useRef<string | null>(null);
  projectKeyRef.current = projectKey;

  useEffect(() => {
    setActiveSession(null);
    setSessionList([]);
    if (!projectKey) return;

    setIsLoadingList(true);
    listSessions(projectKey)
      .then(setSessionList)
      .finally(() => setIsLoadingList(false));
  }, [projectKey]);

  const reloadList = useCallback(async () => {
    const key = projectKeyRef.current;
    if (!key) return;
    const list = await listSessions(key);
    setSessionList(list);
  }, []);

  const startNewSession = useCallback(() => {
    const current = activeSessionRef.current;
    const key = projectKeyRef.current;

    if (current && key && current.messages.length > 0) {
      saveSession(key, current).catch(console.error);
    }

    setActiveSession(null);
  }, []);

  const openSession = useCallback(async (id: string) => {
    const key = projectKeyRef.current;
    if (!key) return;
    const session = await loadSession(key, id);
    if (session) setActiveSession(session);
  }, []);

  const initSession = useCallback((projectPath: string): SessionData => {
    const session = createNewSession(projectPath);
    setActiveSession(session);
    return session;
  }, []);

  const updateSession = useCallback((updated: SessionData) => {
    activeSessionRef.current = updated;
    setActiveSession(updated);
    setSessionList((prev) => {
      const meta: SessionMeta = {
        id: updated.id,
        name: updated.name,
        createdAt: updated.createdAt,
        projectPath: updated.projectPath,
      };
      const idx = prev.findIndex((s) => s.id === updated.id);
      if (idx >= 0) {
        const next = [...prev];
        next[idx] = meta;
        return next;
      }
      return [meta, ...prev];
    });
  }, []);

  const removeSession = useCallback(
    async (id: string) => {
      const key = projectKeyRef.current;
      if (!key) return;
      await deleteSession(key, id);
      setSessionList((prev) => prev.filter((s) => s.id !== id));
      if (activeSessionRef.current?.id === id) setActiveSession(null);
    },
    []
  );

  return (
    <SessionContext.Provider
      value={{
        activeSession,
        sessionList,
        isLoadingList,
        startNewSession,
        openSession,
        updateSession,
        initSession,
        reloadList,
        removeSession,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
}

export function useSession() {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used inside SessionProvider");
  return ctx;
}