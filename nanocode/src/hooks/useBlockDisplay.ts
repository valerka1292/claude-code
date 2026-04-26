import { useMemo } from "react";
import type { ContentBlock, ToolCallDisplay } from "../types/message";

export interface DisplayBlockReasoning {
  type: "reasoning";
  key: string;
  content: string;
  streaming?: boolean;
}

export interface DisplayBlockText {
  type: "text";
  key: string;
  content: string;
  streaming?: boolean;
}

export interface DisplayBlockToolCall {
  type: "tool_call";
  key: string;
  call: ToolCallDisplay;
  statusOverride?: ToolCallDisplay["status"];
  resultOverride?: string;
}

export type DisplayBlock =
  | DisplayBlockReasoning
  | DisplayBlockText
  | DisplayBlockToolCall;

export function useBlockDisplay(blocks: ContentBlock[]): DisplayBlock[] {
  return useMemo(() => {
    const toolResultMap = new Map<
      string,
      { status: ToolCallDisplay["status"]; result?: string }
    >();

    for (const block of blocks) {
      if (block.type === "tool_result") {
        toolResultMap.set(block.callId, {
          status: block.status,
          result: block.result,
        });
      }
    }

    return blocks.flatMap((block, index): DisplayBlock[] => {
      if (block.type === "reasoning") {
        return [
          {
            type: "reasoning",
            key: `reasoning-${index}`,
            content: block.content,
            streaming: block.streaming,
          },
        ];
      }

      if (block.type === "text") {
        return [
          {
            type: "text",
            key: `text-${index}`,
            content: block.content,
            streaming: block.streaming,
          },
        ];
      }

      if (block.type === "tool_call") {
        const resultInfo = toolResultMap.get(block.call.id);
        return [
          {
            type: "tool_call",
            key: `tool-call-${block.call.id}`,
            call: block.call,
            statusOverride: resultInfo?.status,
            resultOverride: resultInfo?.result,
          },
        ];
      }

      return [];
    });
  }, [blocks]);
}
