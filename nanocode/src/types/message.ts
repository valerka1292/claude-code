export interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  reasoning?: string;
  isStreaming?: boolean;
  isReasoningStreaming?: boolean;
}
