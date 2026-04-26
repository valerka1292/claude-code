import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';
import type { Provider, ProvidersData } from '../types';

interface ProvidersContextValue {
  data: ProvidersData;
  isLoaded: boolean;
  activeProvider: Provider | null;
  refresh: () => Promise<void>;
  addProvider: (provider: Provider) => Promise<void>;
  updateProvider: (providerId: string, patch: Partial<Provider>) => Promise<void>;
  deleteProvider: (providerId: string) => Promise<void>;
  setActiveProvider: (providerId: string) => Promise<void>;
}

const defaultProvidersData: ProvidersData = {
  activeProviderId: null,
  providers: {},
};

const ProvidersContext = createContext<ProvidersContextValue | null>(null);

export function ProvidersProvider({ children }: { children: React.ReactNode }) {
  const [data, setData] = useState<ProvidersData>(defaultProvidersData);
  const [isLoaded, setIsLoaded] = useState(false);

  const refresh = useCallback(async () => {
    const bridge = window.electron?.providers;
    if (!bridge) {
      setIsLoaded(true);
      return;
    }

    const incoming = await bridge.getAll();
    setData(incoming);
    setIsLoaded(true);
  }, []);

  React.useEffect(() => {
    refresh();
  }, [refresh]);

  const saveAndReplace = useCallback(async (nextData: ProvidersData) => {
    const bridge = window.electron?.providers;
    if (!bridge) {
      setData(nextData);
      return;
    }

    const saved = await bridge.save(nextData);
    setData(saved);
  }, []);

  const addProvider = useCallback(async (provider: Provider) => {
    const nextData: ProvidersData = {
      activeProviderId: data.activeProviderId ?? provider.id,
      providers: {
        ...data.providers,
        [provider.id]: provider,
      },
    };
    await saveAndReplace(nextData);
  }, [data, saveAndReplace]);

  const updateProvider = useCallback(async (providerId: string, patch: Partial<Provider>) => {
    const current = data.providers[providerId];
    if (!current) {
      return;
    }

    const nextData: ProvidersData = {
      ...data,
      providers: {
        ...data.providers,
        [providerId]: {
          ...current,
          ...patch,
        },
      },
    };
    await saveAndReplace(nextData);
  }, [data, saveAndReplace]);

  const deleteProvider = useCallback(async (providerId: string) => {
    const { [providerId]: _removed, ...restProviders } = data.providers;
    const fallbackId = Object.keys(restProviders)[0] ?? null;
    const nextData: ProvidersData = {
      activeProviderId: data.activeProviderId === providerId ? fallbackId : data.activeProviderId,
      providers: restProviders,
    };
    await saveAndReplace(nextData);
  }, [data, saveAndReplace]);

  const setActiveProvider = useCallback(async (providerId: string) => {
    if (!data.providers[providerId]) {
      return;
    }

    const nextData: ProvidersData = {
      ...data,
      activeProviderId: providerId,
    };
    await saveAndReplace(nextData);
  }, [data, saveAndReplace]);

  const activeProvider = useMemo(() => {
    if (!data.activeProviderId) {
      return null;
    }
    return data.providers[data.activeProviderId] ?? null;
  }, [data]);

  const value = useMemo<ProvidersContextValue>(() => ({
    data,
    isLoaded,
    activeProvider,
    refresh,
    addProvider,
    updateProvider,
    deleteProvider,
    setActiveProvider,
  }), [activeProvider, addProvider, data, deleteProvider, isLoaded, refresh, setActiveProvider, updateProvider]);

  return <ProvidersContext.Provider value={value}>{children}</ProvidersContext.Provider>;
}

export function useProvidersContext() {
  const context = useContext(ProvidersContext);
  if (!context) {
    throw new Error('useProvidersContext must be used within ProvidersProvider');
  }
  return context;
}
