import type { Provider } from '../types';

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
