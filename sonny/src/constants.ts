import { AgentMode, ChatSession, Provider } from './types';

export const AGENT_MODES: { id: AgentMode; label: string; description: string }[] = [
  { id: 'Chat', label: 'Chat', description: 'Standard conversation with the agent.' },
  { id: 'Autonomy', label: 'Autonomy', description: 'Agent executes tasks independently in loops.' },
  { id: 'Improve', label: 'Improve', description: 'Agent analyzes and enhances its own code/logic.' },
  { id: 'Dream', label: 'Dream', description: 'Exploratory mode for creative ideation.' },
];

export const MOCK_CHATS: ChatSession[] = [
  { id: '1', title: 'Implementing Electron Titlebar', updatedAt: new Date() },
  { id: '2', title: 'Rust Backend Architecture', updatedAt: new Date(Date.now() - 86400000) },
  { id: '3', title: 'Agent Loop Synchronization', updatedAt: new Date(Date.now() - 172800000) },
];

export const MOCK_PROVIDERS: Provider[] = [
  {
    id: 'mock-provider',
    visualName: 'Default Provider',
    baseUrl: 'http://localhost:11434/v1',
    apiKey: 'mock-key',
    model: 'gpt-4o-mini',
    contextWindowSize: 32768,
  },
];
