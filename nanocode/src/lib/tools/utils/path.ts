/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export function expandPath(inputPath: string, cwd: string): string {
  const trimmed = inputPath.trim();
  const homeDir = process.env.HOME || process.env.USERPROFILE;

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
  const basename = absolutePath.split(/[\\/]/).pop();
  if (!basename) {
    return null;
  }

  const candidate = `${cwd}/${basename}`;

  const electronApi =
    typeof window !== "undefined" ? window.electronAPI : undefined;

  if (!electronApi?.stat) {
    return null;
  }

  try {
    await electronApi.stat(candidate);
    return `./${basename}`;
  } catch {
    return null;
  }
}
