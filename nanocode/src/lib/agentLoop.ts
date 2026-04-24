/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Provider } from "./providers";

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
  onUsage: (promptTokens: number, completionTokens: number, reasoningTokens: number) => void;
  onError: (err: Error) => void;
  onDone: () => void;
}

interface ToolCallBuffer {
  id: string;
  name: string;
  arguments: string;
}

export async function runAgentStream(
  provider: Provider,
  messages: ChatMessage[],
  callbacks: StreamCallbacks,
  signal?: AbortSignal
): Promise<void> {
  const history: ChatMessage[] = [...messages];

  let iterationSafety = 0;
  const MAX_ITERATIONS = 10;

  while (iterationSafety < MAX_ITERATIONS) {
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
      if ((err as DOMException)?.name === "AbortError") return;
      callbacks.onError(err as Error);
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

    for (const tc of toolCallsList) {
      callbacks.onToolCallDone(tc.id, tc.arguments);
    }

    if (finishReason === "tool_calls" && toolCallsList.length > 0) {
      for (const tc of toolCallsList) {
        history.push({
          role: "tool",
          tool_call_id: tc.id,
          name: tc.name,
          content: JSON.stringify({ error: "Tool not implemented yet" }),
        });
      }
      continue;
    }

    break;
  }

  callbacks.onDone();
}