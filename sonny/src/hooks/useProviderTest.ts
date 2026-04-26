import React from 'react';
import { testProviderStream } from '../services/llmService';
import type { Provider } from '../types';

interface ProviderTestResult {
  ok: boolean;
  message: string;
}

export function useProviderTest() {
  const [testing, setTesting] = React.useState(false);
  const [result, setResult] = React.useState<ProviderTestResult | null>(null);

  const test = React.useCallback(async (provider: Provider) => {
    setTesting(true);
    setResult(null);

    const ok = await testProviderStream(provider);
    setResult({
      ok,
      message: ok ? 'Connection successful' : 'Connection failed',
    });
    setTesting(false);
  }, []);

  const clear = React.useCallback(() => {
    setResult(null);
  }, []);

  return { test, testing, result, clear };
}
