/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Mode } from "../types";
import type { ChatMessage } from "./agentLoop";

const STATIC_SYSTEM_PROMPT = [
  "# Identity",
  "You are **nanocode** — an autonomous coding agent.",
  "",
  "# Tool Protocol",
  "- Solve tasks through available tool calls, not by guessing outcomes.",
  "- Treat tool output as the single source of truth.",
  "- Any read, write, search, or execution action must happen via tools.",
  "",
  "# Agentic Loop",
  "- Continue until user request is fully resolved or a blocker is hit.",
  "- Evaluate, choose best tool, execute.",
  "- Keep CLI text concise. Return short completion report.",
  "- Use absolute paths only in tool calls.",
  "- In user-facing text (reports/commands/headings), use paths relative to CWD",
  "  and never print absolute paths unless the user explicitly asks.",
].join("\n");

export function buildStaticSystemPrompt(): string {
  return STATIC_SYSTEM_PROMPT;
}

export interface DynamicPromptOptions {
  cwd: string;
  projectName: string;
  os?: string;
  shell?: string;
  mode?: Mode;
}

export function buildDynamicSystemPrompt(opts: DynamicPromptOptions): string {
  const {
    cwd,
    projectName,
    os: osName =
      typeof navigator !== "undefined" ? navigator.platform : "unknown",
    shell = "bash",
    mode = "Ask",
  } = opts;

  const modeDesc =
    mode === "Code"
      ? "Current mode: code (READ/WRITE).\nRead and write operations are permitted.\nProceed with file modifications as needed."
      : "Current mode: ask (READ-ONLY).\nWrite operations are not permitted.\nSwitch to code mode for write access.";

  const sections = [
    "# Environment",
    `- CWD (tool calls only): \`${cwd}\``,
    `- Project name (user-facing): \`${projectName}\``,
    `- OS: \`${osName}\``,
    `- Shell: \`${shell}\``,
    "",
    "# Permissions",
    modeDesc,
  ];

  return sections.join("\n");
}

export function buildSystemMessages(opts: DynamicPromptOptions): ChatMessage[] {
  return [
    { role: "system", content: buildStaticSystemPrompt() },
    { role: "system", content: buildDynamicSystemPrompt(opts) },
  ];
}
