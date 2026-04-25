import type { Message } from "../types/message";
import type { ChatMessage } from "./agentLoop";
import type { StoredMessage } from "../types/session";

function toDeterministicSuffix(msg: StoredMessage): string {
  const payload = [
    msg.ts,
    msg.role,
    msg.content ?? "",
    msg.reasoning ?? "",
    msg.tool_call_id ?? "",
    msg.name ?? "",
  ].join("|");

  let hash = 0;
  for (let i = 0; i < payload.length; i++) {
    hash = (hash * 31 + payload.charCodeAt(i)) | 0;
  }

  return Math.abs(hash).toString(36);
}

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
    id: `${msg.ts}-${msg.role}-${toDeterministicSuffix(msg)}`,
    role: msg.role as "user" | "assistant",
    content: msg.content ?? "",
    reasoning: msg.reasoning,
  };
}
