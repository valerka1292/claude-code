/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export interface Provider {
  id: string;
  name: string;
  baseUrl: string;
  model: string;
  apiKey: string;
  contextWindow: number;
}

export interface ProvidersStore {
  providers: Record<string, Provider>;
  activeProviderId: string | null;
}

export function getActiveProvider(store: ProvidersStore): Provider | null {
  if (!store.activeProviderId) return null;
  return store.providers[store.activeProviderId] ?? null;
}

export function createProvider(
  store: ProvidersStore,
  provider: Omit<Provider, "id">
): ProvidersStore {
  const id = `provider_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`;
  return {
    providers: { ...store.providers, [id]: { id, ...provider } },
    activeProviderId: store.activeProviderId ?? id,
  };
}

export function updateProvider(
  store: ProvidersStore,
  id: string,
  patch: Partial<Omit<Provider, "id">>
): ProvidersStore {
  if (!store.providers[id]) return store;
  return {
    ...store,
    providers: {
      ...store.providers,
      [id]: { ...store.providers[id], ...patch },
    },
  };
}

export function deleteProvider(store: ProvidersStore, id: string): ProvidersStore {
  const { [id]: _removed, ...rest } = store.providers;
  const ids = Object.keys(rest);
  return {
    providers: rest,
    activeProviderId:
      store.activeProviderId === id ? (ids[0] ?? null) : store.activeProviderId,
  };
}

export function setActiveProvider(store: ProvidersStore, id: string): ProvidersStore {
  return { ...store, activeProviderId: id };
}
