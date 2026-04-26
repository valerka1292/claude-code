import type { Message, ContentBlock } from "../types/message";

export function applyToolExecutionResultToMessages(
  messages: Message[],
  assistantId: string,
  callId: string,
  status: "success" | "error",
  result: string
): Message[] {
  return messages.map((message) => {
    if (message.id !== assistantId) return message;

    const updatedToolCalls = message.toolCalls?.map((toolCall) =>
      toolCall.id === callId ? { ...toolCall, status, result } : toolCall
    );

    const blocks = message.blocks ?? [];
    const toolResultBlock: ContentBlock = {
      type: "tool_result",
      callId,
      status,
      result,
    };

    let existingBlockIndex = -1;
    for (let i = blocks.length - 1; i >= 0; i -= 1) {
      const block = blocks[i];
      if (block.type === "tool_result" && block.callId === callId) {
        existingBlockIndex = i;
        break;
      }
    }

    const updatedBlocks = [...blocks];
    if (existingBlockIndex >= 0) {
      updatedBlocks[existingBlockIndex] = toolResultBlock;
    } else {
      updatedBlocks.push(toolResultBlock);
    }

    return {
      ...message,
      toolCalls: updatedToolCalls,
      blocks: updatedBlocks,
    };
  });
}
