export type AgentMode = 'Chat' | 'Autonomy' | 'Improve' | 'Dream';

export interface ToolCall {
  index: number;
  id?: string;
  function?: {
    name?: string;
    arguments?: string;
  };
}

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  thinking?: string;
  toolCalls?: ToolCall[];
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
