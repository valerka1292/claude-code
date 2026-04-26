import { describe, expect, test, mock } from "bun:test";
import { startNewSessionWithGuard } from "./sessionStartGuard";
import type { SessionData } from "../types/session";

const createSession = (messages: SessionData["messages"]): SessionData => ({
  id: "session-1",
  name: "New session",
  createdAt: Date.now(),
  projectPath: "/tmp/project",
  messages,
});

describe("startNewSessionWithGuard", () => {
  test("blocks fast New session click while turn is active and keeps current turn data", async () => {
    const activeSession = createSession([
      { id: "u1", role: "user", content: "run tool", ts: 1 },
      {
        id: "a1",
        role: "assistant",
        content: "Calling tool...",
        tool_calls: [{ id: "tc1", type: "function", function: { name: "ls", arguments: "{}" } }],
        ts: 2,
      },
      { id: "t1", role: "tool", content: "ok", tool_call_id: "tc1", name: "ls", ts: 3 },
    ]);
    const clearActiveSession = mock(() => {});
    const saveSession = mock(async () => {});
    const onSessionSaveError = mock(() => {});

    const result = await startNewSessionWithGuard({
      isTurnActive: true,
      currentSession: activeSession,
      projectKey: "project-key",
      clearActiveSession,
      saveSession,
      onSessionSaveError,
    });

    expect(result).toBe(false);
    expect(clearActiveSession).toHaveBeenCalledTimes(0);
    expect(saveSession).toHaveBeenCalledTimes(0);
    expect(activeSession.messages.at(-1)?.role).toBe("tool");
  });
});
