import type { Provider } from '../types';

export interface StreamCallbacks {
  onContent: (text: string) => void;
  onThinking: (text: string) => void;
  onToolCall: (toolCall: ToolCallDelta) => void;
  onDone: (usage?: CompletionUsage) => void;
  onError: (error: unknown) => void;
}

export interface ToolCallDelta {
  index: number;
  id?: string;
  function?: {
    name?: string;
    arguments?: string;
  };
}

export interface CompletionUsage {
  completion_tokens: number;
  prompt_tokens: number;
  total_tokens: number;
}

interface CompletionChunkChoice {
  delta?: {
    content?: string;
    reasoning_content?: string;
    tool_calls?: Array<{
      index: number;
      id?: string;
      function?: {
        name?: string;
        arguments?: string;
      };
    }>;
  };
  finish_reason?: string;
}

interface CompletionChunk {
  choices?: CompletionChunkChoice[];
  usage?: CompletionUsage;
}

export function streamChatCompletion(
  provider: Provider,
  messages: { role: string; content: string }[],
  callbacks: StreamCallbacks,
  signal?: AbortSignal,
  tools?: unknown[],
) {
  const url = `${provider.baseUrl}/chat/completions`;

  const body: Record<string, unknown> = {
    model: provider.model,
    messages,
    stream: true,
    stream_options: { include_usage: true },
  };

  if (tools) {
    body.tools = tools;
  }

  fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${provider.apiKey}`,
    },
    body: JSON.stringify(body),
    signal,
  })
    .then(async (response) => {
      if (!response.ok) {
        const text = await response.text();
        callbacks.onError(new Error(`HTTP ${response.status}: ${text}`));
        return;
      }

      const reader = response.body?.getReader();
      if (!reader) {
        callbacks.onError(new Error('No response body'));
        return;
      }

      const decoder = new TextDecoder();
      let buffer = '';
      const accumulatedToolCalls: Record<number, ToolCallDelta> = {};

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            break;
          }

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() ?? '';

          for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed || !trimmed.startsWith('data:')) {
              continue;
            }

            const jsonStr = trimmed.slice(5).trim();
            if (jsonStr === '[DONE]') {
              callbacks.onDone();
              return;
            }

            try {
              const parsed = JSON.parse(jsonStr) as CompletionChunk;
              const choice = parsed.choices?.[0];
              if (!choice) {
                continue;
              }

              if (parsed.usage) {
                callbacks.onDone(parsed.usage);
                return;
              }

              const delta = choice.delta;
              if (!delta) {
                continue;
              }

              if (delta.reasoning_content) {
                callbacks.onThinking(delta.reasoning_content);
              }

              if (delta.content) {
                callbacks.onContent(delta.content);
              }

              if (delta.tool_calls) {
                for (const tc of delta.tool_calls) {
                  const index = tc.index;
                  if (!accumulatedToolCalls[index]) {
                    accumulatedToolCalls[index] = {
                      index,
                      id: tc.id,
                      function: {
                        name: tc.function?.name,
                        arguments: '',
                      },
                    };
                  }

                  if (tc.function?.name) {
                    accumulatedToolCalls[index].function = {
                      ...accumulatedToolCalls[index].function,
                      name: tc.function.name,
                    };
                  }

                  if (tc.function?.arguments) {
                    const currentArgs = accumulatedToolCalls[index].function?.arguments ?? '';
                    accumulatedToolCalls[index].function = {
                      ...accumulatedToolCalls[index].function,
                      arguments: currentArgs + tc.function.arguments,
                    };
                  }

                  callbacks.onToolCall({ ...accumulatedToolCalls[index] });
                }
              }

              if (choice.finish_reason === 'stop') {
                // Wait for usage or [DONE]
              }
            } catch {
              // Ignore malformed/partial chunks.
            }
          }
        }
      } catch (error) {
        if (error instanceof Error && error.name === 'AbortError') {
          return;
        }
        callbacks.onError(error);
      }
    })
    .catch((error) => callbacks.onError(error));
}

export async function testProviderStream(provider: Provider): Promise<boolean> {
  const controller = new AbortController();
  const timeoutId = window.setTimeout(() => controller.abort(), 15000);

  try {
    const response = await fetch(`${provider.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${provider.apiKey}`,
      },
      body: JSON.stringify({
        model: provider.model,
        messages: [{ role: 'user', content: 'Send OK and nothing more.' }],
        stream: true,
        max_tokens: 8,
      }),
      signal: controller.signal,
    });

    if (!response.ok) {
      return false;
    }

    const reader = response.body?.getReader();
    if (!reader) {
      return false;
    }

    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';

      for (const line of lines) {
        if (!line.startsWith('data: ')) {
          continue;
        }

        const payload = line.slice(6).trim();
        if (payload === '[DONE]') {
          return true;
        }

        try {
          const json = JSON.parse(payload) as {
            choices?: Array<{ delta?: { content?: string }; finish_reason?: string }>;
          };
          const choice = json.choices?.[0];
          if (choice?.delta?.content || choice?.finish_reason === 'stop') {
            return true;
          }
        } catch {
          // Ignore malformed chunks.
        }
      }
    }

    return false;
  } catch {
    return false;
  } finally {
    clearTimeout(timeoutId);
  }
}
