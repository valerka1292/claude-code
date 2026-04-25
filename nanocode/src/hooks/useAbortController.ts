import { useCallback, useRef } from "react";

export function useAbortController() {
  const abortRef = useRef<AbortController | null>(null);

  const replaceActiveController = useCallback(() => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    return controller;
  }, []);

  const abortActiveRequest = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  const resetAbortController = useCallback(() => {
    abortRef.current?.abort();
    abortRef.current = null;
  }, []);

  return {
    replaceActiveController,
    abortActiveRequest,
    resetAbortController,
  };
}
