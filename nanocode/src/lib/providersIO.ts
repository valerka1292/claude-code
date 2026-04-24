import type { Provider, ProvidersStore } from "./providers";

export async function loadProviders(): Promise<ProvidersStore> {
  try {
    if (window.electronAPI?.loadSettings) {
      const raw = await window.electronAPI.loadSettings();
      return {
        providers: (raw.providers as Record<string, Provider>) ?? {},
        activeProviderId: raw.activeProviderId ?? null,
      };
    }
  } catch {
    // fallback
  }
  return { providers: {}, activeProviderId: null };
}

export async function saveProviders(store: ProvidersStore): Promise<void> {
  try {
    await window.electronAPI?.saveSettings?.({
      providers: store.providers,
      activeProviderId: store.activeProviderId,
    });
  } catch (err) {
    console.error("[providers] saveProviders error:", err);
  }
}
