export type AgentMode = 'Chat' | 'Autonomy' | 'Improve' | 'Dream';

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

export interface ChatSession {
  id: string;
  title: string;
  updatedAt: Date;
}

export interface Provider {
  id: string;
  visualName: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  contextWindowSize: number;
}

export interface ProvidersData {
  activeProviderId: string | null;
  providers: Record<string, Provider>;
}
