import type { ProvidersData } from '../types';

export {};

declare global {
  interface Window {
    electron?: {
      minimize: () => void;
      maximize: () => void;
      close: () => void;
      platform: string;
      providers?: {
        getAll: () => Promise<ProvidersData>;
        save: (data: ProvidersData) => Promise<ProvidersData>;
      };
    };
  }
}
