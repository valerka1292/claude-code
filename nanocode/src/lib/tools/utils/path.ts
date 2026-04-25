/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 *
 * Path utilities - ported from Claude Code
 */

import { importNodeModule } from "./node";

export function expandPath(inputPath: string, cwd: string): string {
  const trimmed = inputPath.trim();
  const homeDir = process.env.HOME;

  let expanded = trimmed;
  if (trimmed === "~" && homeDir) {
    expanded = homeDir;
  } else if (trimmed.startsWith("~/") && homeDir) {
    expanded = `${homeDir}/${trimmed.slice(2)}`;
  }

  if (expanded.startsWith("/") || /^[A-Za-z]:[\\/]/.test(expanded)) {
    return expanded;
  }

  const normalizedCwd = cwd.replace(/[\\/]+$/, "");
  return `${normalizedCwd}/${expanded}`;
}

export function toRelativePath(absolutePath: string, cwd: string): string {
  const normalize = (value: string): string => value.replace(/\\/g, "/").replace(/\/+$/, "");

  const base = normalize(cwd);
  const target = normalize(absolutePath);

  if (target === base) {
    return target;
  }

  if (target.startsWith(`${base}/`)) {
    return target.slice(base.length + 1);
  }

  return absolutePath;
}

export function getDisplayPath(inputPath: string, cwd: string): string {
  const absolutePath = expandPath(inputPath, cwd);
  return toRelativePath(absolutePath, cwd);
}

export async function suggestPathUnderCwd(
  absolutePath: string,
  cwd: string
): Promise<string | null> {
  const pathModule = await importNodeModule<typeof import("node:path")>("node:path");
  const fsModule = await importNodeModule<typeof import("node:fs/promises")>("node:fs/promises");

  const basename = pathModule.basename(absolutePath);
  if (!basename) {
    return null;
  }

  const candidate = pathModule.join(cwd, basename);

  try {
    await fsModule.access(candidate);
    return `./${basename}`;
  } catch {
    return null;
  }
}
