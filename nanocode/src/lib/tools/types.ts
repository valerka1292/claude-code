/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export interface ToolDefinition {
  type: "function";
  function: {
    name: string;
    description: string;
    parameters: {
      type: "object";
      properties: Record<string, unknown>;
      required?: string[];
    };
  };
}

export interface ToolMetadata {
  name: string;
  isReadOnly: boolean;
  isConcurrencySafe: boolean;
  maxResultSizeChars: number;
  searchHint?: string;
}

export interface ToolExecutor {
  (input: Record<string, unknown>, context: ToolExecutionContext): Promise<string>;
}

export interface ToolExecutionContext {
  cwd: string;
  signal?: AbortSignal;
}

export interface ValidationResult {
  result: boolean;
  message?: string;
  errorCode?: number;
}

export interface Tool {
  definition: ToolDefinition;
  metadata: ToolMetadata;
  execute: ToolExecutor;
  getUseSummary: (input: Record<string, unknown>) => string | null;
  getUseMessage: (
    input: Record<string, unknown>,
    cwd: string,
    verbose?: boolean
  ) => string;
  getActivityDescription: (input: Record<string, unknown>) => string;
}
