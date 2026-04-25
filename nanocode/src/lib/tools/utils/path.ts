/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 *
 * Path utilities - ported from Claude Code
 */

import path from "path";
import { promises as fs } from "fs";

export function expandPath(inputPath: string, cwd: string): string {
  const trimmed = inputPath.trim();
  const homeDir = process.env.HOME;

  let expanded = trimmed;
  if (trimmed === "~" && homeDir) {
    expanded = homeDir;
  } else if (trimmed.startsWith("~/") && homeDir) {
    expanded = path.join(homeDir, trimmed.slice(2));
  }

  if (path.isAbsolute(expanded)) {
    return path.normalize(expanded);
  }

  return path.resolve(cwd, expanded);
}

export function toRelativePath(absolutePath: string, cwd: string): string {
  const relativePath = path.relative(cwd, absolutePath);

  if (
    relativePath.length > 0 &&
    !relativePath.startsWith("..") &&
    !path.isAbsolute(relativePath)
  ) {
    return relativePath;
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
  const basename = path.basename(absolutePath);
  if (!basename) {
    return null;
  }

  const candidate = path.join(cwd, basename);

  try {
    await fs.access(candidate);
    return `./${basename}`;
  } catch {
    return null;
  }
}
