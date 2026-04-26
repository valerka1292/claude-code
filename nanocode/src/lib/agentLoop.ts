/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Provider } from "./providers";
import { getTool, getToolDefinitions } from "./tools";
import { formatToolError } from "./tools/utils/format";

export interface ChatMessage {
  role: "system" | "user" | "assistant" | "tool";
  content: string | null;
  tool_calls?: ToolCall[];
  tool_call_id?: string;
  name?: string;
}

export interface ToolCall {
  id: string;
  type: "function";
  function: {
    name: string;
    arguments: string;
  };
}

export interface StreamCallbacks {
  onReasoningChunk: (chunk: string) => void;
  onContentChunk: (chunk: string) => void;
  onToolCallStart: (id: string, name: string) => void;
  onToolCallDone: (id: string, args: string) => void;
  onToolExecutionStart: (id: string, name: string, args: string) => void;
  onToolExecutionDone: (id: string, result: string) => void;
  onToolExecutionError: (id: string, error: string) => void;
  onAssistantMessageWithTools: (content: string | null, toolCalls: ToolCall[]) => void;
  onUsage: (promptTokens: number, completionTokens: number, reasoningTokens: number) => void;
  onError: (err: Error) => void;
  onDone: () => void;
}

interface ToolCallBuffer {
  id: string;
  name: string;
  arguments: string;
}

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === "string") {
    return error;
  }

  try {
    return JSON.stringify(error);
  } catch {
    return String(error);
  }
}

export async function runAgentStream(
  provider: Provider,
  messages: ChatMessage[],
  callbacks: StreamCallbacks,
  cwd: string,
  signal?: AbortSignal
): Promise<void> {
  const history: ChatMessage[] = [...messages];

  let iterationSafety = 0;
  const MAX_ITERATIONS = 10;

  while (iterationSafety < MAX_ITERATIONS) {
    if (signal?.aborted) return;
    iterationSafety++;

    let fullContent = "";
    let fullReasoning = "";
    const toolCallsBuffer: Record<number, ToolCallBuffer> = {};

    let finishReason: string | null = null;

    try {
      const response = await fetch(`${provider.baseUrl}/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${provider.apiKey}`,
        },
        signal,
        body: JSON.stringify({
          model: provider.model,
          messages: history,
          stream: true,
          stream_options: { include_usage: true },
          tools: getToolDefinitions(),
        }),
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(`API error ${response.status}: ${text}`);
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error("No response body");

      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed || !trimmed.startsWith("data:")) continue;
          const data = trimmed.slice(5).trim();
          if (data === "[DONE]") continue;

          let chunk: unknown;
          try {
            chunk = JSON.parse(data);
          } catch {
            continue;
          }

          const c = chunk as Record<string, unknown>;

          if (c.usage) {
            const u = c.usage as Record<string, unknown>;
            const reasoningTokens =
              (u.completion_tokens_details as Record<string, unknown>)?.reasoning_tokens ?? 0;
            callbacks.onUsage(
              (u.prompt_tokens as number) ?? 0,
              (u.completion_tokens as number) ?? 0,
              reasoningTokens as number
            );
          }

          const choice = (c.choices as unknown[])?.[0] as Record<string, unknown> | undefined;
          if (!choice) continue;

          const delta = choice.delta as Record<string, unknown> ?? {};

          if (delta.reasoning_content) {
            fullReasoning += delta.reasoning_content as string;
            callbacks.onReasoningChunk(delta.reasoning_content as string);
          }

          if (delta.content) {
            fullContent += delta.content as string;
            callbacks.onContentChunk(delta.content as string);
          }

          if (delta.tool_calls) {
            for (const tc of delta.tool_calls as unknown[] as { index?: number; id?: string; function?: { name?: string; arguments?: string } }[]) {
              const idx: number = tc.index ?? 0;
              if (!toolCallsBuffer[idx]) {
                toolCallsBuffer[idx] = { id: "", name: "", arguments: "" };
              }
              if (tc.id) {
                toolCallsBuffer[idx].id = tc.id;
                callbacks.onToolCallStart(tc.id, tc.function?.name ?? "");
              }
              if (tc.function?.name) {
                toolCallsBuffer[idx].name = tc.function.name;
              }
              if (tc.function?.arguments) {
                toolCallsBuffer[idx].arguments += tc.function.arguments;
              }
            }
          }

          if (choice.finish_reason) {
            finishReason = choice.finish_reason as string;
          }
        }
      }
    } catch (err) {
      if ((err as DOMException)?.name === "AbortError") {
        callbacks.onError(new Error("Generation cancelled"));
        return;
      }
      callbacks.onError(err instanceof Error ? err : new Error(getErrorMessage(err)));
      return;
    }

    const assistantMsg: ChatMessage = {
      role: "assistant",
      content: fullContent || null,
    };
    const toolCallsList = Object.values(toolCallsBuffer);
    if (toolCallsList.length > 0) {
      assistantMsg.tool_calls = toolCallsList.map((tc) => ({
        id: tc.id,
        type: "function",
        function: { name: tc.name, arguments: tc.arguments },
      }));
    }
    history.push(assistantMsg);

    if (toolCallsList.length > 0) {
      callbacks.onAssistantMessageWithTools(
        fullContent || null,
        toolCallsList.map((tc) => ({
          id: tc.id,
          type: "function",
          function: { name: tc.name, arguments: tc.arguments },
        }))
      );
    }

    for (const tc of toolCallsList) {
      callbacks.onToolCallDone(tc.id, tc.arguments);
    }

    if (finishReason === "tool_calls" && toolCallsList.length > 0) {
      for (const tc of toolCallsList) {
        try {
          const tool = getTool(tc.name);
          if (!tool) {
            throw new Error(`Unknown tool: ${tc.name}`);
          }

          const args = JSON.parse(tc.arguments) as Record<string, unknown>;
          callbacks.onToolExecutionStart(tc.id, tc.name, tc.arguments);
          const result = await tool.execute(args, {
            cwd,
            signal,
          });
          callbacks.onToolExecutionDone(tc.id, result);

          history.push({
            role: "tool",
            tool_call_id: tc.id,
            name: tc.name,
            content: result,
          });
        } catch (error) {
          const errorMsg = getErrorMessage(error);
          callbacks.onToolExecutionError(tc.id, errorMsg);
          history.push({
            role: "tool",
            tool_call_id: tc.id,
            name: tc.name,
            content: formatToolError(errorMsg),
          });

          if (signal?.aborted) {
            return;
          }
        }
      }

      continue;
    }

    break;
  }

  if (!signal?.aborted) {
    callbacks.onDone();
  }
}
