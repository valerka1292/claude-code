import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { finalizeLiveTurnMessages } from "./useMessageStream";
import { turnsToMessages } from "../lib/turnsToMessages";
import type { Message } from "../types/message";
import type { StoredMessage } from "../types/session";

describe("message stream state transitions", () => {
  it("stream done + immediate session restore keeps finalized blocks and avoids live rollback", () => {
    const archiveBefore: Message[] = [
      { id: "a-user", role: "user", content: "prev" },
      { id: "a-assistant", role: "assistant", content: "", blocks: [{ type: "text", content: "done" }] },
    ];

    const liveTurn: Message[] = [
      { id: "l-user", role: "user", content: "new" },
      {
        id: "l-assistant",
        role: "assistant",
        content: "",
        isStreaming: true,
        isReasoningStreaming: true,
        blocks: [
          { type: "reasoning", content: "thinking", streaming: true },
          { type: "text", content: "answer", streaming: true },
        ],
      },
    ];

    const finalizedLiveTurn = finalizeLiveTurnMessages(liveTurn);
    const archiveAfterDone = [...archiveBefore, ...finalizedLiveTurn];

    const persistedSessionMessages: StoredMessage[] = [
      { id: "a-user", role: "user", content: "prev", ts: 1 },
      { id: "a-assistant", role: "assistant", content: "done", ts: 2 },
      { id: "l-user", role: "user", content: "new", ts: 3 },
      { id: "l-assistant", role: "assistant", content: "answer", reasoning: "thinking", ts: 4 },
    ];

    // useSessionRestore now updates archive slice only.
    const archiveFromRestore = turnsToMessages(persistedSessionMessages);
    const visibleAfterRestore = [...archiveFromRestore];

    assert.equal(archiveAfterDone[3].isStreaming, false);
    assert.equal(
      archiveAfterDone[3].blocks?.every(
        (b) => b.type === "tool_call" || b.type === "tool_result" || b.streaming === false
      ),
      true
    );
    assert.deepEqual(visibleAfterRestore.at(-1)?.blocks?.[0], {
      type: "reasoning",
      content: "thinking",
    });
    assert.deepEqual(visibleAfterRestore.at(-1)?.blocks?.[1], {
      type: "text",
      content: "answer",
    });
  });
});
