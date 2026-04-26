/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 *
 * Glob tool - Fast file pattern matching
 * Full port from Claude Code with all validations and security checks
 */

import type { Tool, ToolExecutionContext } from "../types";
import { formatToolError } from "../utils/format";
import { expandPath, toRelativePath } from "../utils/path";
import { validateDirectoryPath } from "../utils/validation";
import {
  DESCRIPTION,
  GLOB_TOOL_NAME,
  MAX_RESULTS,
  MAX_RESULT_SIZE_CHARS,
  SEARCH_HINT,
} from "./prompt";
import {
  getActivityDescription,
  getToolUseSummary,
  renderToolUseMessage,
} from "./UI";

interface GlobInput {
  pattern: string;
  path?: string;
}

interface GlobOutput {
  durationMs: number;
  numFiles: number;
  filenames: string[];
  truncated: boolean;
}

async function execute(
  input: Record<string, unknown>,
  context: ToolExecutionContext
): Promise<string> {
  try {
    const electronApi =
      typeof window !== "undefined" ? window.electronAPI : undefined;

    if (!electronApi?.glob || !electronApi?.stat) {
      return formatToolError("Glob tool is unavailable: electronAPI not initialized");
    }
    const { glob, stat } = electronApi;

    const pattern = typeof input.pattern === "string" ? input.pattern.trim() : "";
    const requestedPath = typeof input.path === "string" ? input.path : undefined;

    if (!pattern) {
      return formatToolError("pattern is required and must be a non-empty string");
    }

    const pathValidation = await validateDirectoryPath(requestedPath, context.cwd);
    if (!pathValidation.result) {
      return formatToolError(pathValidation.message ?? "Invalid directory path");
    }

    const searchDir = requestedPath
      ? expandPath(requestedPath, context.cwd)
      : context.cwd;

    const startedAt = Date.now();
    const files = await glob(pattern, {
      cwd: searchDir,
      absolute: true,
      nodir: true,
      dot: true,
      follow: false,
    });

    const filesWithMtime = await Promise.all(
      files.map(async (file) => {
        try {
          const stats = await stat(file);
          return { file, mtime: stats.mtimeMs };
        } catch {
          return { file, mtime: 0 };
        }
      })
    );

    filesWithMtime.sort((a, b) => b.mtime - a.mtime);

    const truncated = filesWithMtime.length > MAX_RESULTS;
    const limitedFiles = filesWithMtime.slice(0, MAX_RESULTS).map(({ file }) =>
      toRelativePath(file, context.cwd)
    );

    return formatResult({
      durationMs: Date.now() - startedAt,
      numFiles: filesWithMtime.length,
      filenames: limitedFiles,
      truncated,
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return formatToolError(message);
  }
}

function formatResult(output: GlobOutput): string {
  if (output.filenames.length === 0) {
    return "No files found";
  }

  const result = output.filenames.join("\n");

  if (!output.truncated) {
    return result;
  }

  return `${result}\n(Results are truncated. Consider using a more specific path or pattern.)`;
}

export const GlobTool: Tool = {
  definition: {
    type: "function",
    function: {
      name: GLOB_TOOL_NAME,
      description: DESCRIPTION,
      parameters: {
        type: "object",
        properties: {
          pattern: {
            type: "string",
            description: "The glob pattern to match files against",
          },
          path: {
            type: "string",
            description:
              'The directory to search in. If not specified, the current working directory will be used. IMPORTANT: Omit this field to use the default directory. DO NOT enter "undefined" or "null" - simply omit it for the default behavior. Must be a valid directory path if provided.',
          },
        },
        required: ["pattern"],
      },
    },
  },
  metadata: {
    name: GLOB_TOOL_NAME,
    isReadOnly: true,
    isConcurrencySafe: true,
    maxResultSizeChars: MAX_RESULT_SIZE_CHARS,
    searchHint: SEARCH_HINT,
  },
  execute,
  getUseSummary: (input) => getToolUseSummary(input as Partial<GlobInput>),
  getUseMessage: (input, cwd, verbose = false) =>
    renderToolUseMessage(input as Partial<GlobInput>, cwd, verbose),
  getActivityDescription: (input) =>
    getActivityDescription(input as unknown as GlobInput),
};
