export interface Message {
  id: string;
  role: "user" | "assistant" | "tool";
  content: string;
  reasoning?: string;
  isStreaming?: boolean;
  isReasoningStreaming?: boolean;
  toolCalls?: ToolCallDisplay[];
  toolCallId?: string;
  toolName?: string;
  blocks?: ContentBlock[];
}

export interface ToolCallDisplay {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  status: "pending" | "running" | "success" | "error";
  result?: string;
}

export type ContentBlock =
  | { type: "reasoning"; content: string; streaming?: boolean }
  | { type: "text"; content: string; streaming?: boolean }
  | { type: "tool_call"; call: ToolCallDisplay }
  | { type: "tool_result"; callId: string; status: "running" | "success" | "error"; result?: string };