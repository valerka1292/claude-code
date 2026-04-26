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
import type { SessionData, SessionMeta } from "../types/session";
import {
  createNewSession,
  sessionRepository,
} from "../lib/sessions/index";
import { useProject } from "./ProjectContext";
import { startNewSessionWithGuard } from "./sessionStartGuard";

interface SessionContextValue {
  activeSession: SessionData | null;
  sessionList: SessionMeta[];
  isLoadingList: boolean;
  isTurnActive: boolean;
  setTurnActive: (isActive: boolean) => void;
  startNewSession: () => Promise<boolean>;
  openSession: (id: string) => Promise<void>;
  updateSession: (updated: SessionData) => void;
  initSession: (projectPath: string) => SessionData;
  reloadList: () => Promise<void>;
  removeSession: (id: string) => Promise<void>;
  getActiveSessionSnapshot: () => SessionData | null;
  onSessionSaveError: (error: unknown) => void;
}

const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const { projectKey } = useProject();

  const [activeSession, setActiveSession] = useState<SessionData | null>(null);
  const [sessionList, setSessionList] = useState<SessionMeta[]>([]);
  const [isLoadingList, setIsLoadingList] = useState(false);
  const [isTurnActive, setIsTurnActive] = useState(false);

  const activeSessionRef = useRef<SessionData | null>(null);
  activeSessionRef.current = activeSession;

  const projectKeyRef = useRef<string | null>(null);
  projectKeyRef.current = projectKey;
  const isTurnActiveRef = useRef(false);
  isTurnActiveRef.current = isTurnActive;
  const isMountedRef = useRef(true);

  const setTurnActive = useCallback((isActive: boolean) => {
    isTurnActiveRef.current = isActive;
    if (isMountedRef.current) {
      setIsTurnActive(isActive);
    }
  }, []);

  const onSessionSaveError = useCallback((error: unknown) => {
    console.error("[sessions] save error:", error);
    // TODO: surface this through non-blocking UI notification.
  }, []);

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    setActiveSession(null);
    setSessionList([]);
    if (!projectKey) return;

    setIsLoadingList(true);
    sessionRepository
      .list(projectKey)
      .then((list) => {
        if (!cancelled && isMountedRef.current) {
          setSessionList(list);
        }
      })
      .finally(() => {
        if (!cancelled && isMountedRef.current) {
          setIsLoadingList(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [projectKey]);

  const reloadList = useCallback(async () => {
    const key = projectKeyRef.current;
    if (!key) return;
    const list = await sessionRepository.list(key);
    if (isMountedRef.current) {
      setSessionList(list);
    }
  }, []);

  const startNewSession = useCallback(
    async () =>
      startNewSessionWithGuard({
        isTurnActive: isTurnActiveRef.current,
        currentSession: activeSessionRef.current,
        projectKey: projectKeyRef.current,
        clearActiveSession: () => {
          if (isMountedRef.current) {
            setActiveSession(null);
          }
        },
        saveSession: sessionRepository.save,
        onSessionSaveError,
      }),
    [onSessionSaveError]
  );

  const openSession = useCallback(async (id: string) => {
    const key = projectKeyRef.current;
    if (!key) return;
    const session = await sessionRepository.load(key, id);
    if (session && isMountedRef.current) {
      setActiveSession(session);
    }
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
      await sessionRepository.delete(key, id);
      if (!isMountedRef.current) return;
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
        isTurnActive,
        setTurnActive,
        startNewSession,
        openSession,
        updateSession,
        initSession,
        reloadList,
        removeSession,
        getActiveSessionSnapshot: () => activeSessionRef.current,
        onSessionSaveError,
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
