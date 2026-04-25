/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export const TOOL_SUMMARY_MAX_LENGTH = 50;

export function truncate(str: string, maxLength: number): string {
  if (str.length <= maxLength) {
    return str;
  }

  return `${str.slice(0, Math.max(0, maxLength - 3))}...`;
}

export function formatToolError(message: string): string {
  return `<tool_use_error>${message}</tool_use_error>`;
}

export function isToolError(result: string): boolean {
  return result.includes("<tool_use_error>") && result.includes("</tool_use_error>");
}

export function extractTag(content: string, tagName: string): string | null {
  const regex = new RegExp(`<${tagName}>([\\s\\S]*?)</${tagName}>`);
  const match = content.match(regex);
  return match?.[1] ?? null;
}
