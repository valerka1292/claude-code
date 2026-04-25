/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { GlobTool } from "./GlobTool/GlobTool";
import type { Tool, ToolDefinition } from "./types";

export const TOOLS: Tool[] = [GlobTool];

export function getToolDefinitions(): ToolDefinition[] {
  return TOOLS.map((tool) => tool.definition);
}

export function getTool(name: string): Tool | undefined {
  return TOOLS.find((tool) => tool.metadata.name === name);
}

export function getReadOnlyTools(): Tool[] {
  return TOOLS.filter((tool) => tool.metadata.isReadOnly);
}

export function isReadOnlyTool(name: string): boolean {
  const tool = getTool(name);
  return tool?.metadata.isReadOnly ?? false;
}

export * from "./types";
export { GlobTool };
