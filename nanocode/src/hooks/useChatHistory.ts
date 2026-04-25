import { useCallback } from "react";
import type { Mode } from "../types";
import { buildSystemMessages } from "../lib/systemPrompt";
import { storedToChat } from "../lib/converters";
import type { ChatMessage } from "../lib/agentLoop";
import type { StoredMessage } from "../types/session";

interface BuildChatHistoryOptions {
  cwd: string;
  projectName: string;
  mode: Mode;
  sessionMessages: StoredMessage[];
  userInput: string;
}

export function useChatHistory() {
  const buildChatHistory = useCallback(
    ({
      cwd,
      projectName,
      mode,
      sessionMessages,
      userInput,
    }: BuildChatHistoryOptions): ChatMessage[] => [
      ...buildSystemMessages({
        cwd,
        projectName,
        mode,
      }),
      ...sessionMessages.map(storedToChat),
      { role: "user", content: userInput },
    ],
    []
  );

  return { buildChatHistory };
}
