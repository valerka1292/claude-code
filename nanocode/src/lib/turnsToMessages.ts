import type { Message, ContentBlock, ToolCallDisplay } from "../types/message";
import type { StoredMessage } from "../types/session";

export function turnsToMessages(storedMessages: StoredMessage[]): Message[] {
  const messages: Message[] = [];
  let i = 0;

  while (i < storedMessages.length) {
    const current = storedMessages[i];

    if (current.role === "user") {
      messages.push({
        id: current.id,
        role: "user",
        content: current.content ?? "",
      });
      i++;
      continue;
    }

    const turnBlocks: ContentBlock[] = [];
    const turnStartIndex = i;

    while (i < storedMessages.length && storedMessages[i].role !== "user") {
      const msg = storedMessages[i];

      if (msg.role === "assistant") {
        if (msg.reasoning) {
          turnBlocks.push({
            type: "reasoning",
            content: msg.reasoning,
          });
        }

        if (msg.tool_calls) {
          for (const tc of msg.tool_calls) {
            let parsedArgs: Record<string, unknown> = {};
            try {
              parsedArgs = JSON.parse(tc.function.arguments) as Record<string, unknown>;
            } catch {
              parsedArgs = { raw: tc.function.arguments };
            }

            const call: ToolCallDisplay = {
              id: tc.id,
              name: tc.function.name,
              arguments: parsedArgs,
              status: "success",
            };
            turnBlocks.push({ type: "tool_call", call });
          }
        }

        if (msg.content) {
          turnBlocks.push({
            type: "text",
            content: msg.content,
          });
        }
      } else if (msg.role === "tool") {
        const toolCallId = msg.tool_call_id;
        if (toolCallId) {
          turnBlocks.push({
            type: "tool_result",
            callId: toolCallId,
            status: msg.content?.includes("<tool_use_error>") ? "error" : "success",
            result: msg.content ?? undefined,
          });
        }
      }

      i++;
    }

    if (turnBlocks.length > 0) {
      const firstMsg = storedMessages[turnStartIndex];
      messages.push({
        id: firstMsg.id,
        role: "assistant",
        content: "",
        blocks: turnBlocks,
      });
    }
  }

  return messages;
}