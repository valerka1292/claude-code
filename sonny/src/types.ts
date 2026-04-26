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
  name: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  contextSize: number;
}
