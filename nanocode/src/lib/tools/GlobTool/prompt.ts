/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 *
 * Glob tool prompts and constants - exact copy from Claude Code
 */

export const GLOB_TOOL_NAME = "Glob";

export const DESCRIPTION = `- Fast file pattern matching tool that works with any codebase size
- Supports glob patterns like "**/*.js" or "src/**/*.ts"
- Returns matching file paths sorted by modification time
- Use this tool when you need to find files by name patterns
- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead`;

export const SEARCH_HINT = "find files by name pattern or wildcard";

export const MAX_RESULTS = 100;
export const MAX_RESULT_SIZE_CHARS = 100_000;
