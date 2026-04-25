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
}

export interface ToolCallDisplay {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  status?: "pending" | "running" | "success" | "error";
}
