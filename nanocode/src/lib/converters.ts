import type { Message } from "../types/message";
import type { ChatMessage } from "./agentLoop";
import type { StoredMessage } from "../types/session";

export function storedToChat(msg: StoredMessage): ChatMessage {
  const m: ChatMessage = {
    role: msg.role,
    content: msg.content,
  };
  if (msg.tool_calls) m.tool_calls = msg.tool_calls as ChatMessage["tool_calls"];
  if (msg.tool_call_id) m.tool_call_id = msg.tool_call_id;
  if (msg.name) m.name = msg.name;
  return m;
}

export function storedToUiMessage(msg: StoredMessage): Message {
  return {
    id: `${msg.ts}-${msg.role}-${Math.random().toString(36).slice(2, 7)}`,
    role: msg.role as "user" | "assistant",
    content: msg.content ?? "",
    reasoning: msg.reasoning,
  };
}
