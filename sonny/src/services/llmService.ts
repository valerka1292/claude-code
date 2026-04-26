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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function parseCompletionChunk(raw: string): CompletionChunk | null {
  try {
    const parsed: unknown = JSON.parse(raw);
    if (!isRecord(parsed)) {
      return null;
    }

    const chunk: CompletionChunk = {};

    if (Array.isArray(parsed.choices)) {
      chunk.choices = parsed.choices as CompletionChunkChoice[];
    }

    if (isRecord(parsed.usage)) {
      const usage = parsed.usage as Record<string, unknown>;
      if (
        typeof usage.completion_tokens === 'number' &&
        typeof usage.prompt_tokens === 'number' &&
        typeof usage.total_tokens === 'number'
      ) {
        chunk.usage = {
          completion_tokens: usage.completion_tokens,
          prompt_tokens: usage.prompt_tokens,
          total_tokens: usage.total_tokens,
        };
      }
    }

    return chunk;
  } catch {
    return null;
  }
}

const SET_DIALOG_NAME_TOOL = {
  type: 'function' as const,
  function: {
    name: 'setDialogName',
    description: 'Set a short human-readable title for this chat (max 6 words).',
    parameters: {
      type: 'object',
      properties: {
        name: {
          type: 'string',
          description: 'A concise conversation title.',
        },
      },
      required: ['name'],
    },
  },
};

export async function streamChatCompletion(
  provider: Provider,
  messages: { role: string; content: string }[],
  callbacks: StreamCallbacks,
  signal?: AbortSignal,
  tools?: unknown[],
): Promise<void> {
  const url = `${provider.baseUrl}/chat/completions`;
  console.log(`[LLM] Starting stream to ${url}, model: ${provider.model}`);

  const body: Record<string, unknown> = {
    model: provider.model,
    messages,
    stream: true,
    stream_options: { include_usage: true },
  };

  if (tools) {
    body.tools = tools;
  }

  let reader: ReadableStreamDefaultReader<Uint8Array> | undefined;
  let doneCalled = false;
  const cleanup = async () => {
    if (!reader) return;
    try {
      console.log('[LLM] Cleaning up reader');
      await reader.cancel();
    } catch (e) {
      console.warn('[LLM] Cleanup error:', e);
    } finally {
      reader = undefined;
    }
  };

  const safeDone = (usage?: CompletionUsage) => {
    if (!doneCalled) {
      console.log('[LLM] safeDone called', usage);
      doneCalled = true;
      callbacks.onDone(usage);
    }
  };

  const accumulatedToolCalls: Record<number, ToolCallDelta> = {};
  let lastUsage: CompletionUsage | undefined;

  const processDataLine = (jsonStr: string): 'done' | 'continue' => {
    if (jsonStr === '[DONE]') {
      safeDone(lastUsage);
      return 'done';
    }

    const parsed = parseCompletionChunk(jsonStr);
    if (!parsed) return 'continue';

    if (parsed.usage) {
      lastUsage = parsed.usage;
    }

    const choice = parsed.choices?.[0];
    if (!choice?.delta) return 'continue';

    const { delta } = choice;

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

    return 'continue';
  };

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${provider.apiKey}`,
      },
      body: JSON.stringify(body),
      signal,
    });

    if (!response.ok) {
      const text = await response.text();
      console.error(`[LLM] HTTP error ${response.status}: ${text}`);
      callbacks.onError(new Error(`HTTP ${response.status}: ${text}`));
      return;
    }

    reader = response.body?.getReader();
    if (!reader) {
      callbacks.onError(new Error('No response body'));
      return;
    }

    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed || !trimmed.startsWith('data:')) continue;

        const jsonStr = trimmed.slice(5).trim();
        if (processDataLine(jsonStr) === 'done') return;
      }
    }

    // Process trailing buffer
    const trailing = buffer.trim();
    if (trailing && trailing.startsWith('data:')) {
      const jsonStr = trailing.slice(5).trim();
      processDataLine(jsonStr);
    }

    safeDone(lastUsage);
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') {
      console.log('[LLM] Stream aborted');
      safeDone();
    } else {
      console.error('[LLM] Stream error:', error);
      callbacks.onError(error);
    }
  } finally {
    await cleanup();
  }
}

export async function generateChatName(
  provider: Provider, 
  firstUserMessage: string,
  signal?: AbortSignal
): Promise<string> {
  console.log('[LLM] Generating chat name...');
  try {
    const response = await fetch(`${provider.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${provider.apiKey}`,
      },
      signal,
      body: JSON.stringify({
        model: provider.model,
        messages: [
          {
            role: 'system',
            content:
              'You name conversations. Always call setDialogName with a short title (max 6 words, no quotes).',
          },
          { role: 'user', content: firstUserMessage },
        ],
        tools: [SET_DIALOG_NAME_TOOL],
        tool_choice: { type: 'function', function: { name: 'setDialogName' } },
        stream: false,
      }),
    });

    if (!response.ok) {
      console.warn(`[LLM] Name generation failed: ${response.status}`);
      return 'Untitled Chat';
    }

    const payload = (await response.json()) as {
      choices?: Array<{
        message?: {
          tool_calls?: Array<{
            function?: {
              arguments?: string;
            };
          }>;
        };
      }>;
    };
    const argumentsText = payload.choices?.[0]?.message?.tool_calls?.[0]?.function?.arguments;
    if (!argumentsText) return 'Untitled Chat';

    const parsed = JSON.parse(argumentsText) as { name?: string };
    const generatedTitle = parsed.name?.trim();
    if (!generatedTitle) return 'Untitled Chat';

    console.log(`[LLM] Generated title: "${generatedTitle}"`);
    return generatedTitle.slice(0, 60);
  } catch (e) {
    console.error('[LLM] generateChatName error:', e);
    return 'Untitled Chat';
  }
}

function mergeAbortSignals(signalA: AbortSignal, signalB: AbortSignal): AbortSignal {
  const controller = new AbortController();
  const abort = () => controller.abort();
  signalA.addEventListener('abort', abort, { once: true });
  signalB.addEventListener('abort', abort, { once: true });
  return controller.signal;
}

export async function testProviderStream(provider: Provider, signal?: AbortSignal): Promise<boolean> {
  const controller = new AbortController();
  const timeoutId = window.setTimeout(() => controller.abort(), 15000);
  const combinedSignal = signal ? mergeAbortSignals(signal, controller.signal) : controller.signal;

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
      signal: combinedSignal,
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
