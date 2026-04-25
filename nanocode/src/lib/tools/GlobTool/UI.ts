/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 *
 * Glob tool UI formatting - ported from Claude Code
 */

import { TOOL_SUMMARY_MAX_LENGTH, truncate } from "../utils/format";
import { expandPath, toRelativePath } from "../utils/path";

interface GlobInput {
  pattern: string;
  path?: string;
}

export function userFacingName(): string {
  return "Search";
}

export function getToolUseSummary(input: Partial<GlobInput>): string | null {
  if (!input.pattern) {
    return null;
  }

  return truncate(input.pattern, TOOL_SUMMARY_MAX_LENGTH);
}

export function renderToolUseMessage(
  input: Partial<GlobInput>,
  cwd: string,
  verbose: boolean
): string {
  if (!input.pattern) {
    return "";
  }

  if (!input.path) {
    return `pattern: "${input.pattern}"`;
  }

  const absolutePath = expandPath(input.path, cwd);
  const displayPath = verbose ? absolutePath : toRelativePath(absolutePath, cwd);

  return `pattern: "${input.pattern}", path: "${displayPath}"`;
}

export function getActivityDescription(input: GlobInput): string {
  const summary = getToolUseSummary(input);

  if (!summary) {
    return "Finding files";
  }

  return `Finding ${summary}`;
}
