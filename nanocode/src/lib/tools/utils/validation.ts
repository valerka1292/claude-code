/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ValidationResult } from "../types";
import { importNodeModule } from "./node";
import { expandPath, suggestPathUnderCwd } from "./path";

export const FILE_NOT_FOUND_CWD_NOTE =
  "Note: paths are relative to the current working directory:";

export function isENOENT(error: unknown): boolean {
  return (
    error instanceof Error &&
    "code" in error &&
    (error as NodeJS.ErrnoException).code === "ENOENT"
  );
}

export async function validateDirectoryPath(
  inputPath: string | undefined,
  cwd: string
): Promise<ValidationResult> {
  if (!inputPath) {
    return { result: true };
  }

  const trimmed = inputPath.trim();
  if (trimmed.startsWith("\\\\") || trimmed.startsWith("//")) {
    return { result: true };
  }

  const absolutePath = expandPath(trimmed, cwd);
  const fsModule = await importNodeModule<typeof import("node:fs/promises")>("node:fs/promises");

  try {
    const stats = await fsModule.stat(absolutePath);

    if (!stats.isDirectory()) {
      return {
        result: false,
        errorCode: 2,
        message: `Path is not a directory: ${inputPath}`,
      };
    }

    return { result: true };
  } catch (error) {
    if (isENOENT(error)) {
      const suggestion = await suggestPathUnderCwd(absolutePath, cwd);
      const suggestionText = suggestion
        ? `\nDid you mean: ${suggestion}`
        : `\n${FILE_NOT_FOUND_CWD_NOTE} ${cwd}`;

      return {
        result: false,
        errorCode: 1,
        message: `Directory not found: ${inputPath}${suggestionText}`,
      };
    }

    const message = error instanceof Error ? error.message : String(error);
    return {
      result: false,
      message: `Failed to validate directory: ${message}`,
    };
  }
}
