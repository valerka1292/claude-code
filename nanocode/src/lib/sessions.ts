/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import type { SessionData, SessionMeta } from "../types/session";
import type { Provider } from "./providers";

export type { SessionData, SessionMeta, StoredMessage } from "../types/session";

const DEFAULT_SESSION_NAME = "New session";

function generateId(): string {
  return crypto.randomUUID();
}

export async function listSessions(
  projectKey: string
): Promise<SessionMeta[]> {
  return (await window.electronAPI?.listSessions(projectKey)) ?? [];
}

export async function loadSession(
  projectKey: string,
  id: string
): Promise<SessionData | null> {
  return (await window.electronAPI?.loadSession(projectKey, id)) ?? null;
}

export async function saveSession(
  projectKey: string,
  session: SessionData
): Promise<void> {
  await window.electronAPI?.saveSession(projectKey, session);
}

export async function deleteSession(
  projectKey: string,
  id: string
): Promise<void> {
  await window.electronAPI?.deleteSession(projectKey, id);
}

export function createNewSession(projectPath: string): SessionData {
  return {
    id: generateId(),
    projectPath,
    name: DEFAULT_SESSION_NAME,
    createdAt: Date.now(),
    messages: [],
  };
}

const SET_DIALOG_NAME_TOOL = {
  type: "function" as const,
  function: {
    name: "setDialogName",
    description:
      "Set a short human-readable name for this conversation (max 6 words)",
    parameters: {
      type: "object",
      properties: {
        name: { type: "string", description: "Short conversation title" },
      },
      required: ["name"],
    },
  },
};

export async function generateSessionName(
  provider: Provider,
  firstUserMessage: string
): Promise<string> {
  try {
    const response = await fetch(`${provider.baseUrl}/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${provider.apiKey}`,
      },
      body: JSON.stringify({
        model: provider.model,
        messages: [
          {
            role: "system",
            content:
              "You name conversations. Call setDialogName with a short title (max 6 words, no quotes).",
          },
          { role: "user", content: firstUserMessage },
        ],
        tools: [SET_DIALOG_NAME_TOOL],
        tool_choice: { type: "function", function: { name: "setDialogName" } },
        stream: false,
      }),
    });

    if (!response.ok) return DEFAULT_SESSION_NAME;

    const data = await response.json();
    const toolCall = data.choices?.[0]?.message?.tool_calls?.[0];
    if (!toolCall) return DEFAULT_SESSION_NAME;

    const args = JSON.parse(toolCall.function.arguments ?? "{}");
    return (args.name as string)?.slice(0, 60) || DEFAULT_SESSION_NAME;
  } catch {
    return DEFAULT_SESSION_NAME;
  }
}
