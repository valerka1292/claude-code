/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from "react";

interface ProjectContextValue {
  folderPath: string | null;
  folderName: string | null;
  projectKey: string | null;
  selectFolder: () => Promise<void>;
  clearFolder: () => void;
}

const ProjectContext = createContext<ProjectContextValue | null>(null);

export function pathToProjectKey(absPath: string): string {
  return absPath
    .replace(/\\/g, "/")
    .replace(/^\//, "")
    .replace(/\/$/, "")
    .replace(/\//g, "-");
}

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [folderPath, setFolderPath] = useState<string | null>(null);

  const folderName = folderPath
    ? folderPath.replace(/\\/g, "/").split("/").filter(Boolean).at(-1) ?? null
    : null;

  const projectKey = folderPath ? pathToProjectKey(folderPath) : null;

  const selectFolder = useCallback(async () => {
    try {
      const result = await window.electronAPI!.selectFolder!();
      if (result) setFolderPath(result);
    } catch (err) {
      console.error("[ProjectContext] selectFolder error:", err);
    }
  }, []);

  const clearFolder = useCallback(() => {
    setFolderPath(null);
  }, []);

  return (
    <ProjectContext.Provider
      value={{ folderPath, folderName, projectKey, selectFolder, clearFolder }}
    >
      {children}
    </ProjectContext.Provider>
  );
}

export function useProject() {
  const ctx = useContext(ProjectContext);
  if (!ctx) throw new Error("useProject must be used inside ProjectProvider");
  return ctx;
}