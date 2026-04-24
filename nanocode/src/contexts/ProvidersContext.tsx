/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import {
  type ProvidersStore,
  type Provider,
  loadProviders,
  saveProviders,
  createProvider,
  updateProvider,
  deleteProvider,
  setActiveProvider,
  getActiveProvider,
} from "../lib/providers";

interface ProvidersContextValue {
  store: ProvidersStore;
  activeProvider: Provider | null;
  isLoading: boolean;
  addProvider: (p: Omit<Provider, "id">) => void;
  editProvider: (id: string, patch: Partial<Omit<Provider, "id">>) => void;
  removeProvider: (id: string) => void;
  switchProvider: (id: string) => void;
}

const ProvidersContext = createContext<ProvidersContextValue | null>(null);

export function ProvidersProvider({ children }: { children: ReactNode }) {
  const [store, setStore] = useState<ProvidersStore>({
    providers: {},
    activeProviderId: null,
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadProviders().then((s) => {
      setStore(s);
      setIsLoading(false);
    });
  }, []);

  const update = useCallback(
    (updater: (s: ProvidersStore) => ProvidersStore) => {
      setStore((prev) => {
        const next = updater(prev);
        saveProviders(next);
        return next;
      });
    },
    []
  );

  const addProvider = (p: Omit<Provider, "id">) => {
    update((s) => createProvider(s, p));
  };

  const editProvider = (id: string, patch: Partial<Omit<Provider, "id">>) => {
    update((s) => updateProvider(s, id, patch));
  };

  const removeProvider = (id: string) => {
    update((s) => deleteProvider(s, id));
  };

  const switchProvider = (id: string) => {
    update((s) => setActiveProvider(s, id));
  };

  return (
    <ProvidersContext.Provider
      value={{
        store,
        activeProvider: getActiveProvider(store),
        isLoading,
        addProvider,
        editProvider,
        removeProvider,
        switchProvider,
      }}
    >
      {children}
    </ProvidersContext.Provider>
  );
}

export function useProviders() {
  const ctx = useContext(ProvidersContext);
  if (!ctx) throw new Error("useProviders must be used inside ProvidersProvider");
  return ctx;
}