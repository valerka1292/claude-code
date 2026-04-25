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
  const base: Message = {
    id: msg.id,
    role: msg.role,
    content: msg.content ?? "",
    reasoning: msg.reasoning,
  };

  if (msg.role === "assistant" && msg.tool_calls) {
    base.toolCalls = msg.tool_calls.map((tc) => {
      let parsedArgs: Record<string, unknown> = {};
      try {
        parsedArgs = JSON.parse(tc.function.arguments) as Record<string, unknown>;
      } catch {
        parsedArgs = { raw: tc.function.arguments };
      }
      return {
        id: tc.id,
        name: tc.function.name,
        arguments: parsedArgs,
        status: "success",
      };
    });
  }

  if (msg.role === "tool") {
    base.toolCallId = msg.tool_call_id;
    base.toolName = msg.name;
  }

  return base;
}
