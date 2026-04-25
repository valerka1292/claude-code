import type { Mode } from "../types";
import type { ChatMessage } from "./agentLoop";
import { storedToChat } from "./converters";
import { buildSystemMessages } from "./systemPrompt";
import type { StoredMessage } from "../types/session";

interface BuildChatHistoryOptions {
  cwd: string;
  projectName: string;
  mode: Mode;
  sessionMessages: StoredMessage[];
  userInput: string;
}

export function buildChatHistory({
  cwd,
  projectName,
  mode,
  sessionMessages,
  userInput,
}: BuildChatHistoryOptions): ChatMessage[] {
  return [
    ...buildSystemMessages({
      cwd,
      projectName,
      mode,
    }),
    ...sessionMessages.map(storedToChat),
    { role: "user", content: userInput },
  ];
}
